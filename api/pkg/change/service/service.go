package service

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	git "github.com/libgit2/git2go/v33"
	"github.com/pkg/errors"
	"go.uber.org/zap"

	"getsturdy.com/api/pkg/change"
	db_change "getsturdy.com/api/pkg/change/db"
	"getsturdy.com/api/pkg/change/message"
	"getsturdy.com/api/pkg/unidiff"
	"getsturdy.com/api/pkg/workspaces"
	"getsturdy.com/api/vcs"
	"getsturdy.com/api/vcs/executor"

	"github.com/google/uuid"
)

type Service struct {
	changeRepo       db_change.Repository
	logger           *zap.Logger
	executorProvider executor.Provider
}

func New(
	changeRepo db_change.Repository,
	logger *zap.Logger,
	executorProvider executor.Provider,
) *Service {
	return &Service{
		changeRepo:       changeRepo,
		logger:           logger.Named("changeService"),
		executorProvider: executorProvider,
	}
}

func (svc *Service) ListChanges(ctx context.Context, ids ...change.ID) ([]*change.Change, error) {
	return svc.changeRepo.ListByIDs(ctx, ids...)
}

func (svc *Service) GetChangeByID(ctx context.Context, id change.ID) (*change.Change, error) {
	ch, err := svc.changeRepo.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	return ch, nil
}

func (svc *Service) GetByCommitAndCodebase(ctx context.Context, commitID, codebaseID string) (*change.Change, error) {
	ch, err := svc.changeRepo.GetByCommitID(ctx, commitID, codebaseID)
	switch {
	case errors.Is(err, sql.ErrNoRows):
		// import
		return svc.importCommitToChange(ctx, codebaseID, commitID)
	case err != nil:
		return nil, err
	default:
		return ch, nil
	}
}

func (svc *Service) CreateOnTop(ctx context.Context, ws *workspaces.Workspace, commitID string) (*change.Change, error) {
	headChange, err := svc.head(ctx, ws.CodebaseID)
	switch {
	case errors.Is(err, ErrNotFound):
		return svc.CreateWithChangeAsParent(ctx, ws, commitID, nil)
	case err != nil:
		return nil, fmt.Errorf("could not get head change: %w", err)
	default:
		return svc.CreateWithChangeAsParent(ctx, ws, commitID, &headChange.ID)
	}
}

func (svc *Service) CreateWithCommitAsParent(ctx context.Context, ws *workspaces.Workspace, commitID, parentCommitID string) (*change.Change, error) {
	var parentChangeID *change.ID

	parent, err := svc.getChangeFromCommit(ctx, ws.CodebaseID, parentCommitID)
	switch {
	case err == nil:
		parentChangeID = &parent.ID
	case errors.Is(err, ErrNotFound):
	// nothing
	default:
		return nil, fmt.Errorf("failed to get change from parent commit: %w", err)
	}

	return svc.CreateWithChangeAsParent(ctx, ws, commitID, parentChangeID)
}

func (svc *Service) CreateWithChangeAsParent(ctx context.Context, ws *workspaces.Workspace, commitID string, parentChangeID *change.ID) (*change.Change, error) {
	changeID := change.ID(uuid.NewString())
	t := time.Now()

	cleanCommitMessage := message.CommitMessage(ws.DraftDescription)
	cleanCommitMessageTitle := strings.Split(cleanCommitMessage, "\n")[0]

	changeChange := change.Change{
		ID:                 changeID,
		CodebaseID:         ws.CodebaseID,
		Title:              &cleanCommitMessageTitle,
		UpdatedDescription: ws.DraftDescription,
		UserID:             &ws.UserID,
		CreatedAt:          &t,
		CommitID:           &commitID,
		ParentChangeID:     parentChangeID,
	}

	if err := svc.changeRepo.Insert(ctx, changeChange); err != nil {
		return nil, fmt.Errorf("failed to insert change: %w", err)
	}

	return &changeChange, nil
}

func (svc *Service) head(ctx context.Context, codebaseID string) (*change.Change, error) {
	// To find the root commit, peek into git
	var headCommitID string

	getHeadCommit := func(repo vcs.RepoGitReader) error {
		headCommit, err := repo.HeadCommit()
		if err != nil {
			return fmt.Errorf("could not find head commit: %w", err)
		}
		headCommitID = headCommit.Id().String()
		return nil
	}

	err := svc.executorProvider.New().GitRead(getHeadCommit).ExecTrunk(codebaseID, "changeServiceChangelog")
	switch {
	case errors.Is(err, vcs.ErrNotFound):
		return nil, ErrNotFound
	case err != nil:
		return nil, fmt.Errorf("could not get head commit: %w", err)
	default:
		return svc.getChangeFromCommit(ctx, codebaseID, headCommitID)
	}
}

func (svc *Service) Changelog(ctx context.Context, codebaseID string, limit int) ([]*change.Change, error) {
	headChange, err := svc.head(ctx, codebaseID)
	switch {
	case errors.Is(err, ErrNotFound):
		return nil, nil
	case err != nil:
		return nil, fmt.Errorf("could not get head change: %w", err)
	}

	var res []*change.Change
	res = append(res, headChange)
	nextChange := headChange

findChanges:
	for {
		if len(res) >= limit {
			break
		}

		next, err := svc.parentChange(ctx, nextChange)
		switch {
		case errors.Is(err, ErrNotFound):
			break findChanges
		case err != nil:
			return nil, err
		case err == nil:
			res = append(res, next)
			nextChange = next
		}
	}

	return res, nil
}

var ErrNotFound = errors.New("not found")

func (svc *Service) parentChange(ctx context.Context, ch *change.Change) (*change.Change, error) {
	if ch.ParentChangeID != nil {
		// get from db
		next, err := svc.changeRepo.Get(ctx, *ch.ParentChangeID)
		switch {
		case err == nil:
			return next, nil
		case errors.Is(err, sql.ErrNoRows):
			// parent not found in db? re-create and import from git
		case err != nil:
			return nil, fmt.Errorf("could not get parent change from db id=%s: %w", *ch.ParentChangeID, err)
		}
	}

	// get parents from git
	var parents []string
	getCurrentFromGit := func(repo vcs.RepoGitReader) error {
		details, err := repo.GetCommitDetails(*ch.CommitID)
		if err != nil {
			return fmt.Errorf("could not get commit details from repo: %w", err)
		}
		parents = details.Parents
		return nil
	}
	if err := svc.executorProvider.New().GitRead(getCurrentFromGit).ExecTrunk(ch.CodebaseID, "changeService.parentChange"); err != nil {
		return nil, fmt.Errorf("could not get from git: %w", err)
	}

	// this commit is a root commit
	if len(parents) == 0 {
		return nil, ErrNotFound
	}

	// the first parent (usually) refers to the state of the branch that that the branch was merged _into_, prior to the merge.
	parent, err := svc.importCommitToChange(ctx, ch.CodebaseID, parents[0])
	if err != nil {
		return nil, fmt.Errorf("failed to import parent: %w", err)
	}

	// update the "current" commit and mark the new commit as it's parent
	ch.ParentChangeID = &parent.ID
	err = svc.changeRepo.Update(ctx, *ch)
	if err != nil {
		return nil, fmt.Errorf("failed to update change parent: %w", err)
	}

	return parent, nil
}

func (svc *Service) getChangeFromCommit(ctx context.Context, codebaseID, commitID string) (*change.Change, error) {
	ch, err := svc.changeRepo.GetByCommitID(ctx, commitID, codebaseID)
	switch {
	case err == nil:
		return ch, nil
	case errors.Is(err, sql.ErrNoRows):
		return svc.importCommitToChange(ctx, codebaseID, commitID)
	default:
		return nil, fmt.Errorf("failed to get change from db: %w", err)
	}
}

func (svc *Service) importCommitToChange(ctx context.Context, codebaseID, commitID string) (*change.Change, error) {
	// if the change exists in the db, use it!
	{
		fromDb, err := svc.changeRepo.GetByCommitID(ctx, commitID, codebaseID)
		switch {
		case err == nil:
			return fromDb, nil
		case errors.Is(err, sql.ErrNoRows):
		case err != nil:
			return nil, fmt.Errorf("could not lookup change by commit: %w", err)
		}
	}

	var details *vcs.CommitDetails
	var err error

	getCommit := func(repo vcs.RepoGitReader) error {
		details, err = repo.GetCommitDetails(commitID)
		if err != nil {
			return fmt.Errorf("could not get commit details: %w", err)
		}
		return nil
	}
	if err := svc.executorProvider.New().GitRead(getCommit).ExecTrunk(codebaseID, "changeServiceChangelog"); err != nil {
		return nil, err
	}

	// don't import Sturdy-style root commits
	if len(details.Parents) == 0 && details.Message == "Root Commit" {
		return nil, ErrNotFound
	}

	meta := change.ParseCommitMessage(details.Message)
	title := firstLine(meta.Description)

	desc := meta.Description
	desc = strings.ReplaceAll(desc, "\n", "<br>")

	// CreateWithCommitAsParent change!
	ch := change.Change{
		ID:                 change.ID(uuid.NewString()),
		CodebaseID:         codebaseID,
		Title:              &title,
		UpdatedDescription: desc,
		UserID:             nil,
		CreatedAt:          nil, // Set?
		GitCreatedAt:       &details.Author.When,
		GitCreatorName:     &details.Author.Name,
		GitCreatorEmail:    &details.Author.Email,
		CommitID:           &commitID,
		ParentChangeID:     nil, // Parent is starts out as nil. If/when the parent commit is imported, this value will be set.
	}

	if err := svc.changeRepo.Insert(ctx, ch); err != nil {
		return nil, fmt.Errorf("could not write new change to db: %w", err)
	}

	return &ch, nil
}

func firstLine(in string) string {
	idx := strings.IndexByte(in, '\n')
	if idx < 0 {
		return in
	}
	return in[0:idx]
}

func (svc *Service) Diffs(ctx context.Context, ch *change.Change, allower *unidiff.Allower) ([]unidiff.FileDiff, error) {
	parent, err := svc.parentChange(ctx, ch)
	switch {
	case errors.Is(err, ErrNotFound):
		// use diffToRoot
	case err != nil:
		return nil, fmt.Errorf("could not get change parent: %w", err)
	}

	var diff *git.Diff
	diffBetweenCommits := func(repo vcs.RepoGitReader) error {
		diff, err = repo.DiffCommits(*parent.CommitID, *ch.CommitID)
		if err != nil {
			return fmt.Errorf("could not get diffs: %w", err)
		}
		return nil
	}

	diffToRoot := func(repo vcs.RepoGitReader) error {
		diff, err = repo.DiffCommitToRoot(*ch.CommitID)
		if err != nil {
			return fmt.Errorf("could not get diff to root: %w", err)
		}
		return nil
	}

	var fn func(repo vcs.RepoGitReader) error
	if parent != nil {
		fn = diffBetweenCommits
	} else {
		fn = diffToRoot
	}

	err = svc.executorProvider.New().GitRead(fn).ExecTrunk(ch.CodebaseID, "changeService.Diffs")
	if err != nil {
		return nil, err
	}

	decoratedDiff, err := unidiff.NewUnidiff(
		unidiff.NewGitPatchReader(diff),
		svc.logger,
	).WithAllower(allower).Decorate()
	if err != nil {
		return nil, fmt.Errorf("failed to generate unidiff for diff: %w", err)
	}

	return decoratedDiff, nil
}

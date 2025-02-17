package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"getsturdy.com/api/pkg/analytics"
	service_analytics "getsturdy.com/api/pkg/analytics/service"
	"getsturdy.com/api/pkg/auth"
	"getsturdy.com/api/pkg/codebase"
	db_codebase "getsturdy.com/api/pkg/codebase/db"
	"getsturdy.com/api/pkg/codebase/vcs"
	"getsturdy.com/api/pkg/events"
	"getsturdy.com/api/pkg/shortid"
	service_user "getsturdy.com/api/pkg/users/service"
	service_workspace "getsturdy.com/api/pkg/workspaces/service"
	"getsturdy.com/api/vcs/executor"
	"getsturdy.com/api/vcs/provider"
)

type Service struct {
	repo             db_codebase.CodebaseRepository
	codebaseUserRepo db_codebase.CodebaseUserRepository

	workspaceService service_workspace.Service
	userService      service_user.Service

	logger           *zap.Logger
	executorProvider executor.Provider
	eventsSender     events.EventSender
	analyticsService *service_analytics.Service
}

func New(
	repo db_codebase.CodebaseRepository,
	codebaseUserRepo db_codebase.CodebaseUserRepository,

	workspaceService service_workspace.Service,
	userService service_user.Service,

	logger *zap.Logger,
	executorProvider executor.Provider,
	eventsSender events.EventSender,
	analyticsServcie *service_analytics.Service,
) *Service {
	return &Service{
		repo:             repo,
		codebaseUserRepo: codebaseUserRepo,

		workspaceService: workspaceService,
		userService:      userService,

		logger:           logger,
		executorProvider: executorProvider,
		eventsSender:     eventsSender,
		analyticsService: analyticsServcie,
	}
}

func (svc *Service) GetByID(ctx context.Context, id string) (*codebase.Codebase, error) {
	cb, err := svc.repo.Get(id)
	if err != nil {
		return nil, err
	}
	return cb, nil
}

func (svc *Service) GetByShortID(ctx context.Context, shortID string) (*codebase.Codebase, error) {
	cb, err := svc.repo.GetByShortID(shortID)
	if err != nil {
		return nil, err
	}
	return cb, nil
}

func (svc *Service) CanAccess(ctx context.Context, userID string, codebaseID string) (bool, error) {
	_, err := svc.codebaseUserRepo.GetByUserAndCodebase(userID, codebaseID)
	switch {
	case err == nil:
		return true, nil
	case errors.Is(err, sql.ErrNoRows):
		return false, nil
	default:
		return false, fmt.Errorf("failed to check user %s access to codebase %s: %w", userID, codebaseID, err)
	}
}

func (svc *Service) ListByOrganization(ctx context.Context, organizationID string) ([]*codebase.Codebase, error) {
	res, err := svc.repo.ListByOrganization(ctx, organizationID)
	if err != nil {
		return nil, fmt.Errorf("could not ListByOrganization: %w", err)
	}
	return res, nil
}

func (svc *Service) ListByOrganizationAndUser(ctx context.Context, organizationID, userID string) ([]*codebase.Codebase, error) {
	codebases, err := svc.repo.ListByOrganization(ctx, organizationID)
	if err != nil {
		return nil, fmt.Errorf("could not ListByOrganization: %w", err)
	}

	var res []*codebase.Codebase

	for _, cb := range codebases {
		_, err := svc.codebaseUserRepo.GetByUserAndCodebase(userID, cb.ID)
		switch {
		case errors.Is(err, sql.ErrNoRows):
			continue
		case err != nil:
			return nil, fmt.Errorf("could not codebase user: %w", err)
		case err == nil:
			res = append(res, cb)
		}
	}

	return res, nil
}

// ListOrgsByUser returns a list of organization IDs that the user can _see_ through it's explicit membership
// of one of it's codebases.
func (svc *Service) ListOrgsByUser(ctx context.Context, userID string) ([]string, error) {
	orgIDs, err := svc.orgsByUser(ctx, userID)
	if err != nil {
		return nil, err
	}

	var res []string
	for k := range orgIDs {
		res = append(res, k)
	}

	return res, nil
}

func (svc *Service) UserIsMemberOfCodebaseInOrganization(ctx context.Context, userID, organizationID string) (bool, error) {
	orgIDs, err := svc.orgsByUser(ctx, userID)
	if err != nil {
		return false, err
	}

	_, ok := orgIDs[organizationID]
	return ok, nil
}

func (svc *Service) orgsByUser(ctx context.Context, userID string) (map[string]struct{}, error) {
	codebaseUsers, err := svc.codebaseUserRepo.GetByUser(userID)
	if err != nil {
		return nil, fmt.Errorf("could not ListByUser: %w", err)
	}

	orgIDs := make(map[string]struct{})

	for _, cu := range codebaseUsers {
		cb, err := svc.repo.Get(cu.CodebaseID)
		switch {
		case errors.Is(err, sql.ErrNoRows):
			// ignore
			continue
		case err != nil:
			return nil, fmt.Errorf("could not get codebase: %w", err)
		case cb.OrganizationID != nil:
			orgIDs[*cb.OrganizationID] = struct{}{}
		}
	}

	return orgIDs, nil
}

func (svc *Service) Update(ctx context.Context, cb *codebase.Codebase) error {
	if err := svc.repo.Update(cb); err != nil {
		return fmt.Errorf("could not update codebase: %w", err)
	}
	if err := svc.eventsSender.Codebase(cb.ID, events.CodebaseUpdated, cb.ID); err != nil {
		svc.logger.Error("failed to send codebase updated event", zap.Error(err))
	}
	svc.analyticsService.IdentifyCodebase(ctx, cb)
	return nil
}

func (svc *Service) Create(ctx context.Context, name string, organizationID *string) (*codebase.Codebase, error) {
	userID, err := auth.UserID(ctx)
	if err != nil {
		return nil, err
	}

	codebaseID := uuid.NewString()
	t := time.Now()

	cb := codebase.Codebase{
		ID:              codebaseID,
		ShortCodebaseID: codebase.ShortCodebaseID(shortid.New()),
		Name:            name,
		Description:     "",
		Emoji:           "",
		CreatedAt:       &t,
		IsReady:         true,           // No additional setup needed
		OrganizationID:  organizationID, // Optional
	}

	// Create codebase in database
	if err := svc.repo.Create(cb); err != nil {
		return nil, fmt.Errorf("failed to create codebase: %w", err)
	}

	if err := svc.executorProvider.New().
		AllowRebasingState(). // allowed because the repo does not exist yet
		Schedule(func(trunkProvider provider.RepoProvider) error {
			return vcs.Create(trunkProvider, cb.ID)
		}).ExecTrunk(cb.ID, "createCodebase"); err != nil {

		return nil, fmt.Errorf("failed to create codebase on disk: %w", err)
	}

	// Add user
	err = svc.codebaseUserRepo.Create(codebase.CodebaseUser{
		ID:         uuid.New().String(),
		UserID:     userID,
		CodebaseID: cb.ID,
		CreatedAt:  &t,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to add creator as member: %w", err)
	}

	svc.analyticsService.IdentifyCodebase(ctx, &cb)

	opts := []analytics.CaptureOption{
		analytics.CodebaseID(cb.ID),
	}
	if cb.OrganizationID != nil {
		analytics.OrganizationID(*cb.OrganizationID)
	}
	svc.analyticsService.Capture(ctx, "create codebase", opts...)

	if err := svc.workspaceService.CreateWelcomeWorkspace(ctx, cb.ID, userID, cb.Name); err != nil {
		svc.logger.Error("failed to create welcome workspace", zap.Error(err))
		// not a critical error, continue
	}

	// Send events
	if err := svc.eventsSender.Codebase(cb.ID, events.CodebaseUpdated, cb.ID); err != nil {
		return nil, fmt.Errorf("failed to send events: %w", err)
	}

	return &cb, nil
}

func (svc *Service) CodebaseCount(ctx context.Context) (uint64, error) {
	return svc.repo.Count(ctx)
}

func (svc *Service) AddUserByEmail(ctx context.Context, codebaseID, email string) (*codebase.CodebaseUser, error) {
	inviteUser, err := svc.userService.GetByEmail(ctx, email)
	if err != nil {
		return nil, fmt.Errorf("could not get user: %w", err)
	}

	// Check that the user isn't already a member
	_, err = svc.codebaseUserRepo.GetByUserAndCodebase(inviteUser.ID, codebaseID)
	if err == nil {
		return nil, fmt.Errorf("already a member")
	}

	t := time.Now()
	member := codebase.CodebaseUser{
		ID:         uuid.New().String(),
		UserID:     inviteUser.ID,
		CodebaseID: codebaseID,
		CreatedAt:  &t,
	}

	err = svc.codebaseUserRepo.Create(member)
	if err != nil {
		return nil, fmt.Errorf("could not add user: %w", err)
	}

	// Send events
	if err := svc.eventsSender.Codebase(codebaseID, events.CodebaseUpdated, codebaseID); err != nil {
		svc.logger.Error("failed to send events", zap.Error(err))
	}

	svc.analyticsService.Capture(ctx, "add user to codebase",
		analytics.CodebaseID(codebaseID),
		analytics.Property("user_id", inviteUser.ID),
	)

	return &member, nil
}

func (svc *Service) RemoveUser(ctx context.Context, codebaseID, userID string) error {
	member, err := svc.codebaseUserRepo.GetByUserAndCodebase(userID, codebaseID)
	switch {
	case errors.Is(err, sql.ErrNoRows):
		return errors.New("user is not a member")
	case err != nil:
		return fmt.Errorf("failed to remove user: %w", err)
	}

	if err := svc.codebaseUserRepo.DeleteByID(ctx, member.ID); err != nil {
		return fmt.Errorf("failed to delete: %w", err)
	}

	// Send events
	if err := svc.eventsSender.Codebase(codebaseID, events.CodebaseUpdated, codebaseID); err != nil {
		svc.logger.Error("failed to send events", zap.Error(err))
	}

	svc.analyticsService.Capture(ctx, "remove user from codebase",
		analytics.CodebaseID(codebaseID),
		analytics.Property("user_id", userID),
	)

	return nil
}

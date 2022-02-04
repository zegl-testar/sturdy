package vcs

import (
	"fmt"

	"getsturdy.com/api/vcs"
)

func Create(repo vcs.RepoGitWriter, workspaceID string) error {
	if err := repo.CreateNewBranchOnHEAD(workspaceID); err != nil {
		return fmt.Errorf("failed to create workspaceID %s: %w", workspaceID, err)
	}
	return nil
}

func CreateAtChange(repo vcs.RepoGitWriter, workspaceID, changeID string) error {
	if err := repo.CreateNewBranchAt(workspaceID, changeID); err != nil {
		return fmt.Errorf("failed to create workspaceID %s: %w", workspaceID, err)
	}
	return nil
}

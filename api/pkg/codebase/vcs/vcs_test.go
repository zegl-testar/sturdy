package vcs

import (
	"testing"

	"getsturdy.com/api/vcs/testutil"

	"github.com/stretchr/testify/assert"
)

func TestCreateCodebase(t *testing.T) {
	repoProvider := testutil.TestingRepoProvider(t)
	err := Create(repoProvider, "codebaseID")
	assert.NoError(t, err)
}

func TestListChangesInCodebaseTrunk(t *testing.T) {
	repoProvider := testutil.TestingRepoProvider(t)
	codebaseID := "codebaseID"
	err := Create(repoProvider, codebaseID)
	assert.NoError(t, err)
	repo, err := repoProvider.TrunkRepo(codebaseID)
	assert.NoError(t, err)
	log, err := ListChanges(repo, 10)
	assert.NoError(t, err)
	assert.Equal(t, 0, len(log))
}

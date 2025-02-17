package inmemory

import (
	"context"
	"database/sql"

	"getsturdy.com/api/pkg/codebase"
	db_codebase "getsturdy.com/api/pkg/codebase/db"
)

type inMemoryCodebaseRepository struct {
	codebases []codebase.Codebase
}

func NewInMemoryCodebaseRepo() db_codebase.CodebaseRepository {
	return &inMemoryCodebaseRepository{codebases: make([]codebase.Codebase, 0)}
}

func (r *inMemoryCodebaseRepository) Create(entity codebase.Codebase) error {
	r.codebases = append(r.codebases, entity)
	return nil
}

func (r *inMemoryCodebaseRepository) Get(id string) (*codebase.Codebase, error) {
	for _, cb := range r.codebases {
		if cb.ID == id && cb.ArchivedAt == nil {
			return &cb, nil
		}
	}
	return nil, sql.ErrNoRows
}

func (r *inMemoryCodebaseRepository) GetAllowArchived(id string) (*codebase.Codebase, error) {
	for _, cb := range r.codebases {
		if cb.ID == id {
			return &cb, nil
		}
	}
	return nil, sql.ErrNoRows
}

func (r *inMemoryCodebaseRepository) GetByInviteCode(inviteCode string) (*codebase.Codebase, error) {
	for _, cb := range r.codebases {
		if cb.InviteCode == nil && *cb.InviteCode == inviteCode && cb.ArchivedAt == nil {
			return &cb, nil
		}
	}
	return nil, sql.ErrNoRows
}

func (r *inMemoryCodebaseRepository) GetByShortID(shortID string) (*codebase.Codebase, error) {
	for _, cb := range r.codebases {
		if string(cb.ShortCodebaseID) == shortID && cb.ArchivedAt == nil {
			return &cb, nil
		}
	}
	return nil, sql.ErrNoRows
}

func (r *inMemoryCodebaseRepository) Update(entity *codebase.Codebase) error {
	for k, cb := range r.codebases {
		if cb.ID == entity.ID {
			r.codebases[k] = *entity
			return nil
		}
	}
	return sql.ErrNoRows
}

func (r *inMemoryCodebaseRepository) ListByOrganization(_ context.Context, id string) ([]*codebase.Codebase, error) {
	var res []*codebase.Codebase
	for _, cb := range r.codebases {
		if cb.OrganizationID != nil && *cb.OrganizationID == id {
			c2 := cb
			res = append(res, &c2)
		}
	}
	return res, nil
}

func (r *inMemoryCodebaseRepository) Count(_ context.Context) (uint64, error) {
	return uint64(len(r.codebases)), nil
}

package graphql

import (
	"context"
	"errors"

	"getsturdy.com/api/pkg/auth"
	gq_errors "getsturdy.com/api/pkg/graphql/errors"
	"getsturdy.com/api/pkg/graphql/resolvers"
	"getsturdy.com/api/pkg/events"
	"strings"

	"github.com/graph-gophers/graphql-go"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"go.uber.org/zap"
)

var (
	concurrentUpdatedWorkspaceConnections = promauto.NewGauge(prometheus.GaugeOpts{
		Name:        "sturdy_graphql_concurrent_subscriptions",
		ConstLabels: prometheus.Labels{"subscription": "updatedWorkspace"},
	})
)

func (r *WorkspaceRootResolver) UpdatedWorkspace(ctx context.Context, args resolvers.UpdatedWorkspaceArgs) (<-chan resolvers.WorkspaceResolver, error) {
	var codebaseID string
	var workspaceID *string

	userID, err := auth.UserID(ctx)
	if err != nil {
		return nil, gq_errors.Error(err)
	}

	if args.WorkspaceID != nil {
		// Get codebaseID by the workspaceID
		ws, err := r.workspaceReader.Get(string(*args.WorkspaceID))
		if err != nil {
			return nil, gq_errors.Error(err)
		}

		if err := r.authService.CanRead(ctx, ws); err != nil {
			return nil, gq_errors.Error(err)
		}

		codebaseID = ws.CodebaseID
		workspaceID = &ws.ID
	} else if args.ShortCodebaseID != nil {
		s := string(*args.ShortCodebaseID)
		if idx := strings.LastIndex(s, "-"); idx >= 0 {
			s = s[idx+1:]
		}
		cb, err := r.codebaseRepo.GetByShortID(s)
		if err != nil {
			return nil, gq_errors.Error(err)
		}
		if err := r.authService.CanRead(ctx, cb); err != nil {
			return nil, gq_errors.Error(err)
		}
		codebaseID = cb.ID
	} else {
		return nil, gq_errors.Error(gq_errors.ErrBadRequest, "mesasge", "one of shortCodebaseID or workspaceID must be set")
	}

	c := make(chan resolvers.WorkspaceResolver, 100)
	didErrorOut := false

	concurrentUpdatedWorkspaceConnections.Inc()

	listenTo := map[events.EventType]bool{
		events.WorkspaceUpdated:           true,
		events.WorkspaceUpdatedReviews:    true,
		events.WorkspaceUpdatedSuggestion: true,
	}

	cancelFunc := r.viewEvents.SubscribeUser(userID, func(eventType events.EventType, reference string) error {
		if !listenTo[eventType] {
			return nil
		}

		if workspaceID != nil && *workspaceID != reference {
			// Subscribed to a specific workspace which is not this one
			return nil
		}

		resolver, err := r.Workspace(ctx, resolvers.WorkspaceArgs{ID: graphql.ID(reference)})
		if err != nil {
			return err
		}
		select {
		case <-ctx.Done():
			return errors.New("disconnected")
		case c <- resolver:
			if didErrorOut {
				didErrorOut = false
			}
			return nil
		default:
			r.logger.Error("dropped subscription event",
				zap.String("user_id", userID),
				zap.String("codebase_id", codebaseID),
				zap.Stringer("event_type", eventType),
				zap.Int("channel_size", len(c)),
			)
			didErrorOut = true
			return nil
		}
	})

	go func() {
		<-ctx.Done()
		cancelFunc()
		close(c)
		concurrentUpdatedWorkspaceConnections.Dec()
	}()

	return c, nil
}

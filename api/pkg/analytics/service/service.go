package service

import (
	"context"
	"fmt"

	"getsturdy.com/api/pkg/analytics"
	"getsturdy.com/api/pkg/auth"
	"getsturdy.com/api/pkg/codebase"
	"getsturdy.com/api/pkg/organization"
	"getsturdy.com/api/pkg/users"

	"github.com/posthog/posthog-go"
	"go.uber.org/zap"
)

type Service struct {
	logger *zap.Logger

	client posthog.Client
}

func New(
	logger *zap.Logger,
	client posthog.Client,
) *Service {
	return &Service{
		logger: logger,
		client: client,
	}
}

func (s *Service) Capture(ctx context.Context, event string, oo ...analytics.CaptureOption) {
	userID, _ := auth.UserID(ctx)
	options := &analytics.CaptureOptions{
		DistinctId: userID,
	}
	for _, o := range oo {
		o(options)
	}

	s.client.Enqueue(posthog.Capture{
		DistinctId: options.DistinctId,
		Properties: options.Properties,
		Event:      event,
		Groups:     options.Groups,
	})
}

func (s *Service) IdentifyOrganization(ctx context.Context, org *organization.Organization) {
	if err := s.client.Enqueue(posthog.GroupIdentify{
		Type: "organization", // this should match other event's property key
		Key:  org.ID,
		Properties: map[string]interface{}{
			"name": org.Name,
		},
	}); err != nil {
		s.logger.Error("failed to identify codebase", zap.Error(err))
	}
}

func (s *Service) IdentifyCodebase(ctx context.Context, cb *codebase.Codebase) {
	if err := s.client.Enqueue(posthog.GroupIdentify{
		Type: "codebase",
		Key:  cb.ID,
		Properties: map[string]interface{}{
			"name":      cb.Name,
			"is_public": cb.IsPublic,
		},
	}); err != nil {
		s.logger.Error("failed to identify codebase", zap.Error(err))
	}
}

func (s *Service) IdentifyUser(ctx context.Context, user *users.User) {
	if err := s.client.Enqueue(posthog.Identify{
		DistinctId: user.ID,
		Properties: map[string]interface{}{
			"name":  user.Name,
			"email": user.Email,
		},
	}); err != nil {
		s.logger.Error("failed to identify user", zap.Error(err))
	}
}

func (s *Service) IdentifyGitHubInstallation(ctx context.Context, installationID int64, accountLogin, accountEmail string) {
	if err := s.client.Enqueue(posthog.Identify{
		DistinctId: fmt.Sprintf("%d", installationID), // Using the installation ID as a person?
		Properties: map[string]interface{}{
			"installation_org":        accountLogin,
			"email":                   accountEmail,
			"github_app_installation": true,
		},
	}); err != nil {
		s.logger.Error("failed to identify github installation", zap.Error(err))
	}
}

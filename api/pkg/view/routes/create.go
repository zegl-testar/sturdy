package routes

import (
	"net/http"
	"time"

	"getsturdy.com/api/pkg/analytics"
	service_analytics "getsturdy.com/api/pkg/analytics/service"
	"getsturdy.com/api/pkg/auth"
	"getsturdy.com/api/pkg/codebase/access"
	db_codebase "getsturdy.com/api/pkg/codebase/db"
	"getsturdy.com/api/pkg/events"
	db_snapshots "getsturdy.com/api/pkg/snapshots/db"
	"getsturdy.com/api/pkg/snapshots/snapshotter"
	"getsturdy.com/api/pkg/view"
	"getsturdy.com/api/pkg/view/db"
	"getsturdy.com/api/pkg/view/open"
	"getsturdy.com/api/pkg/view/vcs"
	db_workspaces "getsturdy.com/api/pkg/workspaces/db"
	"getsturdy.com/api/vcs/executor"
	"getsturdy.com/api/vcs/provider"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type CreateRequest struct {
	CodebaseID    string `json:"codebase_id" binding:"required"`
	WorkspaceID   string `json:"workspace_id" binding:"required"`
	Name          string `json:"name"`
	MountPath     string `json:"mount_path"`
	MountHostname string `json:"mount_hostname"`
}

func Create(
	logger *zap.Logger,
	viewRepo db.Repository,
	codebaseUserRepo db_codebase.CodebaseUserRepository,
	analyticsService *service_analytics.Service,
	workspaceReader db_workspaces.WorkspaceReader,
	snapshotter snapshotter.Snapshotter,
	snapshotRepo db_snapshots.Repository,
	workspaceWriter db_workspaces.WorkspaceWriter,
	executorProvider executor.Provider,
	eventSender events.EventSender,
) func(c *gin.Context) {
	return func(c *gin.Context) {
		var req CreateRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			logger.Warn("failed to parse input", zap.Error(err))
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "failed to parse or validate input"})
			return
		}

		userID, err := auth.UserID(c.Request.Context())
		if err != nil {
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		if !access.UserHasAccessToCodebase(codebaseUserRepo, userID, req.CodebaseID) {
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		t := time.Now()
		e := view.View{
			ID:            uuid.New().String(),
			UserID:        userID,
			CodebaseID:    req.CodebaseID,
			WorkspaceID:   req.WorkspaceID,
			Name:          &req.Name,          // It's optional
			MountPath:     &req.MountPath,     // It's optional
			MountHostname: &req.MountHostname, // It's optional
			CreatedAt:     &t,
		}

		if err := viewRepo.Create(e); err != nil {
			logger.Error("failed to create view", zap.Error(err))
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "failed to create view"})
			return
		}

		if err := executorProvider.New().
			AllowRebasingState(). // allowed because the view does not exist yet
			Schedule(func(repoProvider provider.RepoProvider) error {
				return vcs.Create(repoProvider, req.CodebaseID, req.WorkspaceID, e.ID)
			}).ExecView(req.CodebaseID, e.ID, "createView"); err != nil {
			logger.Error("failed to create view", zap.Error(err))
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "failed to create view"})
			return
		}

		ws, err := workspaceReader.Get(req.WorkspaceID)
		if err != nil {
			logger.Error("failed to get workspace", zap.Error(err))
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}

		// Use workspace on view
		if err := open.OpenWorkspaceOnView(c.Request.Context(), logger, &e, ws, viewRepo, workspaceReader, snapshotter, snapshotRepo, workspaceWriter, executorProvider, eventSender); err != nil {
			logger.Error("failed to open workspace on view", zap.Error(err))
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}

		analyticsService.Capture(c.Request.Context(), "create view",
			analytics.CodebaseID(req.CodebaseID),
			analytics.Property("workspace_id", req.WorkspaceID),
			analytics.Property("view_id", e.ID),
			analytics.Property("mount_path", req.MountPath),
			analytics.Property("mount_hostname", req.MountHostname),
		)

		c.JSON(http.StatusOK, e)
	}
}

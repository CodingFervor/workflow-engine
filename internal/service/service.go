package service

import (
	"github.com/CodingFervor/workflow-engine/internal/repository"
	"github.com/CodingFervor/workflow-engine/pkg/logger"
)

// Context holds shared dependencies for all services.
type Context struct {
	Repo *repository.Base
}

func NewContext() *Context {
	return &Context{Repo: &repository.Base{}}
}

// HealthCheck verifies all dependencies are reachable.
func (c *Context) HealthCheck() map[string]string {
	status := map[string]string{}
	if err := c.Repo.DB().Ping(); err != nil {
		status["database"] = "unhealthy"
		logger.Error("database health check failed", "error", err)
	} else {
		status["database"] = "healthy"
	}
	status["server"] = "healthy"
	return status
}

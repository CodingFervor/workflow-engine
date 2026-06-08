package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/CodingFervor/workflow-engine/internal/cache"
	"github.com/CodingFervor/workflow-engine/internal/config"
	"github.com/CodingFervor/workflow-engine/internal/database"
	"github.com/CodingFervor/workflow-engine/internal/middleware"
	"github.com/CodingFervor/workflow-engine/internal/service"
	"github.com/CodingFervor/workflow-engine/pkg/jwt"
	"github.com/CodingFervor/workflow-engine/pkg/logger"
)

var svc *service.Context

func main() {
	// Load config
	cfg, err := config.Load("configs/config.yaml")
	if err != nil {
		log.Fatalf("config load error: %v", err)
	}

	// Logger
	logger.SetLevel(cfg.Server.Mode)
	logger.Info("starting " + "workflow-engine")

	// Database
	if err := database.Connect(cfg.Database); err != nil {
		logger.Error("database connect failed", "error", err)
		log.Fatalf("database: %v", err)
	}
	defer database.Close()

	// Redis
	if err := cache.Connect(cfg.Redis); err != nil {
		logger.Warn("redis connect failed, running without cache", "error", err)
	}
	defer cache.Close()

	// JWT & Service
	jwt.SetSecret(cfg.JWT.Secret)
	svc = service.NewContext()

	// Gin
	if cfg.Server.Mode == "release" {
		gin.SetMode(gin.ReleaseMode)
	}
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(middleware.CORS())
	r.Use(requestLogger())

	r.GET("/health", healthCheck)
	r.GET("/ready", readinessCheck)

	api := r.Group("/api/v1")
	{

		api.POST("/auth/login", Login)

		auth := api.Group("/")
		auth.Use(middleware.AuthMiddleware())
		{
			auth.GET("/process-definitions", ListProcessDefinitions)
			auth.POST("/process-definitions", DeployProcessDefinition)
			auth.GET("/process-definitions/:id", GetProcessDefinition)
			auth.PUT("/process-definitions/:id", UpdateProcessDefinition)
			auth.DELETE("/process-definitions/:id", DeleteProcessDefinition)
			auth.GET("/process-definitions/:id/versions", ListProcessVersions)
			auth.POST("/process-definitions/:id/activate", ActivateProcess)

			auth.POST("/process-instances", StartProcessInstance)
			auth.GET("/process-instances", ListProcessInstances)
			auth.GET("/process-instances/:id", GetProcessInstance)
			auth.PUT("/process-instances/:id/pause", PauseInstance)
			auth.PUT("/process-instances/:id/resume", ResumeInstance)
			auth.PUT("/process-instances/:id/terminate", TerminateInstance)
			auth.GET("/process-instances/:id/history", InstanceHistory)
			auth.GET("/process-instances/:id/variables", InstanceVariables)

			auth.GET("/tasks", ListTasks)
			auth.GET("/tasks/my", MyTasks)
			auth.GET("/tasks/candidate", CandidateTasks)
			auth.GET("/tasks/:id", GetTask)
			auth.PUT("/tasks/:id/claim", ClaimTask)
			auth.PUT("/tasks/:id/unclaim", UnclaimTask)
			auth.PUT("/tasks/:id/complete", CompleteTask)
			auth.PUT("/tasks/:id/delegate", DelegateTask)
			auth.PUT("/tasks/:id/transfer", TransferTask)
			auth.POST("/tasks/:id/comments", AddTaskComment)
			auth.GET("/tasks/:id/comments", ListTaskComments)

			auth.GET("/users", ListUsers)
			auth.POST("/users", CreateUser)
			auth.GET("/groups", ListGroups)
			auth.POST("/groups", CreateGroup)
			auth.POST("/groups/:id/members", AddGroupMember)

			auth.GET("/tenants", ListTenants)
			auth.POST("/tenants", CreateTenant)

			auth.GET("/deployments", ListDeployments)
			auth.DELETE("/deployments/:id", DeleteDeployment)

			auth.GET("/jobs", ListJobs)
			auth.POST("/jobs/:id/execute", ExecuteJob)
			auth.DELETE("/jobs/:id", DeleteJob)

			auth.GET("/event-subscriptions", ListEventSubscriptions)

			auth.GET("/dashboard", Dashboard)
			auth.GET("/analytics/instances", InstanceAnalytics)
			auth.GET("/analytics/tasks", TaskAnalytics)
			auth.GET("/analytics/duration", DurationAnalytics)
		}
	}

	// Graceful shutdown
	addr := ":" + strconv.Itoa(cfg.Server.Port)
	srv := &http.Server{Addr: addr, Handler: r}
	go func() {
		logger.Info("server listening", "port", cfg.Server.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("server error", "error", err)
		}
	}()
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	logger.Info("shutting down server...")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		logger.Error("forced shutdown", "error", err)
	}
	logger.Info("server exited")
}

func requestLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()
		logger.Info(fmt.Sprintf("%s %s %d", c.Request.Method, c.Request.URL.Path, c.Writer.Status()),
			"method", c.Request.Method,
			"path", c.Request.URL.Path,
			"status", c.Writer.Status(),
			"latency", time.Since(start).String(),
			"ip", c.ClientIP(),
		)
	}
}

func healthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status": "ok",
		"time":   time.Now().Format(time.RFC3339),
	})
}

func readinessCheck(c *gin.Context) {
	status := svc.HealthCheck()
	code := http.StatusOK
	for _, v := range status {
		if v != "healthy" {
			code = http.StatusServiceUnavailable
			break
		}
	}
	c.JSON(code, gin.H{"checks": status})
}

func Login(c *gin.Context)   { c.JSON(http.StatusOK, gin.H{"message": "login"}) }

func ListProcessDefinitions(c *gin.Context)    { c.JSON(http.StatusOK, gin.H{"data": []gin.H{}}) }
func DeployProcessDefinition(c *gin.Context)   { c.JSON(http.StatusCreated, gin.H{"message": "process deployed"}) }
func GetProcessDefinition(c *gin.Context)      { c.JSON(http.StatusOK, gin.H{"data": gin.H{}}) }
func UpdateProcessDefinition(c *gin.Context)   { c.JSON(http.StatusOK, gin.H{"message": "updated"}) }
func DeleteProcessDefinition(c *gin.Context)   { c.JSON(http.StatusOK, gin.H{"message": "deleted"}) }
func ListProcessVersions(c *gin.Context)       { c.JSON(http.StatusOK, gin.H{"data": []gin.H{}}) }
func ActivateProcess(c *gin.Context)           { c.JSON(http.StatusOK, gin.H{"message": "activated"}) }

func StartProcessInstance(c *gin.Context)      { c.JSON(http.StatusCreated, gin.H{"message": "instance started"}) }
func ListProcessInstances(c *gin.Context)      { c.JSON(http.StatusOK, gin.H{"data": []gin.H{}}) }
func GetProcessInstance(c *gin.Context)        { c.JSON(http.StatusOK, gin.H{"data": gin.H{}}) }
func PauseInstance(c *gin.Context)             { c.JSON(http.StatusOK, gin.H{"message": "paused"}) }
func ResumeInstance(c *gin.Context)            { c.JSON(http.StatusOK, gin.H{"message": "resumed"}) }
func TerminateInstance(c *gin.Context)         { c.JSON(http.StatusOK, gin.H{"message": "terminated"}) }
func InstanceHistory(c *gin.Context)           { c.JSON(http.StatusOK, gin.H{"data": []gin.H{}}) }
func InstanceVariables(c *gin.Context)         { c.JSON(http.StatusOK, gin.H{"data": []gin.H{}}) }

func ListTasks(c *gin.Context)         { c.JSON(http.StatusOK, gin.H{"data": []gin.H{}}) }
func MyTasks(c *gin.Context)           { c.JSON(http.StatusOK, gin.H{"data": []gin.H{}}) }
func CandidateTasks(c *gin.Context)    { c.JSON(http.StatusOK, gin.H{"data": []gin.H{}}) }
func GetTask(c *gin.Context)           { c.JSON(http.StatusOK, gin.H{"data": gin.H{}}) }
func ClaimTask(c *gin.Context)         { c.JSON(http.StatusOK, gin.H{"message": "task claimed"}) }
func UnclaimTask(c *gin.Context)       { c.JSON(http.StatusOK, gin.H{"message": "task unclaimed"}) }
func CompleteTask(c *gin.Context)      { c.JSON(http.StatusOK, gin.H{"message": "task completed"}) }
func DelegateTask(c *gin.Context)      { c.JSON(http.StatusOK, gin.H{"message": "task delegated"}) }
func TransferTask(c *gin.Context)      { c.JSON(http.StatusOK, gin.H{"message": "task transferred"}) }
func AddTaskComment(c *gin.Context)    { c.JSON(http.StatusCreated, gin.H{"message": "comment added"}) }
func ListTaskComments(c *gin.Context)  { c.JSON(http.StatusOK, gin.H{"data": []gin.H{}}) }

func ListUsers(c *gin.Context)         { c.JSON(http.StatusOK, gin.H{"data": []gin.H{}}) }
func CreateUser(c *gin.Context)        { c.JSON(http.StatusCreated, gin.H{"message": "user created"}) }
func ListGroups(c *gin.Context)        { c.JSON(http.StatusOK, gin.H{"data": []gin.H{}}) }
func CreateGroup(c *gin.Context)       { c.JSON(http.StatusCreated, gin.H{"message": "group created"}) }
func AddGroupMember(c *gin.Context)    { c.JSON(http.StatusOK, gin.H{"message": "member added"}) }

func ListTenants(c *gin.Context)       { c.JSON(http.StatusOK, gin.H{"data": []gin.H{}}) }
func CreateTenant(c *gin.Context)      { c.JSON(http.StatusCreated, gin.H{"message": "tenant created"}) }

func ListDeployments(c *gin.Context)   { c.JSON(http.StatusOK, gin.H{"data": []gin.H{}}) }
func DeleteDeployment(c *gin.Context)  { c.JSON(http.StatusOK, gin.H{"message": "deployment deleted"}) }

func ListJobs(c *gin.Context)          { c.JSON(http.StatusOK, gin.H{"data": []gin.H{}}) }
func ExecuteJob(c *gin.Context)        { c.JSON(http.StatusOK, gin.H{"message": "job executed"}) }
func DeleteJob(c *gin.Context)         { c.JSON(http.StatusOK, gin.H{"message": "job deleted"}) }
func ListEventSubscriptions(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"data": []gin.H{}}) }

func Dashboard(c *gin.Context)         { c.JSON(http.StatusOK, gin.H{"data": gin.H{}}) }
func InstanceAnalytics(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"data": []gin.H{}}) }
func TaskAnalytics(c *gin.Context)     { c.JSON(http.StatusOK, gin.H{"data": []gin.H{}}) }
func DurationAnalytics(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"data": []gin.H{}}) }

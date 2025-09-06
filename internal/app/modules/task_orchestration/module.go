package task_orchestration

import (
	"assistant/internal/app/module"
	"assistant/internal/app/repo"
	"assistant/internal/db"

	"github.com/gin-gonic/gin"
)

type TaskOrchestrationModule struct {
	taskAtomicSvc *TaskAtomicService
	workflowSvc   *WorkflowService
	scheduleSvc   *ScheduleService
}

func NewTaskOrchestrationModule() module.Module {
	return &TaskOrchestrationModule{
		taskAtomicSvc: NewTaskAtomicService(repo.New(db.GetDB())),
		workflowSvc:   NewWorkflowService(repo.New(db.GetDB())),
		scheduleSvc:   NewScheduleService(repo.New(db.GetDB())),
	}
}

func (m *TaskOrchestrationModule) Name() string { return "task_orchestration" }

func (m *TaskOrchestrationModule) Register(r *gin.Engine) {
	api := r.Group("/api/v1")

	atomicTasks := api.Group("/task_atomic")
	atomicTasks.GET("/count", m.taskAtomicSvc.Count)
	atomicTasks.POST("/create", m.taskAtomicSvc.Create)
	atomicTasks.GET("/get/:id", m.taskAtomicSvc.GetByID)
	atomicTasks.GET("/list", m.taskAtomicSvc.List)
	atomicTasks.PUT("/update/:id", m.taskAtomicSvc.Update)
	atomicTasks.DELETE("/delete/:id", m.taskAtomicSvc.Delete)

	workflows := api.Group("/workflow")
	workflows.GET("/count", m.workflowSvc.Count)
	workflows.POST("/create", m.workflowSvc.Create)
	workflows.GET("/get/:id", m.workflowSvc.GetByID)
	workflows.GET("/list", m.workflowSvc.List)
	workflows.PUT("/update/:id", m.workflowSvc.Update)
	workflows.DELETE("/delete/:id", m.workflowSvc.Delete)
	workflows.POST("/start/:id", m.workflowSvc.Start)
	workflows.POST("/pause/:id", m.workflowSvc.Pause)
	workflows.POST("/resume/:id", m.workflowSvc.Resume)
	workflows.POST("/cancel/:id", m.workflowSvc.Cancel)
	workflows.POST("/rollback/:id", m.workflowSvc.Rollback)
	workflows.GET("/get/:id/nodes", m.workflowSvc.GetNodes)
	workflows.GET("/get/:id/edges", m.workflowSvc.GetEdges)

	schedules := api.Group("/schedule")
	schedules.GET("/count", m.scheduleSvc.Count)
	schedules.POST("/create", m.scheduleSvc.Create)
	schedules.GET("/get/:id", m.scheduleSvc.GetByID)
	schedules.GET("/list", m.scheduleSvc.List)
	schedules.PUT("/update/:id", m.scheduleSvc.Update)
	schedules.DELETE("/delete/:id", m.scheduleSvc.Delete)
	schedules.POST("/enable/:id", m.scheduleSvc.Enable)
	schedules.POST("/disable/:id", m.scheduleSvc.Disable)

	instances := api.Group("/workflow_instance")
	instances.GET("/count", m.workflowSvc.CountInstance)
	instances.GET("/get/:id", m.workflowSvc.GetInstanceByID)
	instances.GET("/list", m.workflowSvc.ListInstances)
	instances.GET("/get/:id/logs", m.workflowSvc.GetExecutionLogs)
}

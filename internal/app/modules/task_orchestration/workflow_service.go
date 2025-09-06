package task_orchestration

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"time"

	"assistant/internal/app/repo"
	"assistant/pkg/response"

	"github.com/gin-gonic/gin"
)

type WorkflowService struct {
	q *repo.Queries
}

func NewWorkflowService(q *repo.Queries) *WorkflowService {
	return &WorkflowService{q: q}
}

type CreateWorkflowRequest struct {
	Name               string          `json:"name" binding:"required"`
	Description        string          `json:"description"`
	WorkflowType       string          `json:"workflow_type"`
	GraphConfig        json.RawMessage `json:"graph_config" binding:"required"`
	Parameters         json.RawMessage `json:"parameters"`
	Timeout            int             `json:"timeout"`
	OnErrorStrategy    string          `json:"on_error_strategy"`
	NotificationConfig json.RawMessage `json:"notification_config"`
}

type CreateWorkflowNodeRequest struct {
	NodeID          string          `json:"node_id" binding:"required"`
	NodeType        string          `json:"node_type" binding:"required"`
	DisplayName     string          `json:"display_name" binding:"required"`
	AtomicTaskDefID int64           `json:"task_atomic_def_id"`
	SubWorkflowID   int64           `json:"sub_workflow_id"`
	ConditionExpr   string          `json:"condition_expr"`
	NodeConfig      json.RawMessage `json:"node_config"`
	RetryPolicy     json.RawMessage `json:"retry_policy"`
	Timeout         int             `json:"timeout"`
	Ord             int             `json:"ord"`
}

type CreateWorkflowEdgeRequest struct {
	FromNodeID    string `json:"from_node_id" binding:"required"`
	ToNodeID      string `json:"to_node_id" binding:"required"`
	EdgeType      string `json:"edge_type"`
	ConditionExpr string `json:"condition_expr"`
}

type StartWorkflowRequest struct {
	InputParams   json.RawMessage `json:"input_params"`
	ExecutionMode string          `json:"execution_mode"`
	Priority      int             `json:"priority"`
}

// CountWorkflows godoc
// @Summary 获取工作流定义总数
// @Description 统计系统中所有工作流定义数量
// @Tags 工作流管理
// @Produce json
// @Success 200 {object} response.Response "成功"
// @Failure 500 {object} response.Response "服务器错误"
// @Router /workflow/count [get]
func (s *WorkflowService) Count(c *gin.Context) {
	cnt, err := s.q.CountTaskWorkflowDefs(c)
	if err != nil {
		response.Err(c, http.StatusInternalServerError, err.Error())
		return
	}
	response.Ok(c, cnt)
}

// CreateWorkflow godoc
// @Summary 创建工作流
// @Description 创建一个新的工作流定义，支持 DAG/SEQUENTIAL/PARALLEL/CONDITIONAL 类型
// @Tags 工作流管理
// @Accept json
// @Produce json
// @Param data body CreateWorkflowRequest true "工作流创建参数"
// @Success 200 {object} response.Response "成功"
// @Failure 400 {object} response.Response "请求参数错误"
// @Failure 500 {object} response.Response "服务器错误"
// @Router /workflow/create [post]
func (s *WorkflowService) Create(c *gin.Context) {
	var req CreateWorkflowRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Err(c, http.StatusBadRequest, err.Error())
		return
	}

	params := repo.CreateTaskWorkflowDefParams{
		Name:         req.Name,
		Description:  sql.NullString{String: req.Description, Valid: req.Description != ""},
		Version:      sql.NullInt32{Int32: 1, Valid: true},
		WorkflowType: repo.TaskWorkflowDefWorkflowType(req.WorkflowType),
		GraphConfig:  req.GraphConfig,
		Parameters:   req.Parameters,
		Status:       repo.NullTaskWorkflowDefStatus{TaskWorkflowDefStatus: "DRAFT", Valid: true},
	}

	if req.Timeout > 0 {
		params.Timeout = sql.NullInt32{Int32: int32(req.Timeout), Valid: true}
	} else {
		params.Timeout = sql.NullInt32{Int32: 3600, Valid: true}
	}
	if req.OnErrorStrategy != "" {
		params.OnErrorStrategy = repo.NullTaskWorkflowDefOnErrorStrategy{TaskWorkflowDefOnErrorStrategy: repo.TaskWorkflowDefOnErrorStrategy(req.OnErrorStrategy), Valid: true}
	} else {
		params.OnErrorStrategy = repo.NullTaskWorkflowDefOnErrorStrategy{TaskWorkflowDefOnErrorStrategy: "STOP", Valid: true}
	}
	if req.NotificationConfig != nil {
		params.NotificationConfig = req.NotificationConfig
	}

	result, err := s.q.CreateTaskWorkflowDef(c, params)
	if err != nil {
		response.Err(c, http.StatusInternalServerError, err.Error())
		return
	}

	workflowID, _ := result.LastInsertId()
	response.Ok(c, workflowID)
}

// GetWorkflowByID godoc
// @Summary 获取工作流详情
// @Description 根据 ID 获取工作流定义详细信息
// @Tags 工作流管理
// @Produce json
// @Param id path int true "工作流ID"
// @Success 200 {object} response.Response "成功"
// @Failure 400 {object} response.Response "请求参数错误"
// @Failure 404 {object} response.Response "工作流不存在"
// @Failure 500 {object} response.Response "服务器错误"
// @Router /workflow/get/{id} [get]
func (s *WorkflowService) GetByID(c *gin.Context) {
	id, err := parseID(c)
	if err != nil {
		response.Err(c, http.StatusBadRequest, err.Error())
		return
	}

	workflow, err := s.q.GetTaskWorkflowDefByID(c, id)
	if err != nil {
		if err == sql.ErrNoRows {
			response.Err(c, http.StatusNotFound, "workflow not found")
			return
		}
		response.Err(c, http.StatusInternalServerError, err.Error())
		return
	}
	response.Ok(c, workflow)
}

// ListWorkflows godoc
// @Summary 获取工作流列表
// @Description 分页查询工作流定义列表
// @Tags 工作流管理
// @Accept json
// @Produce json
// @Param limit query int false "每页数量 (默认20, 最大100)" default(20)
// @Param offset query int false "偏移量" default(0)
// @Success 200 {object} response.Response "成功"
// @Failure 500 {object} response.Response "服务器错误"
// @Router /workflow/list [get]
func (s *WorkflowService) List(c *gin.Context) {
	limit, offset := getPagination(c)
	workflows, err := s.q.ListTaskWorkflowDefs(c, repo.ListTaskWorkflowDefsParams{Limit: int32(limit), Offset: int32(offset)})
	if err != nil {
		response.Err(c, http.StatusInternalServerError, err.Error())
		return
	}
	response.Ok(c, workflows)
}

// UpdateWorkflow godoc
// @Summary 更新工作流
// @Description 根据 ID 更新工作流定义信息
// @Tags 工作流管理
// @Accept json
// @Produce json
// @Param id path int true "工作流ID"
// @Param data body CreateWorkflowRequest true "更新参数"
// @Success 200 {object} response.Response "成功"
// @Failure 400 {object} response.Response "请求参数错误"
// @Failure 500 {object} response.Response "服务器错误"
// @Router /workflow/update/{id} [put]
func (s *WorkflowService) Update(c *gin.Context) {
	id, err := parseID(c)
	if err != nil {
		response.Err(c, http.StatusBadRequest, err.Error())
		return
	}

	var req CreateWorkflowRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Err(c, http.StatusBadRequest, err.Error())
		return
	}

	params := repo.UpdateTaskWorkflowDefByIDParams{
		ID:           id,
		Name:         req.Name,
		Description:  sql.NullString{String: req.Description, Valid: req.Description != ""},
		WorkflowType: repo.TaskWorkflowDefWorkflowType(req.WorkflowType),
		GraphConfig:  req.GraphConfig,
		Parameters:   req.Parameters,
		Status:       repo.NullTaskWorkflowDefStatus{TaskWorkflowDefStatus: "DRAFT", Valid: true},
	}

	if req.Timeout > 0 {
		params.Timeout = sql.NullInt32{Int32: int32(req.Timeout), Valid: true}
	}
	if req.OnErrorStrategy != "" {
		params.OnErrorStrategy = repo.NullTaskWorkflowDefOnErrorStrategy{TaskWorkflowDefOnErrorStrategy: repo.TaskWorkflowDefOnErrorStrategy(req.OnErrorStrategy), Valid: true}
	}
	if req.NotificationConfig != nil {
		params.NotificationConfig = req.NotificationConfig
	}

	_, err = s.q.UpdateTaskWorkflowDefByID(c, params)
	if err != nil {
		response.Err(c, http.StatusInternalServerError, err.Error())
		return
	}
	response.Ok(c, nil)
}

// DeleteWorkflow godoc
// @Summary 删除工作流
// @Description 根据 ID 删除工作流定义
// @Tags 工作流管理
// @Produce json
// @Param id path int true "工作流ID"
// @Success 200 {object} response.Response "成功"
// @Failure 400 {object} response.Response "请求参数错误"
// @Failure 500 {object} response.Response "服务器错误"
// @Router /workflow/delete/{id} [delete]
func (s *WorkflowService) Delete(c *gin.Context) {
	id, err := parseID(c)
	if err != nil {
		response.Err(c, http.StatusBadRequest, err.Error())
		return
	}

	_, err = s.q.DeleteTaskWorkflowDefByID(c, id)
	if err != nil {
		response.Err(c, http.StatusInternalServerError, err.Error())
		return
	}
	response.Ok(c, nil)
}

// GetWorkflowNodes godoc
// @Summary 获取工作流节点列表
// @Description 根据工作流 ID 获取所有节点定义
// @Tags 工作流管理
// @Produce json
// @Param id path int true "工作流ID"
// @Success 200 {object} response.Response "成功"
// @Failure 400 {object} response.Response "请求参数错误"
// @Failure 500 {object} response.Response "服务器错误"
// @Router /workflow/get/{id}/nodes [get]
func (s *WorkflowService) GetNodes(c *gin.Context) {
	id, err := parseID(c)
	if err != nil {
		response.Err(c, http.StatusBadRequest, err.Error())
		return
	}

	nodes, err := s.q.ListTaskWorkflowNodes(c, id)
	if err != nil {
		response.Err(c, http.StatusInternalServerError, err.Error())
		return
	}
	response.Ok(c, nodes)
}

// GetWorkflowEdges godoc
// @Summary 获取工作流边列表
// @Description 根据工作流 ID 获取所有边定义
// @Tags 工作流管理
// @Produce json
// @Param id path int true "工作流ID"
// @Success 200 {object} response.Response "成功"
// @Failure 400 {object} response.Response "请求参数错误"
// @Failure 500 {object} response.Response "服务器错误"
// @Router /workflow/get/{id}/edges [get]
func (s *WorkflowService) GetEdges(c *gin.Context) {
	id, err := parseID(c)
	if err != nil {
		response.Err(c, http.StatusBadRequest, err.Error())
		return
	}

	edges, err := s.q.ListTaskWorkflowEdges(c, id)
	if err != nil {
		response.Err(c, http.StatusInternalServerError, err.Error())
		return
	}
	response.Ok(c, edges)
}

// StartWorkflow godoc
// @Summary 启动工作流
// @Description 根据工作流定义 ID 启动一个新的工作流实例
// @Tags 工作流执行
// @Accept json
// @Produce json
// @Param id path int true "工作流定义ID"
// @Param data body StartWorkflowRequest true "启动参数"
// @Success 200 {object} response.Response "成功"
// @Failure 400 {object} response.Response "请求参数错误"
// @Failure 404 {object} response.Response "工作流不存在"
// @Failure 500 {object} response.Response "服务器错误"
// @Router /workflow/start/{id} [post]
func (s *WorkflowService) Start(c *gin.Context) {
	workflowID, err := parseID(c)
	if err != nil {
		response.Err(c, http.StatusBadRequest, err.Error())
		return
	}

	var req StartWorkflowRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Err(c, http.StatusBadRequest, err.Error())
		return
	}

	workflow, err := s.q.GetTaskWorkflowDefByID(c, workflowID)
	if err != nil {
		response.Err(c, http.StatusNotFound, "workflow not found")
		return
	}

	executionMode := repo.TaskWorkflowInstanceExecutionModeASYNCHRONOUS
	if req.ExecutionMode == "SYNCHRONOUS" {
		executionMode = repo.TaskWorkflowInstanceExecutionModeSYNCHRONOUS
	}
	priority := req.Priority
	if priority <= 0 {
		priority = 5
	}

	instanceParams := repo.CreateTaskWorkflowInstanceParams{
		WorkflowDefID:      workflowID,
		WorkflowDefVersion: workflow.Version,
		TriggerType:        repo.TaskWorkflowInstanceTriggerTypeMANUAL,
		InputParams:        req.InputParams,
		Status:             repo.TaskWorkflowInstanceStatusPENDING,
		ExecutionMode:      repo.NullTaskWorkflowInstanceExecutionMode{TaskWorkflowInstanceExecutionMode: executionMode, Valid: true},
		Priority:           sql.NullInt32{Int32: int32(priority), Valid: true},
	}

	result, err := s.q.CreateTaskWorkflowInstance(c, instanceParams)
	if err != nil {
		response.Err(c, http.StatusInternalServerError, err.Error())
		return
	}

	instanceID, _ := result.LastInsertId()

	nodes, _ := s.q.ListTaskWorkflowNodes(c, workflowID)
	for _, node := range nodes {
		_, _ = s.q.CreateTaskNodeInstance(c, repo.CreateTaskNodeInstanceParams{
			WorkflowInstanceID: instanceID,
			NodeDefID:          node.ID,
			NodeID:             node.NodeID,
			Status:             repo.TaskNodeInstanceStatusPENDING,
			InputParams:        req.InputParams,
		})
	}

	_, _ = s.q.UpdateTaskWorkflowInstanceProgress(c, repo.UpdateTaskWorkflowInstanceProgressParams{
		TotalNodes:     sql.NullInt32{Int32: int32(len(nodes)), Valid: true},
		CompletedNodes: sql.NullInt32{Int32: 0, Valid: true},
		FailedNodes:    sql.NullInt32{Int32: 0, Valid: true},
		ID:             instanceID,
	})

	response.Ok(c, map[string]interface{}{
		"instance_id": instanceID,
		"status":      "PENDING",
	})
}

// PauseWorkflow godoc
// @Summary 暂停工作流
// @Description 暂停正在运行的工作流实例
// @Tags 工作流执行
// @Produce json
// @Param id path int true "工作流实例ID"
// @Success 200 {object} response.Response "成功"
// @Failure 400 {object} response.Response "请求参数错误"
// @Failure 500 {object} response.Response "服务器错误"
// @Router /workflow/pause/{id} [post]
func (s *WorkflowService) Pause(c *gin.Context) {
	instanceID, err := parseID(c)
	if err != nil {
		response.Err(c, http.StatusBadRequest, err.Error())
		return
	}

	_, err = s.q.UpdateTaskWorkflowInstanceStatus(c, repo.UpdateTaskWorkflowInstanceStatusParams{
		Status:    repo.TaskWorkflowInstanceStatusPAUSED,
		GmtPaused: sql.NullTime{Time: time.Now(), Valid: true},
		ID:        instanceID,
	})
	if err != nil {
		response.Err(c, http.StatusInternalServerError, err.Error())
		return
	}
	response.Ok(c, nil)
}

// ResumeWorkflow godoc
// @Summary 恢复工作流
// @Description 恢复已暂停的工作流实例
// @Tags 工作流执行
// @Produce json
// @Param id path int true "工作流实例ID"
// @Success 200 {object} response.Response "成功"
// @Failure 400 {object} response.Response "请求参数错误"
// @Failure 500 {object} response.Response "服务器错误"
// @Router /workflow/resume/{id} [post]
func (s *WorkflowService) Resume(c *gin.Context) {
	instanceID, err := parseID(c)
	if err != nil {
		response.Err(c, http.StatusBadRequest, err.Error())
		return
	}

	_, err = s.q.UpdateTaskWorkflowInstanceStatus(c, repo.UpdateTaskWorkflowInstanceStatusParams{
		Status:    repo.TaskWorkflowInstanceStatusRUNNING,
		GmtPaused: sql.NullTime{},
		ID:        instanceID,
	})
	if err != nil {
		response.Err(c, http.StatusInternalServerError, err.Error())
		return
	}
	response.Ok(c, nil)
}

// CancelWorkflow godoc
// @Summary 取消工作流
// @Description 取消正在运行或暂停的工作流实例
// @Tags 工作流执行
// @Produce json
// @Param id path int true "工作流实例ID"
// @Success 200 {object} response.Response "成功"
// @Failure 400 {object} response.Response "请求参数错误"
// @Failure 500 {object} response.Response "服务器错误"
// @Router /workflow/cancel/{id} [post]
func (s *WorkflowService) Cancel(c *gin.Context) {
	instanceID, err := parseID(c)
	if err != nil {
		response.Err(c, http.StatusBadRequest, err.Error())
		return
	}

	_, err = s.q.UpdateTaskWorkflowInstanceStatus(c, repo.UpdateTaskWorkflowInstanceStatusParams{
		Status: repo.TaskWorkflowInstanceStatusCANCELLED,
		GmtEnd: sql.NullTime{Time: time.Now(), Valid: true},
		ID:     instanceID,
	})
	if err != nil {
		response.Err(c, http.StatusInternalServerError, err.Error())
		return
	}
	response.Ok(c, nil)
}

// RollbackWorkflow godoc
// @Summary 回滚工作流
// @Description 对已完成的工作流执行回滚操作，逆序执行所有可回滚节点的回滚脚本
// @Tags 工作流执行
// @Produce json
// @Param id path int true "工作流实例ID"
// @Success 200 {object} response.Response "成功"
// @Failure 400 {object} response.Response "请求参数错误"
// @Failure 500 {object} response.Response "服务器错误"
// @Router /workflow/rollback/{id} [post]
func (s *WorkflowService) Rollback(c *gin.Context) {
	instanceID, err := parseID(c)
	if err != nil {
		response.Err(c, http.StatusBadRequest, err.Error())
		return
	}

	_, err = s.q.UpdateTaskWorkflowInstanceStatus(c, repo.UpdateTaskWorkflowInstanceStatusParams{
		Status: repo.TaskWorkflowInstanceStatusROLLINGBACK,
		GmtEnd: sql.NullTime{},
		ID:     instanceID,
	})
	if err != nil {
		response.Err(c, http.StatusInternalServerError, err.Error())
		return
	}
	response.Ok(c, map[string]interface{}{
		"message": "rollback started",
	})
}

// CountWorkflowInstances godoc
// @Summary 获取工作流实例总数
// @Description 统计系统中所有工作流实例数量
// @Tags 工作流实例
// @Produce json
// @Success 200 {object} response.Response "成功"
// @Failure 500 {object} response.Response "服务器错误"
// @Router /workflow_instance/count [get]
func (s *WorkflowService) CountInstance(c *gin.Context) {
	cnt, err := s.q.CountTaskWorkflowInstances(c)
	if err != nil {
		response.Err(c, http.StatusInternalServerError, err.Error())
		return
	}
	response.Ok(c, cnt)
}

// GetWorkflowInstanceByID godoc
// @Summary 获取工作流实例详情
// @Description 根据 ID 获取工作流实例详细信息
// @Tags 工作流实例
// @Produce json
// @Param id path int true "工作流实例ID"
// @Success 200 {object} response.Response "成功"
// @Failure 400 {object} response.Response "请求参数错误"
// @Failure 404 {object} response.Response "实例不存在"
// @Failure 500 {object} response.Response "服务器错误"
// @Router /workflow_instance/get/{id} [get]
func (s *WorkflowService) GetInstanceByID(c *gin.Context) {
	id, err := parseID(c)
	if err != nil {
		response.Err(c, http.StatusBadRequest, err.Error())
		return
	}

	instance, err := s.q.GetTaskWorkflowInstanceByID(c, id)
	if err != nil {
		if err == sql.ErrNoRows {
			response.Err(c, http.StatusNotFound, "instance not found")
			return
		}
		response.Err(c, http.StatusInternalServerError, err.Error())
		return
	}
	response.Ok(c, instance)
}

// ListWorkflowInstances godoc
// @Summary 获取工作流实例列表
// @Description 分页查询工作流实例列表
// @Tags 工作流实例
// @Accept json
// @Produce json
// @Param limit query int false "每页数量 (默认20, 最大100)" default(20)
// @Param offset query int false "偏移量" default(0)
// @Success 200 {object} response.Response "成功"
// @Failure 500 {object} response.Response "服务器错误"
// @Router /workflow_instance/list [get]
func (s *WorkflowService) ListInstances(c *gin.Context) {
	limit, offset := getPagination(c)
	instances, err := s.q.ListTaskWorkflowInstances(c, repo.ListTaskWorkflowInstancesParams{Limit: int32(limit), Offset: int32(offset)})
	if err != nil {
		response.Err(c, http.StatusInternalServerError, err.Error())
		return
	}
	response.Ok(c, instances)
}

// GetWorkflowExecutionLogs godoc
// @Summary 获取工作流执行日志
// @Description 根据工作流实例 ID 获取完整的执行日志
// @Tags 工作流实例
// @Produce json
// @Param id path int true "工作流实例ID"
// @Success 200 {object} response.Response "成功"
// @Failure 400 {object} response.Response "请求参数错误"
// @Failure 500 {object} response.Response "服务器错误"
// @Router /workflow_instance/get/{id}/logs [get]
func (s *WorkflowService) GetExecutionLogs(c *gin.Context) {
	id, err := parseID(c)
	if err != nil {
		response.Err(c, http.StatusBadRequest, err.Error())
		return
	}

	logs, err := s.q.ListTaskExecutionLogs(c, id)
	if err != nil {
		response.Err(c, http.StatusInternalServerError, err.Error())
		return
	}
	response.Ok(c, logs)
}

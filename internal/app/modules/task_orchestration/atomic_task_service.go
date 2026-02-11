package task_orchestration

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"

	"assistant/internal/app/repo"
	"assistant/internal/bootstrap/psl"
	"assistant/pkg/response"

	"github.com/gin-gonic/gin"
)

type TaskAtomicService struct {
	q *repo.Queries
}

func NewTaskAtomicService(q *repo.Queries) *TaskAtomicService {
	return &TaskAtomicService{q: q}
}

type CreateTaskAtomicRequest struct {
	Name                  string          `json:"name" binding:"required"`
	Description           string          `json:"description"`
	TaskCategory          string          `json:"task_category" binding:"required"`
	ScriptType            string          `json:"script_type"`
	ScriptContent         string          `json:"script_content" binding:"required"`
	RollbackScriptType    string          `json:"rollback_script_type"`
	RollbackScriptContent string          `json:"rollback_script_content"`
	HTTPConfig            json.RawMessage `json:"http_config"`
	BuiltinConfig         json.RawMessage `json:"builtin_config"`
	Timeout               int             `json:"timeout"`
	RetryCount            int             `json:"retry_count"`
	RetryInterval         int             `json:"retry_interval"`
	IsRollbackSupported   bool            `json:"is_rollback_supported"`
	Parameters            json.RawMessage `json:"parameters"`
	OutputSchema          json.RawMessage `json:"output_schema"`
	EnvVars               json.RawMessage `json:"env_vars"`
	WorkingDir            string          `json:"working_dir"`
}

// CountTaskAtomics godoc
// @Summary 获取原子任务总数
// @Description 统计系统中所有启用的原子任务数量
// @Tags 原子任务管理
// @Produce json
// @Success 200 {object} response.Response "成功"
// @Failure 500 {object} response.Response "服务器错误"
// @Router /task_atomic/count [get]
func (s *TaskAtomicService) Count(c *gin.Context) {
	cnt, err := s.q.CountTaskAtomicDefs(c)
	if err != nil {
		response.Err(c, http.StatusInternalServerError, err.Error())
		return
	}
	response.Ok(c, cnt)
}

// CreateTaskAtomic godoc
// @Summary 创建原子任务
// @Description 创建一个新的原子任务，支持 Shell/Python/Lua 脚本或 HTTP API 调用
// @Tags 原子任务管理
// @Accept json
// @Produce json
// @Param data body CreateTaskAtomicRequest true "原子任务创建参数"
// @Success 200 {object} response.Response "成功"
// @Failure 400 {object} response.Response "请求参数错误"
// @Failure 500 {object} response.Response "服务器错误"
// @Router /task_atomic/create [post]
func (s *TaskAtomicService) Create(c *gin.Context) {
	logger := psl.GetLogger()

	var req CreateTaskAtomicRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Err(c, http.StatusBadRequest, err.Error())
		return
	}

	params := repo.CreateTaskAtomicDefParams{
		Name:                req.Name,
		Description:         sql.NullString{String: req.Description, Valid: req.Description != ""},
		TaskCategory:        repo.TaskAtomicDefTaskCategory(req.TaskCategory),
		ScriptContent:       req.ScriptContent,
		IsRollbackSupported: sql.NullBool{Bool: req.IsRollbackSupported, Valid: true},
		Parameters:          req.Parameters,
		OutputSchema:        req.OutputSchema,
		EnvVars:             req.EnvVars,
		Status:              repo.NullTaskAtomicDefStatus{TaskAtomicDefStatus: "ENABLED", Valid: true},
	}

	if req.ScriptType != "" {
		params.ScriptType = repo.NullTaskAtomicDefScriptType{TaskAtomicDefScriptType: repo.TaskAtomicDefScriptType(req.ScriptType), Valid: true}
	}
	if req.RollbackScriptContent != "" {
		params.RollbackScriptType = repo.NullTaskAtomicDefRollbackScriptType{TaskAtomicDefRollbackScriptType: repo.TaskAtomicDefRollbackScriptType(req.RollbackScriptType), Valid: true}
		params.RollbackScriptContent = sql.NullString{String: req.RollbackScriptContent, Valid: true}
	}

	result, err := s.q.CreateTaskAtomicDef(c, params)
	if err != nil {
		logger.WithFields(map[string]interface{}{
			"name":          req.Name,
			"task_category": req.TaskCategory,
			"script_type":   req.ScriptType,
		}).WithError(err).Warn("create atomic task failed")
		response.Err(c, http.StatusInternalServerError, err.Error())
		return
	}

	id, _ := result.LastInsertId()
	logger.WithFields(map[string]interface{}{
		"task_atomic_def_id": id,
		"name":               req.Name,
		"task_category":      req.TaskCategory,
		"script_type":        req.ScriptType,
		"rollback_supported": req.IsRollbackSupported,
	}).Info("atomic task created")
	response.Ok(c, id)
}

// GetTaskAtomicByID godoc
// @Summary 获取原子任务详情
// @Description 根据 ID 获取原子任务详细信息
// @Tags 原子任务管理
// @Produce json
// @Param id path int true "原子任务ID"
// @Success 200 {object} response.Response "成功"
// @Failure 400 {object} response.Response "请求参数错误"
// @Failure 404 {object} response.Response "任务不存在"
// @Failure 500 {object} response.Response "服务器错误"
// @Router /task_atomic/get/{id} [get]
func (s *TaskAtomicService) GetByID(c *gin.Context) {
	id, err := parseID(c)
	if err != nil {
		response.Err(c, http.StatusBadRequest, err.Error())
		return
	}

	task, err := s.q.GetTaskAtomicDefByID(c, id)
	if err != nil {
		if err == sql.ErrNoRows {
			response.Err(c, http.StatusNotFound, "atomic task not found")
			return
		}
		response.Err(c, http.StatusInternalServerError, err.Error())
		return
	}
	response.Ok(c, task)
}

// ListTaskAtomics godoc
// @Summary 获取原子任务列表
// @Description 分页查询原子任务列表
// @Tags 原子任务管理
// @Accept json
// @Produce json
// @Param limit query int false "每页数量 (默认20, 最大100)" default(20)
// @Param offset query int false "偏移量" default(0)
// @Success 200 {object} response.Response "成功"
// @Failure 500 {object} response.Response "服务器错误"
// @Router /task_atomic/list [get]
func (s *TaskAtomicService) List(c *gin.Context) {
	limit, offset := getPagination(c)
	tasks, err := s.q.ListTaskAtomicDefs(c, repo.ListTaskAtomicDefsParams{Limit: int32(limit), Offset: int32(offset)})
	if err != nil {
		response.Err(c, http.StatusInternalServerError, err.Error())
		return
	}
	response.Ok(c, tasks)
}

// UpdateTaskAtomic godoc
// @Summary 更新原子任务
// @Description 根据 ID 更新原子任务信息
// @Tags 原子任务管理
// @Accept json
// @Produce json
// @Param id path int true "原子任务ID"
// @Param data body CreateTaskAtomicRequest true "更新参数"
// @Success 200 {object} response.Response "成功"
// @Failure 400 {object} response.Response "请求参数错误"
// @Failure 500 {object} response.Response "服务器错误"
// @Router /task_atomic/update/{id} [put]
func (s *TaskAtomicService) Update(c *gin.Context) {
	logger := psl.GetLogger()

	id, err := parseID(c)
	if err != nil {
		response.Err(c, http.StatusBadRequest, err.Error())
		return
	}

	var req CreateTaskAtomicRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Err(c, http.StatusBadRequest, err.Error())
		return
	}

	params := repo.UpdateTaskAtomicDefByIDParams{
		Name:                req.Name,
		Description:         sql.NullString{String: req.Description, Valid: req.Description != ""},
		TaskCategory:        repo.TaskAtomicDefTaskCategory(req.TaskCategory),
		ScriptContent:       req.ScriptContent,
		IsRollbackSupported: sql.NullBool{Bool: req.IsRollbackSupported, Valid: true},
		Parameters:          req.Parameters,
		OutputSchema:        req.OutputSchema,
		EnvVars:             req.EnvVars,
		Status:              repo.NullTaskAtomicDefStatus{TaskAtomicDefStatus: "ENABLED", Valid: true},
		ID:                  id,
	}

	if req.ScriptType != "" {
		params.ScriptType = repo.NullTaskAtomicDefScriptType{TaskAtomicDefScriptType: repo.TaskAtomicDefScriptType(req.ScriptType), Valid: true}
	}
	if req.RollbackScriptContent != "" {
		params.RollbackScriptType = repo.NullTaskAtomicDefRollbackScriptType{TaskAtomicDefRollbackScriptType: repo.TaskAtomicDefRollbackScriptType(req.RollbackScriptType), Valid: true}
		params.RollbackScriptContent = sql.NullString{String: req.RollbackScriptContent, Valid: true}
	}
	if req.HTTPConfig != nil {
		params.HttpConfig = req.HTTPConfig
	}
	if req.BuiltinConfig != nil {
		params.BuiltinConfig = req.BuiltinConfig
	}
	if req.Timeout > 0 {
		params.Timeout = sql.NullInt32{Int32: int32(req.Timeout), Valid: true}
	}
	if req.RetryCount > 0 {
		params.RetryCount = sql.NullInt32{Int32: int32(req.RetryCount), Valid: true}
	}
	if req.RetryInterval > 0 {
		params.RetryInterval = sql.NullInt32{Int32: int32(req.RetryInterval), Valid: true}
	}
	if req.WorkingDir != "" {
		params.WorkingDir = sql.NullString{String: req.WorkingDir, Valid: true}
	}

	_, err = s.q.UpdateTaskAtomicDefByID(c, params)
	if err != nil {
		logger.WithFields(map[string]interface{}{
			"task_atomic_def_id": id,
			"name":               req.Name,
			"task_category":      req.TaskCategory,
			"script_type":        req.ScriptType,
		}).WithError(err).Warn("update atomic task failed")
		response.Err(c, http.StatusInternalServerError, err.Error())
		return
	}
	logger.WithFields(map[string]interface{}{
		"task_atomic_def_id": id,
		"name":               req.Name,
		"task_category":      req.TaskCategory,
		"script_type":        req.ScriptType,
		"rollback_supported": req.IsRollbackSupported,
	}).Info("atomic task updated")
	response.Ok(c, nil)
}

// DeleteTaskAtomic godoc
// @Summary 删除原子任务
// @Description 根据 ID 删除原子任务
// @Tags 原子任务管理
// @Produce json
// @Param id path int true "原子任务ID"
// @Success 200 {object} response.Response "成功"
// @Failure 400 {object} response.Response "请求参数错误"
// @Failure 500 {object} response.Response "服务器错误"
// @Router /task_atomic/delete/{id} [delete]
func (s *TaskAtomicService) Delete(c *gin.Context) {
	logger := psl.GetLogger()

	id, err := parseID(c)
	if err != nil {
		response.Err(c, http.StatusBadRequest, err.Error())
		return
	}

	_, err = s.q.DeleteTaskAtomicDefByID(c, id)
	if err != nil {
		logger.WithField("task_atomic_def_id", id).WithError(err).Warn("delete atomic task failed")
		response.Err(c, http.StatusInternalServerError, err.Error())
		return
	}
	logger.WithField("task_atomic_def_id", id).Info("atomic task deleted")
	response.Ok(c, nil)
}

func parseID(c *gin.Context) (int64, error) {
	var id int64
	fmt.Sscanf(c.Param("id"), "%d", &id)
	return id, nil
}

func getPagination(c *gin.Context) (int, int) {
	limit := 20
	offset := 0
	fmt.Sscanf(c.Query("limit"), "%d", &limit)
	fmt.Sscanf(c.Query("offset"), "%d", &offset)
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}
	return limit, offset
}

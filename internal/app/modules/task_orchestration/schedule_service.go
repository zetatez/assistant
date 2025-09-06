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

type ScheduleService struct {
	q *repo.Queries
}

func NewScheduleService(q *repo.Queries) *ScheduleService {
	return &ScheduleService{q: q}
}

type CreateScheduleRequest struct {
	Name            string          `json:"name" binding:"required"`
	Description     string          `json:"description"`
	WorkflowDefID   int64           `json:"workflow_def_id" binding:"required"`
	ScheduleType    string          `json:"schedule_type" binding:"required"`
	CronExpr        string          `json:"cron_expr"`
	IntervalSeconds int             `json:"interval_seconds"`
	ExecuteAt       string          `json:"execute_at"`
	InputParams     json.RawMessage `json:"input_params"`
}

func (s *ScheduleService) Count(c *gin.Context) {
	// CountSchedules godoc
	// @Summary 获取任务调度总数
	// @Description 统计系统中所有启用的任务调度数量
	// @Tags 任务调度
	// @Produce json
	// @Success 200 {object} response.Response "成功"
	// @Failure 500 {object} response.Response "服务器错误"
	// @Router /schedule/count [get]
	cnt, err := s.q.CountTaskSchedules(c)
	if err != nil {
		response.Err(c, http.StatusInternalServerError, err.Error())
		return
	}
	response.Ok(c, cnt)
}

// CreateSchedule godoc
// @Summary 创建任务调度
// @Description 创建一个新的任务调度，支持 Cron/Interval/Once 三种类型
// @Tags 任务调度
// @Accept json
// @Produce json
// @Param data body CreateScheduleRequest true "调度创建参数"
// @Success 200 {object} response.Response "成功"
// @Failure 400 {object} response.Response "请求参数错误"
// @Failure 500 {object} response.Response "服务器错误"
// @Router /schedule/create [post]
func (s *ScheduleService) Create(c *gin.Context) {
	var req CreateScheduleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Err(c, http.StatusBadRequest, err.Error())
		return
	}

	params := repo.CreateTaskScheduleParams{
		Name:          req.Name,
		Description:   sql.NullString{String: req.Description, Valid: req.Description != ""},
		WorkflowDefID: req.WorkflowDefID,
		ScheduleType:  repo.TaskScheduleScheduleType(req.ScheduleType),
		InputParams:   req.InputParams,
		Status:        repo.NullTaskScheduleStatus{TaskScheduleStatus: "ENABLED", Valid: true},
	}

	if req.CronExpr != "" {
		params.CronExpr = sql.NullString{String: req.CronExpr, Valid: true}
	}
	if req.IntervalSeconds > 0 {
		params.IntervalSeconds = sql.NullInt32{Int32: int32(req.IntervalSeconds), Valid: true}
	}
	if req.ExecuteAt != "" {
		t, _ := time.Parse("2006-01-02 15:04:05", req.ExecuteAt)
		params.ExecuteAt = sql.NullTime{Time: t, Valid: true}
	}

	result, err := s.q.CreateTaskSchedule(c, params)
	if err != nil {
		response.Err(c, http.StatusInternalServerError, err.Error())
		return
	}

	id, _ := result.LastInsertId()
	response.Ok(c, id)
}

// GetScheduleByID godoc
// @Summary 获取任务调度详情
// @Description 根据 ID 获取任务调度详细信息
// @Tags 任务调度
// @Produce json
// @Param id path int true "调度ID"
// @Success 200 {object} response.Response "成功"
// @Failure 400 {object} response.Response "请求参数错误"
// @Failure 404 {object} response.Response "调度不存在"
// @Failure 500 {object} response.Response "服务器错误"
// @Router /schedule/get/{id} [get]
func (s *ScheduleService) GetByID(c *gin.Context) {
	id, err := parseID(c)
	if err != nil {
		response.Err(c, http.StatusBadRequest, err.Error())
		return
	}

	schedule, err := s.q.GetTaskScheduleByID(c, id)
	if err != nil {
		if err == sql.ErrNoRows {
			response.Err(c, http.StatusNotFound, "schedule not found")
			return
		}
		response.Err(c, http.StatusInternalServerError, err.Error())
		return
	}
	response.Ok(c, schedule)
}

// ListSchedules godoc
// @Summary 获取任务调度列表
// @Description 分页查询任务调度列表
// @Tags 任务调度
// @Accept json
// @Produce json
// @Param limit query int false "每页数量 (默认20, 最大100)" default(20)
// @Param offset query int false "偏移量" default(0)
// @Success 200 {object} response.Response "成功"
// @Failure 500 {object} response.Response "服务器错误"
// @Router /schedule/list [get]
func (s *ScheduleService) List(c *gin.Context) {
	limit, offset := getPagination(c)
	schedules, err := s.q.ListTaskSchedules(c, repo.ListTaskSchedulesParams{Limit: int32(limit), Offset: int32(offset)})
	if err != nil {
		response.Err(c, http.StatusInternalServerError, err.Error())
		return
	}
	response.Ok(c, schedules)
}

// UpdateSchedule godoc
// @Summary 更新任务调度
// @Description 根据 ID 更新任务调度信息
// @Tags 任务调度
// @Accept json
// @Produce json
// @Param id path int true "调度ID"
// @Param data body CreateScheduleRequest true "更新参数"
// @Success 200 {object} response.Response "成功"
// @Failure 400 {object} response.Response "请求参数错误"
// @Failure 500 {object} response.Response "服务器错误"
// @Router /schedule/update/{id} [put]
func (s *ScheduleService) Update(c *gin.Context) {
	id, err := parseID(c)
	if err != nil {
		response.Err(c, http.StatusBadRequest, err.Error())
		return
	}

	var req CreateScheduleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Err(c, http.StatusBadRequest, err.Error())
		return
	}

	params := repo.UpdateTaskScheduleByIDParams{
		ID:           id,
		Name:         req.Name,
		Description:  sql.NullString{String: req.Description, Valid: req.Description != ""},
		ScheduleType: repo.TaskScheduleScheduleType(req.ScheduleType),
		InputParams:  req.InputParams,
	}

	if req.CronExpr != "" {
		params.CronExpr = sql.NullString{String: req.CronExpr, Valid: true}
	}
	if req.IntervalSeconds > 0 {
		params.IntervalSeconds = sql.NullInt32{Int32: int32(req.IntervalSeconds), Valid: true}
	}
	if req.ExecuteAt != "" {
		t, _ := time.Parse("2006-01-02 15:04:05", req.ExecuteAt)
		params.ExecuteAt = sql.NullTime{Time: t, Valid: true}
	}

	_, err = s.q.UpdateTaskScheduleByID(c, params)
	if err != nil {
		response.Err(c, http.StatusInternalServerError, err.Error())
		return
	}
	response.Ok(c, nil)
}

// DeleteSchedule godoc
// @Summary 删除任务调度
// @Description 根据 ID 删除任务调度
// @Tags 任务调度
// @Produce json
// @Param id path int true "调度ID"
// @Success 200 {object} response.Response "成功"
// @Failure 400 {object} response.Response "请求参数错误"
// @Failure 500 {object} response.Response "服务器错误"
// @Router /schedule/delete/{id} [delete]
func (s *ScheduleService) Delete(c *gin.Context) {
	id, err := parseID(c)
	if err != nil {
		response.Err(c, http.StatusBadRequest, err.Error())
		return
	}

	_, err = s.q.DeleteTaskScheduleByID(c, id)
	if err != nil {
		response.Err(c, http.StatusInternalServerError, err.Error())
		return
	}
	response.Ok(c, nil)
}

// EnableSchedule godoc
// @Summary 启用任务调度
// @Description 根据 ID 启用已禁用的任务调度
// @Tags 任务调度
// @Produce json
// @Param id path int true "调度ID"
// @Success 200 {object} response.Response "成功"
// @Failure 400 {object} response.Response "请求参数错误"
// @Failure 500 {object} response.Response "服务器错误"
// @Router /schedule/enable/{id} [post]
func (s *ScheduleService) Enable(c *gin.Context) {
	id, err := parseID(c)
	if err != nil {
		response.Err(c, http.StatusBadRequest, err.Error())
		return
	}

	_, err = s.q.UpdateTaskScheduleByID(c, repo.UpdateTaskScheduleByIDParams{
		Status: repo.NullTaskScheduleStatus{TaskScheduleStatus: "ENABLED", Valid: true},
		ID:     id,
	})
	if err != nil {
		response.Err(c, http.StatusInternalServerError, err.Error())
		return
	}
	response.Ok(c, nil)
}

// DisableSchedule godoc
// @Summary 禁用任务调度
// @Description 根据 ID 禁用任务调度
// @Tags 任务调度
// @Produce json
// @Param id path int true "调度ID"
// @Success 200 {object} response.Response "成功"
// @Failure 400 {object} response.Response "请求参数错误"
// @Failure 500 {object} response.Response "服务器错误"
// @Router /schedule/disable/{id} [post]
// DisableSchedule godoc
// @Summary 禁用任务调度
// @Description 根据 ID 禁用任务调度
// @Tags 任务调度
// @Produce json
// @Param id path int true "调度ID"
// @Success 200 {object} response.Response "成功"
// @Failure 400 {object} response.Response "请求参数错误"
// @Failure 500 {object} response.Response "服务器错误"
// @Router /schedule/disable/{id} [post]
func (s *ScheduleService) Disable(c *gin.Context) {
	id, err := parseID(c)
	if err != nil {
		response.Err(c, http.StatusBadRequest, err.Error())
		return
	}

	_, err = s.q.UpdateTaskScheduleByID(c, repo.UpdateTaskScheduleByIDParams{
		Status: repo.NullTaskScheduleStatus{TaskScheduleStatus: "DISABLED", Valid: true},
		ID:     id,
	})
	if err != nil {
		response.Err(c, http.StatusInternalServerError, err.Error())
		return
	}
	response.Ok(c, nil)
}

package sys_distributed_locker

import (
	"assistant/internal/bootstrap/psl"
	"assistant/pkg/response"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type SysDistributedLockHandler struct {
	svc *SysDistributedLockService
}

func NewSysDistributedLockHandler(svc *SysDistributedLockService) *SysDistributedLockHandler {
	return &SysDistributedLockHandler{svc: svc}
}

func (h *SysDistributedLockHandler) Register(r *gin.RouterGroup) {
	r.POST("/acquire", h.Acquire)
	r.POST("/release", h.Release)
	r.POST("/renew", h.Renew)
	r.GET("/query/:lock_key", h.Query)
	r.GET("/check/:lock_key", h.Check)
}

// Acquire godoc
// @Summary 尝试获取分布式锁
// @Description 原子操作：如果锁不存在或已过期则获取成功
// @Tags 分布式锁
// @Accept json
// @Produce json
// @Param data body acquireReq true "锁参数"
// @Success 200 {object} response.Response "成功"
// @Failure 400 {object} response.Response "请求参数错误"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Router /sys_distributed_locker/acquire [post]
func (h *SysDistributedLockHandler) Acquire(c *gin.Context) {
	var req acquireReq
	if err := c.BindJSON(&req); err != nil {
		response.Err(c, response.CodeInvalidParams, "invalid request parameters")
		return
	}
	if req.LockKey == "" {
		response.Err(c, response.CodeInvalidParams, "lock_key is required")
		return
	}
	if req.LockHolder == "" {
		req.LockHolder = uuid.New().String()
	}
	psl.GetLogger().Infof("[distributed_locker] acquire request: key=%s ttl=%d", req.LockKey, req.TTL)
	acquired, err := h.svc.TryAcquire(c, req.LockKey, req.LockHolder, req.TTL)
	if err != nil {
		response.ErrWithInternal(c, response.CodeServerError, "failed to acquire lock", err)
		return
	}
	response.Ok(c, gin.H{
		"acquired":    acquired,
		"lock_holder": req.LockHolder,
	})
}

// Release godoc
// @Summary 释放分布式锁
// @Description 释放锁，必须提供 lock_holder 验证所有权
// @Tags 分布式锁
// @Accept json
// @Produce json
// @Param data body releaseReq true "释放参数"
// @Success 200 {object} response.Response "成功"
// @Failure 400 {object} response.Response "请求参数错误"
// @Failure 403 {object} response.Response "无权限（holder不匹配）"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Router /sys_distributed_locker/release [post]
func (h *SysDistributedLockHandler) Release(c *gin.Context) {
	var req releaseReq
	if err := c.BindJSON(&req); err != nil {
		response.Err(c, response.CodeInvalidParams, "invalid request parameters")
		return
	}
	if req.LockKey == "" || req.LockHolder == "" {
		response.Err(c, response.CodeInvalidParams, "lock_key and lock_holder are required")
		return
	}
	psl.GetLogger().Infof("[distributed_locker] release request: key=%s holder=%s", req.LockKey, req.LockHolder[:8])
	released, err := h.svc.Release(c, req.LockKey, req.LockHolder)
	if err != nil {
		response.ErrWithInternal(c, response.CodeServerError, "failed to release lock", err)
		return
	}
	if !released {
		response.Err(c, response.CodeForbidden, "lock not found or not owned by this holder")
		return
	}
	response.Ok(c, true)
}

// Renew godoc
// @Summary 续期分布式锁
// @Description 为锁延长TTL，必须验证 lock_holder 是否匹配
// @Tags 分布式锁
// @Accept json
// @Produce json
// @Param data body renewReq true "续期参数"
// @Success 200 {object} response.Response "成功"
// @Failure 400 {object} response.Response "请求参数错误"
// @Failure 403 {object} response.Response "无权限（holder不匹配或锁已过期）"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Router /sys_distributed_locker/renew [post]
func (h *SysDistributedLockHandler) Renew(c *gin.Context) {
	var req renewReq
	if err := c.BindJSON(&req); err != nil {
		response.Err(c, response.CodeInvalidParams, "invalid request parameters")
		return
	}
	if req.LockKey == "" || req.LockHolder == "" {
		response.Err(c, response.CodeInvalidParams, "lock_key and lock_holder are required")
		return
	}
	psl.GetLogger().Infof("[distributed_locker] renew request: key=%s ttl=%d", req.LockKey, req.TTL)
	renewed, err := h.svc.Renew(c, req.LockKey, req.LockHolder, req.TTL)
	if err != nil {
		response.ErrWithInternal(c, response.CodeServerError, "failed to renew lock", err)
		return
	}
	if !renewed {
		response.Err(c, response.CodeForbidden, "lock not found, not owned by this holder, or already expired")
		return
	}
	response.Ok(c, true)
}

// Query godoc
// @Summary 查询锁信息
// @Description 获取锁的详细信息（包含holder、TTL等）
// @Tags 分布式锁
// @Produce json
// @Param lock_key path string true "锁Key"
// @Success 200 {object} response.Response "成功"
// @Failure 400 {object} response.Response "请求参数错误"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Router /sys_distributed_locker/query/{lock_key} [get]
func (h *SysDistributedLockHandler) Query(c *gin.Context) {
	key := c.Param("lock_key")
	if key == "" {
		response.Err(c, response.CodeInvalidParams, "lock_key is required")
		return
	}
	psl.GetLogger().Infof("[distributed_locker] query request: key=%s", key)
	lock, err := h.svc.Get(c, key)
	if err != nil {
		response.ErrWithInternal(c, response.CodeServerError, "failed to query lock", err)
		return
	}
	response.Ok(c, lock)
}

// Check godoc
// @Summary 检查锁是否被持有
// @Description 返回锁是否被持有（未过期）以及holder
// @Tags 分布式锁
// @Produce json
// @Param lock_key path string true "锁Key"
// @Success 200 {object} response.Response "成功"
// @Failure 400 {object} response.Response "请求参数错误"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Router /sys_distributed_locker/check/{lock_key} [get]
func (h *SysDistributedLockHandler) Check(c *gin.Context) {
	key := c.Param("lock_key")
	if key == "" {
		response.Err(c, response.CodeInvalidParams, "lock_key is required")
		return
	}
	psl.GetLogger().Infof("[distributed_locker] check request: key=%s", key)
	held, holder, err := h.svc.IsHeld(c, key)
	if err != nil {
		response.ErrWithInternal(c, response.CodeServerError, "failed to check lock", err)
		return
	}
	response.Ok(c, gin.H{
		"held":        held,
		"lock_holder": holder,
	})
}

type acquireReq struct {
	LockKey    string `json:"lock_key"`
	LockHolder string `json:"lock_holder"`
	TTL        int    `json:"ttl"`
}

type releaseReq struct {
	LockKey    string `json:"lock_key"`
	LockHolder string `json:"lock_holder"`
}

type renewReq struct {
	LockKey    string `json:"lock_key"`
	LockHolder string `json:"lock_holder"`
	TTL        int    `json:"ttl"`
}

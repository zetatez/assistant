package sys_user

import (
	"assistant/internal/app/repo"
	"assistant/internal/bootstrap/psl"
	"assistant/pkg/middleware"
	"assistant/pkg/response"
	"database/sql"
	"errors"
	"strconv"

	"github.com/gin-gonic/gin"
)

type SysUserHandler struct {
	svc *SysUserService
}

func NewSysUserHandler(svc *SysUserService) *SysUserHandler {
	return &SysUserHandler{svc: svc}
}

func (h *SysUserHandler) Register(r *gin.RouterGroup) {
	r.GET("/count", h.CountSysUser)
	r.POST("/create", h.CreateSysUser)
	r.DELETE("/delete/:id", h.DeleteSysUserByID)
	r.GET("/get/:id", h.GetSysUserByID)
	r.GET("/list", h.ListSysUsers)
	r.POST("/search_by_email", h.SearchSysUsersByEmail)
	r.POST("/search_by_user_name", h.SearchSysUsersByUserName)
	r.PUT("/update/:id", h.UpdateSysUserByID)
}

func (h *SysUserHandler) RegisterWithoutAuth(r *gin.RouterGroup) {
	h.Register(r)
}

func (h *SysUserHandler) DebugUser(c *gin.Context) {
	psl.GetLogger().Infof("[sys_user] debug list users request")
	users, err := h.svc.DebugListUsers(c)
	if err != nil {
		response.ErrWithInternal(c, response.CodeServerError, "failed to list users", err)
		return
	}
	response.Ok(c, users)
}

type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type LoginResponse struct {
	Token    string `json:"token"`
	UserID   int64  `json:"user_id"`
	Username string `json:"username"`
}

func (h *SysUserHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.BindJSON(&req); err != nil {
		response.Err(c, response.CodeInvalidParams, "invalid request parameters")
		return
	}
	psl.GetLogger().Infof("[sys_user] login request: username=%s", req.Username)
	user, err := h.svc.Login(c, req.Username, req.Password)
	if err != nil {
		if errors.Is(err, ErrInvalidCredentials) {
			response.Err(c, response.CodeUnauthorized, "invalid username or password")
			return
		}
		response.ErrWithInternal(c, response.CodeServerError, "login failed", err)
		return
	}
	token, err := middleware.GenerateToken(psl.GetConfig().App.JWT.Secret, user.ID, user.UserName, psl.GetConfig().App.JWT.Expiry)
	if err != nil {
		response.ErrWithInternal(c, response.CodeServerError, "failed to generate token", err)
		return
	}
	response.Ok(c, LoginResponse{
		Token:    token,
		UserID:   user.ID,
		Username: user.UserName,
	})
}

// CountSysUser godoc
// @Summary 统计用户数量
// @Description 获取系统中所有用户的数量
// @Tags 用户管理
// @Produce json
// @Success 200 {object} response.Response "成功"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Router /sys_user/count [get]
func (h *SysUserHandler) CountSysUser(c *gin.Context) {
	psl.GetLogger().Infof("[sys_user] count request")
	data, err := h.svc.CountSysUsers(c)
	if err != nil {
		response.ErrWithInternal(c, response.CodeServerError, "failed to count users", err)
		return
	}
	response.Ok(c, data)
}

// CreateSysUser godoc
// @Summary 创建用户
// @Description 创建一个新的用户
// @Tags 用户管理
// @Accept json
// @Produce json
// @Param data body repo.CreateSysUserParams true "用户创建参数"
// @Success 200 {object} response.Response "成功"
// @Failure 400 {object} response.Response "请求参数错误"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Router /sys_user/create [post]
func (h *SysUserHandler) CreateSysUser(c *gin.Context) {
	var req repo.CreateSysUserParams
	if err := c.BindJSON(&req); err != nil {
		response.Err(c, response.CodeInvalidParams, "invalid request parameters")
		return
	}
	psl.GetLogger().Infof("[sys_user] create request: username=%s", req.UserName)
	data, err := h.svc.CreateSysUser(c, req)
	if err != nil {
		response.ErrWithInternal(c, response.CodeServerError, "failed to create user", err)
		return
	}
	response.Ok(c, data)
}

// DeleteSysUserByID godoc
// @Summary 删除用户
// @Description 根据用户ID删除用户
// @Tags 用户管理
// @Produce json
// @Param id path int true "用户ID"
// @Success 200 {object} response.Response "成功"
// @Failure 400 {object} response.Response "请求参数错误"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Router /sys_user/delete/{id} [delete]
func (h *SysUserHandler) DeleteSysUserByID(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Err(c, response.CodeInvalidParams, "invalid user id")
		return
	}
	if u, err := h.svc.GetSysUserByID(c, id); err == nil {
		if u.IsInternal == 1 {
			response.Err(c, response.CodeForbidden, "cannot delete internal default user")
			return
		}
	} else if err != sql.ErrNoRows {
		response.ErrWithInternal(c, response.CodeServerError, "failed to get user", err)
		return
	}
	data, err := h.svc.DeleteSysUserByID(c, id)
	if err != nil {
		response.ErrWithInternal(c, response.CodeServerError, "failed to delete user", err)
		return
	}
	response.Ok(c, data)
}

// GetSysUserByID godoc
// @Summary 获取用户详情
// @Description 根据用户ID获取用户信息
// @Tags 用户管理
// @Produce json
// @Param id path int true "用户ID"
// @Success 200 {object} response.Response "成功"
// @Failure 400 {object} response.Response "请求参数错误"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Router /sys_user/get/{id} [get]
func (h *SysUserHandler) GetSysUserByID(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Err(c, response.CodeInvalidParams, "invalid user id")
		return
	}
	data, err := h.svc.GetSysUserByID(c, id)
	if err != nil {
		if err == sql.ErrNoRows {
			response.Err(c, response.CodeNotFound, "user not found")
			return
		}
		response.ErrWithInternal(c, response.CodeServerError, "failed to get user", err)
		return
	}
	response.Ok(c, data)
}

// ListSysUsers godoc
// @Summary 获取用户列表
// @Description 支持分页、排序的用户列表查询
// @Tags 用户管理
// @Accept json
// @Produce json
// @Param data query repo.ListSysUsersParams true "用户列表参数"
func (h *SysUserHandler) ListSysUsers(c *gin.Context) {
	var req repo.ListSysUsersParams
	if err := c.ShouldBindQuery(&req); err != nil {
		response.Err(c, response.CodeInvalidParams, "invalid query parameters")
		return
	}
	page := c.Query("page")
	pageSize := c.Query("page_size")
	if page != "" {
		if p, err := strconv.Atoi(page); err == nil && p > 0 {
			req.Offset = int32((p - 1) * int(req.Limit))
		}
	}
	if pageSize != "" {
		if ps, err := strconv.Atoi(pageSize); err == nil && ps > 0 {
			req.Limit = int32(ps)
		}
	}
	if req.Limit <= 0 {
		req.Limit = 10
	}
	if req.Offset < 0 {
		req.Offset = 0
	}
	data, err := h.svc.ListSysUsers(c, req)
	if err != nil {
		response.ErrWithInternal(c, response.CodeServerError, "failed to list users", err)
		return
	}
	response.Ok(c, data)
}

// SearchSysUsersByEmail godoc
// @Summary 通过邮箱搜索用户
// @Description 根据邮箱地址模糊搜索用户
// @Tags 用户管理
// @Accept json
// @Produce json
// @Param data body repo.SearchSysUsersByEmailParams true "搜索参数"
// @Success 200 {object} response.Response "成功"
// @Failure 400 {object} response.Response "请求参数错误"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Router /sys_user/search_by_email [post]
func (h *SysUserHandler) SearchSysUsersByEmail(c *gin.Context) {
	var req repo.SearchSysUsersByEmailParams
	if err := c.BindJSON(&req); err != nil {
		response.Err(c, response.CodeInvalidParams, "invalid request parameters")
		return
	}
	req.Email = "%" + req.Email + "%"
	data, err := h.svc.SearchSysUsersByEmail(c, req)
	if err != nil {
		response.ErrWithInternal(c, response.CodeServerError, "failed to search users", err)
		return
	}
	response.Ok(c, data)
}

// SearchSysUsersByUserName godoc
// @Summary 通过用户名搜索用户
// @Description 根据用户名模糊搜索用户
// @Tags 用户管理
// @Accept json
// @Produce json
// @Param data body repo.SearchSysUsersByUserNameParams true "搜索参数"
// @Success 200 {object} response.Response "成功"
// @Failure 400 {object} response.Response "请求参数错误"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Router /sys_user/search_by_user_name [post]
func (h *SysUserHandler) SearchSysUsersByUserName(c *gin.Context) {
	var req repo.SearchSysUsersByUserNameParams
	if err := c.BindJSON(&req); err != nil {
		response.Err(c, response.CodeInvalidParams, "invalid request parameters")
		return
	}
	req.UserName = "%" + req.UserName + "%"
	data, err := h.svc.SearchSysUsersByUserName(c, req)
	if err != nil {
		response.ErrWithInternal(c, response.CodeServerError, "failed to search users", err)
		return
	}
	response.Ok(c, data)
}

// UpdateSysUserByID godoc
// @Summary 更新用户信息
// @Description 根据用户ID更新用户信息
// @Tags 用户管理
// @Accept json
// @Produce json
// @Param id path int true "用户ID"
// @Param data body repo.UpdateSysUserByIDParams true "用户更新参数"
// @Success 200 {object} response.Response "成功"
// @Failure 400 {object} response.Response "请求参数错误"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Router /sys_user/update/{id} [put]
func (h *SysUserHandler) UpdateSysUserByID(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Err(c, response.CodeInvalidParams, "invalid user id")
		return
	}
	var req repo.UpdateSysUserByIDParams
	req.ID = id
	if err = c.BindJSON(&req); err != nil {
		response.Err(c, response.CodeInvalidParams, "invalid request parameters")
		return
	}
	_, err = h.svc.UpdateSysUserByID(c, req)
	if err != nil {
		response.ErrWithInternal(c, response.CodeServerError, "failed to update user", err)
		return
	}
	response.Ok(c, nil)
}

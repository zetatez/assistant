package sys_user

import (
	"assistant/internal/app/repo"
	"assistant/pkg/response"
	"net/http"
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

// CountSysUser godoc
// @Summary 统计用户数量
// @Description 获取系统中所有用户的数量
// @Tags 用户管理
// @Produce json
// @Success 200 {object} response.Response "成功"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Router /sys_user/count [get]
func (h *SysUserHandler) CountSysUser(c *gin.Context) {
	data, err := h.svc.CountSysUsers(c)
	if err != nil {
		response.Err(c, http.StatusInternalServerError, err.Error())
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
		response.Err(c, http.StatusBadRequest, err.Error())
		return
	}
	data, err := h.svc.CreateSysUser(c, req)
	if err != nil {
		response.Err(c, http.StatusInternalServerError, err.Error())
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
		response.Err(c, http.StatusBadRequest, err.Error())
		return
	}
	data, err := h.svc.DeleteSysUserByID(c, id)
	if err != nil {
		response.Err(c, http.StatusInternalServerError, err.Error())
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
		response.Err(c, http.StatusBadRequest, err.Error())
		return
	}
	data, err := h.svc.GetSysUserByID(c, id)
	if err != nil {
		response.Err(c, http.StatusInternalServerError, err.Error())
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
// @Param data body repo.ListSysUsersParams true "用户列表参数"
// @Success 200 {object} response.Response "成功"
// @Failure 400 {object} response.Response "请求参数错误"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Router /sys_user/list [get]
func (h *SysUserHandler) ListSysUsers(c *gin.Context) {
	var req repo.ListSysUsersParams
	if err := c.BindJSON(&req); err != nil {
		response.Err(c, http.StatusBadRequest, err.Error())
		return
	}
	data, err := h.svc.ListSysUsers(c, req)
	if err != nil {
		response.Err(c, http.StatusInternalServerError, err.Error())
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
		response.Err(c, http.StatusBadRequest, err.Error())
		return
	}
	data, err := h.svc.SearchSysUsersByEmail(c, req)
	if err != nil {
		response.Err(c, http.StatusInternalServerError, err.Error())
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
		response.Err(c, http.StatusBadRequest, err.Error())
		return
	}
	data, err := h.svc.SearchSysUsersByUserName(c, req)
	if err != nil {
		response.Err(c, http.StatusInternalServerError, err.Error())
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
		response.Err(c, http.StatusBadRequest, err.Error())
		return
	}
	var req repo.UpdateSysUserByIDParams
	req.ID = id
	if err := c.BindJSON(&req); err != nil {
		response.Err(c, http.StatusBadRequest, err.Error())
		return
	}
	data, err := h.svc.UpdateSysUserByID(c, req)
	if err != nil {
		response.Err(c, http.StatusInternalServerError, err.Error())
		return
	}
	response.Ok(c, data)
}

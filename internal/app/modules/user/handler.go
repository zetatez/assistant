package user

import (
	"assistant/internal/app/repo"
	"assistant/pkg/response"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type UserHandler struct {
	svc *UserService
}

func NewUserHandler(svc *UserService) *UserHandler {
	return &UserHandler{svc: svc}
}

func (h *UserHandler) Register(r *gin.RouterGroup) {
	r.GET("/count", h.CountUser)
	r.POST("/create", h.CreateUser)
	r.DELETE("/delete/:id", h.DeleteUserByID)
	r.GET("/get/:id", h.GetUserByID)
	r.GET("/list", h.ListUsers)
	r.POST("/search_by_email", h.SearchUsersByEmail)
	r.POST("/search_by_user_name", h.SearchUsersByUserName)
	r.POST("/update", h.UpdateUserByID)
}

// CountUser godoc
// @Summary 统计用户数量
// @Description 获取系统中所有用户的数量
// @Tags 用户管理
// @Produce json
// @Success 200 {object} response.Response "成功"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Router /user/count [get]
func (h *UserHandler) CountUser(c *gin.Context) {
	data, err := h.svc.CountUsers(c)
	if err != nil {
		response.Err(c, http.StatusInternalServerError, err.Error())
		return
	}
	response.Ok(c, data)
}

// CreateUser godoc
// @Summary 创建用户
// @Description 创建一个新的用户
// @Tags 用户管理
// @Accept json
// @Produce json
// @Param data body repo.CreateUserParams true "用户创建参数"
// @Success 200 {object} response.Response "成功"
// @Failure 400 {object} response.Response "请求参数错误"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Router /user/create [post]
func (h *UserHandler) CreateUser(c *gin.Context) {
	var req repo.CreateUserParams
	if err := c.BindJSON(&req); err != nil {
		response.Err(c, http.StatusBadRequest, err.Error())
		return
	}
	err := h.svc.CreateUser(c, req)
	if err != nil {
		response.Err(c, http.StatusInternalServerError, err.Error())
		return
	}
	response.Ok(c, nil)
}

// DeleteUserByID godoc
// @Summary 删除用户
// @Description 根据用户ID删除用户
// @Tags 用户管理
// @Produce json
// @Param id path int true "用户ID"
// @Success 200 {object} response.Response "成功"
// @Failure 400 {object} response.Response "请求参数错误"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Router /user/delete/{id} [delete]
func (h *UserHandler) DeleteUserByID(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Err(c, http.StatusBadRequest, err.Error())
		return
	}
	err = h.svc.DeleteUserByID(c, id)
	if err != nil {
		response.Err(c, http.StatusInternalServerError, err.Error())
		return
	}
	response.Ok(c, nil)
}

// GetUserByID godoc
// @Summary 获取用户详情
// @Description 根据用户ID获取用户信息
// @Tags 用户管理
// @Produce json
// @Param id path int true "用户ID"
// @Success 200 {object} response.Response "成功"
// @Failure 400 {object} response.Response "请求参数错误"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Router /user/get/{id} [get]
func (h *UserHandler) GetUserByID(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Err(c, http.StatusBadRequest, err.Error())
		return
	}
	data, err := h.svc.GetUserByID(c, id)
	if err != nil {
		response.Err(c, http.StatusInternalServerError, err.Error())
		return
	}
	response.Ok(c, data)
}

// ListUsers godoc
// @Summary 获取用户列表
// @Description 支持分页、排序的用户列表查询
// @Tags 用户管理
// @Accept json
// @Produce json
// @Param data body repo.ListUsersParams true "用户列表参数"
// @Success 200 {object} response.Response "成功"
// @Failure 400 {object} response.Response "请求参数错误"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Router /user/list [get]
func (h *UserHandler) ListUsers(c *gin.Context) {
	var req repo.ListUsersParams
	if err := c.BindJSON(&req); err != nil {
		response.Err(c, http.StatusBadRequest, err.Error())
		return
	}
	data, err := h.svc.ListUsers(c, req)
	if err != nil {
		response.Err(c, http.StatusInternalServerError, err.Error())
		return
	}
	response.Ok(c, data)
}

// SearchUsersByEmail godoc
// @Summary 通过邮箱搜索用户
// @Description 根据邮箱地址模糊搜索用户
// @Tags 用户管理
// @Accept json
// @Produce json
// @Param data body repo.SearchUsersByEmailParams true "搜索参数"
// @Success 200 {object} response.Response "成功"
// @Failure 400 {object} response.Response "请求参数错误"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Router /user/search_by_email [post]
func (h *UserHandler) SearchUsersByEmail(c *gin.Context) {
	var req repo.SearchUsersByEmailParams
	if err := c.BindJSON(&req); err != nil {
		response.Err(c, http.StatusBadRequest, err.Error())
		return
	}
	data, err := h.svc.SearchUsersByEmail(c, req)
	if err != nil {
		response.Err(c, http.StatusInternalServerError, err.Error())
		return
	}
	response.Ok(c, data)
}

// SearchUsersByUserName godoc
// @Summary 通过用户名搜索用户
// @Description 根据用户名模糊搜索用户
// @Tags 用户管理
// @Accept json
// @Produce json
// @Param data body repo.SearchUsersByUserNameParams true "搜索参数"
// @Success 200 {object} response.Response "成功"
// @Failure 400 {object} response.Response "请求参数错误"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Router /user/search_by_user_name [post]
func (h *UserHandler) SearchUsersByUserName(c *gin.Context) {
	var req repo.SearchUsersByUserNameParams
	if err := c.BindJSON(&req); err != nil {
		response.Err(c, http.StatusBadRequest, err.Error())
		return
	}
	data, err := h.svc.SearchUsersByUserName(c, req)
	if err != nil {
		response.Err(c, http.StatusInternalServerError, err.Error())
		return
	}
	response.Ok(c, data)
}

// UpdateUserByID godoc
// @Summary 更新用户信息
// @Description 根据用户ID更新用户信息
// @Tags 用户管理
// @Accept json
// @Produce json
// @Param data body repo.UpdateUserByIDParams true "用户更新参数"
// @Success 200 {object} response.Response "成功"
// @Failure 400 {object} response.Response "请求参数错误"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Router /user/update [post]
func (h *UserHandler) UpdateUserByID(c *gin.Context) {
	var req repo.UpdateUserByIDParams
	if err := c.BindJSON(&req); err != nil {
		response.Err(c, http.StatusBadRequest, err.Error())
		return
	}
	err := h.svc.UpdateUserByID(c, req)
	if err != nil {
		response.Err(c, http.StatusInternalServerError, err.Error())
		return
	}
	response.Ok(c, nil)
}

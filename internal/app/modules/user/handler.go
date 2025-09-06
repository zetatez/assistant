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
	r.DELETE("/delete/:id", h.DeleteUser)
	r.GET("/get/:id", h.GetUser)
	r.GET("/list", h.ListUsers)
	r.POST("/search_by_email", h.SearchUsersByEmail)
	r.POST("/search_by_user_name", h.SearchUsersByUserName)
	r.POST("/update", h.UpdateUser)
}

func (h *UserHandler) CountUser(c *gin.Context) {
	data, err := h.svc.CountUsers(c)
	if err != nil {
		response.Err(c, http.StatusInternalServerError, err.Error())
		return
	}
	response.Ok(c, data)
}

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

func (h *UserHandler) DeleteUser(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Err(c, http.StatusBadRequest, err.Error())
		return
	}
	err = h.svc.DeleteUser(c, id)
	if err != nil {
		response.Err(c, http.StatusInternalServerError, err.Error())
		return
	}
	response.Ok(c, nil)
}

func (h *UserHandler) GetUser(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Err(c, http.StatusBadRequest, err.Error())
		return
	}
	data, err := h.svc.GetUser(c, id)
	if err != nil {
		response.Err(c, http.StatusInternalServerError, err.Error())
		return
	}
	response.Ok(c, data)
}

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

func (h *UserHandler) UpdateUser(c *gin.Context) {
	var req repo.UpdateUserParams
	if err := c.BindJSON(&req); err != nil {
		response.Err(c, http.StatusBadRequest, err.Error())
		return
	}
	err := h.svc.UpdateUser(c, req)
	if err != nil {
		response.Err(c, http.StatusInternalServerError, err.Error())
		return
	}
	response.Ok(c, nil)
}

package user

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) Register(r *gin.RouterGroup) {
	r.GET("/", h.ListUsers)
	r.POST("/", h.CreateUser)
}

// ListUsers godoc
// @Summary 获取用户列表
// @Description 获取系统中所有用户
// @Tags 用户
// @Accept json
// @Produce json
// @Success 200 {object} response.Result{data=[]User}
// @Failure 500 {object} response.Result
// @Router /users [get]
func (h *Handler) ListUsers(c *gin.Context) {
	users, err := h.svc.ListUsers()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, users)
}

// CreateUser godoc
// @Summary 创建用户
// @Description 创建一个新的用户
// @Tags 用户
// @Accept json
// @Produce json
// @Param request body CreateUserRequest true "用户信息"
// @Success 200 {object} response.Result{data=User}
// @Failure 400 {object} response.Result
// @Failure 500 {object} response.Result
// @Router /users [post]
func (h *Handler) CreateUser(c *gin.Context) {
	var req struct {
		Name string `json:"name"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	u, err := h.svc.CreateUser(req.Name)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, u)
}

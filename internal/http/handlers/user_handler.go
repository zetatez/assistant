package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/zetatez/assistant/internal/service"
	"github.com/zetatez/assistant/pkg/response"
)

type userCreateReq struct {
	Name  string `json:"name" binding:"required"`
	Email string `json:"email" binding:"required,email"`
}

type userUpdateReq struct {
	Name  string `json:"name" binding:"required"`
	Email string `json:"email" binding:"required,email"`
}

func createUser(c *gin.Context, svc service.UserService) {
	var req userCreateReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.Err(err.Error(), 400))
		return
	}
	u, err := svc.Create(req.Name, req.Email)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.Err(err.Error(), 400))
		return
	}
	c.JSON(http.StatusOK, response.OK(u))
}

func getUser(c *gin.Context, svc service.UserService) {
	id, _ := strconv.Atoi(c.Param("id"))
	u, err := svc.Get(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, response.Err(err.Error(), 404))
		return
	}
	c.JSON(http.StatusOK, response.OK(u))
}

func listUsers(c *gin.Context, svc service.UserService) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "10"))
	list, total, err := svc.List(page, size)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.Err(err.Error(), 400))
		return
	}
	c.JSON(http.StatusOK, response.OK(gin.H{"list": list, "total": total, "page": page, "size": size}))
}

func updateUser(c *gin.Context, svc service.UserService) {
	id, _ := strconv.Atoi(c.Param("id"))
	var req userUpdateReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.Err(err.Error(), 400))
		return
	}
	if err := svc.Update(uint(id), req.Name, req.Email); err != nil {
		c.JSON(http.StatusBadRequest, response.Err(err.Error(), 400))
		return
	}
	c.JSON(http.StatusOK, response.OK("updated"))
}

func deleteUser(c *gin.Context, svc service.UserService) {
	id, _ := strconv.Atoi(c.Param("id"))
	if err := svc.Delete(uint(id)); err != nil {
		c.JSON(http.StatusNotFound, response.Err(err.Error(), 404))
		return
	}
	c.JSON(http.StatusOK, response.OK("deleted"))
}

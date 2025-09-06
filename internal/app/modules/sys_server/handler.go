package sys_server

import (
	"strconv"

	"assistant/internal/app/repo"
	"assistant/internal/bootstrap/psl"
	"assistant/pkg/response"

	"github.com/gin-gonic/gin"
)

type SysServerHandler struct {
	svc *SysServerService
}

func NewSysServerHandler(svc *SysServerService) *SysServerHandler {
	return &SysServerHandler{svc: svc}
}

func (h *SysServerHandler) Register(r *gin.RouterGroup) {
	r.GET("/count", h.CountSysServers)
	r.GET("/get/:id", h.GetSysServerByID)
	r.GET("/list", h.ListSysServers)
	r.POST("/search_by_idc", h.SearchSysServersByIDC)
	r.POST("/search_by_svr_ip", h.SearchSysServersBySvrIP)
	r.POST("/search_by_idc_and_svr_ip", h.SearchSysServersByIDCAndSvrIP)
}

// CountSysServers godoc
// @Summary 统计服务器数量
// @Description 获取系统中所有服务器记录数量
// @Tags 服务器管理
// @Produce json
// @Success 200 {object} response.Response "成功"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Router /sys_server/count [get]
func (h *SysServerHandler) CountSysServers(c *gin.Context) {
	psl.GetLogger().Infof("[sys_server] count request")
	data, err := h.svc.CountSysServers(c)
	if err != nil {
		response.ErrWithInternal(c, response.CodeServerError, "failed to count servers", err)
		return
	}
	response.Ok(c, data)
}

// GetSysServerByID godoc
// @Summary 获取服务器详情
// @Description 根据服务器ID获取服务器信息
// @Tags 服务器管理
// @Produce json
// @Param id path int true "服务器ID"
// @Success 200 {object} response.Response "成功"
// @Failure 400 {object} response.Response "请求参数错误"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Router /sys_server/get/{id} [get]
func (h *SysServerHandler) GetSysServerByID(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Err(c, response.CodeInvalidParams, "invalid server id")
		return
	}
	psl.GetLogger().Infof("[sys_server] get request: id=%d", id)
	data, err := h.svc.GetSysServerByID(c, id)
	if err != nil {
		response.ErrWithInternal(c, response.CodeServerError, "failed to get server", err)
		return
	}
	response.Ok(c, data)
}

// ListSysServers godoc
// @Summary 获取服务器列表
// @Description 支持分页、排序的服务器列表查询
// @Tags 服务器管理
// @Accept json
// @Produce json
// @Param data query repo.ListSysServersParams true "查询参数"
// @Success 200 {object} response.Response "成功"
// @Failure 400 {object} response.Response "请求参数错误"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Router /sys_server/list [get]
func (h *SysServerHandler) ListSysServers(c *gin.Context) {
	var req repo.ListSysServersParams
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
	psl.GetLogger().Infof("[sys_server] list request: offset=%d limit=%d", req.Offset, req.Limit)
	data, err := h.svc.ListSysServers(c, req)
	if err != nil {
		response.ErrWithInternal(c, response.CodeServerError, "failed to list servers", err)
		return
	}
	response.Ok(c, data)
}

// SearchSysServersByIDC godoc
// @Summary 通过IDC搜索服务器
// @Description 根据IDC精确查询服务器
// @Tags 服务器管理
// @Accept json
// @Produce json
// @Param data body repo.SearchSysServersByIDCParams true "搜索参数"
// @Success 200 {object} response.Response "成功"
// @Failure 400 {object} response.Response "请求参数错误"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Router /sys_server/search_by_idc [post]
func (h *SysServerHandler) SearchSysServersByIDC(c *gin.Context) {
	var req repo.SearchSysServersByIDCParams
	if err := c.BindJSON(&req); err != nil {
		response.Err(c, response.CodeInvalidParams, "invalid request parameters")
		return
	}
	psl.GetLogger().Infof("[sys_server] search_by_idc request: idc=%s", req.IDC)
	data, err := h.svc.SearchSysServersByIDC(c, req)
	if err != nil {
		response.ErrWithInternal(c, response.CodeServerError, "failed to search servers", err)
		return
	}
	response.Ok(c, data)
}

// SearchSysServersBySvrIP godoc
// @Summary 通过IP搜索服务器
// @Description 根据IP模糊查询服务器
// @Tags 服务器管理
// @Accept json
// @Produce json
// @Param data body repo.SearchSysServersBySvrIPParams true "搜索参数"
// @Success 200 {object} response.Response "成功"
// @Failure 400 {object} response.Response "请求参数错误"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Router /sys_server/search_by_svr_ip [post]
func (h *SysServerHandler) SearchSysServersBySvrIP(c *gin.Context) {
	var req repo.SearchSysServersBySvrIPParams
	if err := c.BindJSON(&req); err != nil {
		response.Err(c, response.CodeInvalidParams, "invalid request parameters")
		return
	}
	psl.GetLogger().Infof("[sys_server] search_by_svr_ip request: ip=%s", req.SvrIP)
	data, err := h.svc.SearchSysServersBySvrIP(c, req)
	if err != nil {
		response.ErrWithInternal(c, response.CodeServerError, "failed to search servers", err)
		return
	}
	response.Ok(c, data)
}

// SearchSysServersByIDCAndSvrIP godoc
// @Summary 通过IDC和IP搜索服务器
// @Description IDC和IP组合模糊查询服务器
// @Tags 服务器管理
// @Accept json
// @Produce json
// @Param data body repo.SearchSysServersByIDCAndSvrIPParams true "搜索参数"
// @Success 200 {object} response.Response "成功"
// @Failure 400 {object} response.Response "请求参数错误"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Router /sys_server/search_by_idc_and_svr_ip [post]
func (h *SysServerHandler) SearchSysServersByIDCAndSvrIP(c *gin.Context) {
	var req repo.SearchSysServersByIDCAndSvrIPParams
	if err := c.BindJSON(&req); err != nil {
		response.Err(c, response.CodeInvalidParams, "invalid request parameters")
		return
	}
	psl.GetLogger().Infof("[sys_server] search_by_idc_and_svr_ip request: idc=%s ip=%s", req.IDC, req.SvrIP)
	data, err := h.svc.SearchSysServersByIDCAndSvrIP(c, req)
	if err != nil {
		response.ErrWithInternal(c, response.CodeServerError, "failed to search servers", err)
		return
	}
	response.Ok(c, data)
}

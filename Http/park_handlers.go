package httpserver

import (
	dao "RPW_Detection/Dao"
	models "RPW_Detection/Models"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type ParkHandler struct {
	repo *dao.Repo
}

type CreateParkRequest struct {
	Name     string `json:"name" binding:"required"`
	Code     string `json:"code" binding:"required"`
	Location string `json:"location"`
	Status   *int   `json:"status"`
}

type UpdateParkRequest struct {
	Name     string `json:"name"`
	Code     string `json:"code"`
	Location string `json:"location"`
	Status   *int   `json:"status"`
}

func NewParkHandler(repo *dao.Repo) *ParkHandler {
	return &ParkHandler{repo: repo}
}

func (h *ParkHandler) HandleListParks(c *gin.Context) {
	parks, err := h.repo.ListParks(dao.WithOrder("id DESC"))
	if err != nil {
		errorResponse(c, http.StatusInternalServerError, "查询园区列表失败")
		return
	}
	successResponse(c, gin.H{
		"total": len(parks),
		"parks": parks,
	})
}

func (h *ParkHandler) HandleGetPark(c *gin.Context) {
	id, ok := parseUintIDParam(c, "id")
	if !ok {
		return
	}

	park, err := h.repo.FindParkByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			errorResponse(c, http.StatusNotFound, "园区不存在")
			return
		}
		errorResponse(c, http.StatusInternalServerError, "查询园区失败")
		return
	}
	successResponse(c, park)
}

func (h *ParkHandler) HandleCreatePark(c *gin.Context) {
	if !requireAdmin(c, h.repo) {
		return
	}

	var req CreateParkRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		errorResponse(c, http.StatusBadRequest, "请求参数错误: "+err.Error())
		return
	}

	req.Name = strings.TrimSpace(req.Name)
	req.Code = strings.TrimSpace(req.Code)
	if req.Name == "" || req.Code == "" {
		errorResponse(c, http.StatusBadRequest, "园区名称和编码不能为空")
		return
	}

	if _, err := h.repo.FindParkByName(req.Name); err == nil {
		errorResponse(c, http.StatusConflict, "园区名称已存在")
		return
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		errorResponse(c, http.StatusInternalServerError, "查询园区失败")
		return
	}

	if _, err := h.repo.FindParkByCode(req.Code); err == nil {
		errorResponse(c, http.StatusConflict, "园区编码已存在")
		return
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		errorResponse(c, http.StatusInternalServerError, "查询园区失败")
		return
	}

	status := 1
	if req.Status != nil {
		status = *req.Status
	}

	park := &models.Park{
		Name:     req.Name,
		Code:     req.Code,
		Location: strings.TrimSpace(req.Location),
		Status:   status,
	}
	if err := h.repo.Create(park); err != nil {
		errorResponse(c, http.StatusInternalServerError, "创建园区失败")
		return
	}
	successResponse(c, park)
}

func (h *ParkHandler) HandleUpdatePark(c *gin.Context) {
	if !requireAdmin(c, h.repo) {
		return
	}

	id, ok := parseUintIDParam(c, "id")
	if !ok {
		return
	}

	park, err := h.repo.FindParkByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			errorResponse(c, http.StatusNotFound, "园区不存在")
			return
		}
		errorResponse(c, http.StatusInternalServerError, "查询园区失败")
		return
	}

	var req UpdateParkRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		errorResponse(c, http.StatusBadRequest, "请求参数错误: "+err.Error())
		return
	}

	updates := map[string]interface{}{}
	if name := strings.TrimSpace(req.Name); name != "" && name != park.Name {
		if _, err := h.repo.FindParkByName(name); err == nil {
			errorResponse(c, http.StatusConflict, "园区名称已存在")
			return
		} else if !errors.Is(err, gorm.ErrRecordNotFound) {
			errorResponse(c, http.StatusInternalServerError, "查询园区失败")
			return
		}
		updates["name"] = name
	}
	if code := strings.TrimSpace(req.Code); code != "" && code != park.Code {
		if _, err := h.repo.FindParkByCode(code); err == nil {
			errorResponse(c, http.StatusConflict, "园区编码已存在")
			return
		} else if !errors.Is(err, gorm.ErrRecordNotFound) {
			errorResponse(c, http.StatusInternalServerError, "查询园区失败")
			return
		}
		updates["code"] = code
	}
	if req.Location != "" {
		updates["location"] = strings.TrimSpace(req.Location)
	}
	if req.Status != nil {
		updates["status"] = *req.Status
	}
	if len(updates) == 0 {
		successResponse(c, park)
		return
	}

	if err := h.repo.Update(&models.Park{}, updates, dao.WithWhere("id = ?", id)); err != nil {
		errorResponse(c, http.StatusInternalServerError, "更新园区失败")
		return
	}

	updated, err := h.repo.FindParkByID(id)
	if err != nil {
		errorResponse(c, http.StatusInternalServerError, "查询更新后的园区失败")
		return
	}
	successResponse(c, updated)
}

func (h *ParkHandler) HandleDeletePark(c *gin.Context) {
	if !requireAdmin(c, h.repo) {
		return
	}

	id, ok := parseUintIDParam(c, "id")
	if !ok {
		return
	}

	if _, err := h.repo.FindParkByID(id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			errorResponse(c, http.StatusNotFound, "园区不存在")
			return
		}
		errorResponse(c, http.StatusInternalServerError, "查询园区失败")
		return
	}

	if err := h.repo.DeleteByID(&models.Park{}, id); err != nil {
		errorResponse(c, http.StatusInternalServerError, "删除园区失败")
		return
	}
	successResponse(c, gin.H{"id": id, "deleted": true})
}

func parseUintIDParam(c *gin.Context, key string) (uint, bool) {
	raw := strings.TrimSpace(c.Param(key))
	if raw == "" {
		errorResponse(c, http.StatusBadRequest, "ID不能为空")
		return 0, false
	}
	n, err := strconv.ParseUint(raw, 10, 64)
	if err != nil {
		errorResponse(c, http.StatusBadRequest, "ID格式错误")
		return 0, false
	}
	return uint(n), true
}

func requireAdmin(c *gin.Context, repo *dao.Repo) bool {
	user, err := getCurrentUser(c, repo)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			errorResponse(c, http.StatusUnauthorized, "用户不存在")
			return false
		}
		errorResponse(c, http.StatusUnauthorized, err.Error())
		return false
	}
	if user.IsAdmin != 1 {
		errorResponse(c, http.StatusForbidden, "需要管理员权限")
		return false
	}
	return true
}

func getCurrentUser(c *gin.Context, repo *dao.Repo) (*models.User, error) {
	if repo == nil {
		return nil, errors.New("权限服务未初始化")
	}

	rawUserID, exists := c.Get("user_id")
	if !exists {
		return nil, errors.New("未登录")
	}

	userIDStr, ok := rawUserID.(string)
	if !ok || strings.TrimSpace(userIDStr) == "" {
		return nil, errors.New("用户身份无效")
	}

	userID, err := strconv.ParseUint(userIDStr, 10, 64)
	if err != nil {
		return nil, errors.New("用户身份无效")
	}

	var user models.User
	if err := repo.GetByID(&user, uint(userID)); err != nil {
		return nil, err
	}
	return &user, nil
}

package httpserver

import (
	dao "RPW_Detection/Dao"
	models "RPW_Detection/Models"
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type TreeHandler struct {
	repo *dao.Repo
}

type CreateTreeRequest struct {
	ParkID   uint   `json:"park_id" binding:"required"`
	TreeCode string `json:"tree_code" binding:"required"`
	Species  string `json:"species"`
	Status   *int   `json:"status"`
}

type UpdateTreeRequest struct {
	ParkID   *uint  `json:"park_id"`
	TreeCode string `json:"tree_code"`
	Species  string `json:"species"`
	Status   *int   `json:"status"`
}

func NewTreeHandler(repo *dao.Repo) *TreeHandler {
	return &TreeHandler{repo: repo}
}

func (h *TreeHandler) HandleListTrees(c *gin.Context) {
	opts := []dao.Option{
		dao.WithOrder("id DESC"),
		dao.WithPreload("Park"),
	}

	if parkID := strings.TrimSpace(c.Query("park_id")); parkID != "" {
		opts = append(opts, dao.WithWhere("park_id = ?", parkID))
	}
	if status := strings.TrimSpace(c.Query("status")); status != "" {
		opts = append(opts, dao.WithWhere("status = ?", status))
	}

	trees, err := h.repo.ListTrees(opts...)
	if err != nil {
		errorResponse(c, http.StatusInternalServerError, "查询树木列表失败")
		return
	}
	successResponse(c, gin.H{
		"total": len(trees),
		"trees": trees,
	})
}

func (h *TreeHandler) HandleGetTree(c *gin.Context) {
	id, ok := parseUintIDParam(c, "id")
	if !ok {
		return
	}

	tree, err := h.repo.FindTreeByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			errorResponse(c, http.StatusNotFound, "树木不存在")
			return
		}
		errorResponse(c, http.StatusInternalServerError, "查询树木失败")
		return
	}
	successResponse(c, tree)
}

func (h *TreeHandler) HandleCreateTree(c *gin.Context) {
	if !requireAdmin(c, h.repo) {
		return
	}

	var req CreateTreeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		errorResponse(c, http.StatusBadRequest, "请求参数错误: "+err.Error())
		return
	}

	treeCode := strings.TrimSpace(req.TreeCode)
	if treeCode == "" {
		errorResponse(c, http.StatusBadRequest, "树木编号不能为空")
		return
	}

	if _, err := h.repo.FindParkByID(req.ParkID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			errorResponse(c, http.StatusBadRequest, "所属园区不存在")
			return
		}
		errorResponse(c, http.StatusInternalServerError, "查询园区失败")
		return
	}

	if _, err := h.repo.FindTreeByParkAndCode(req.ParkID, treeCode); err == nil {
		errorResponse(c, http.StatusConflict, "该园区下树木编号已存在")
		return
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		errorResponse(c, http.StatusInternalServerError, "查询树木失败")
		return
	}

	status := 1
	if req.Status != nil {
		status = *req.Status
	}

	tree := &models.Tree{
		ParkID:   req.ParkID,
		TreeCode: treeCode,
		Species:  strings.TrimSpace(req.Species),
		Status:   status,
	}
	if err := h.repo.Create(tree); err != nil {
		errorResponse(c, http.StatusInternalServerError, "创建树木失败")
		return
	}

	created, err := h.repo.FindTreeByID(tree.ID)
	if err != nil {
		errorResponse(c, http.StatusInternalServerError, "查询新建树木失败")
		return
	}
	successResponse(c, created)
}

func (h *TreeHandler) HandleUpdateTree(c *gin.Context) {
	if !requireAdmin(c, h.repo) {
		return
	}

	id, ok := parseUintIDParam(c, "id")
	if !ok {
		return
	}

	tree, err := h.repo.FindTreeByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			errorResponse(c, http.StatusNotFound, "树木不存在")
			return
		}
		errorResponse(c, http.StatusInternalServerError, "查询树木失败")
		return
	}

	var req UpdateTreeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		errorResponse(c, http.StatusBadRequest, "请求参数错误: "+err.Error())
		return
	}

	targetParkID := tree.ParkID
	if req.ParkID != nil {
		targetParkID = *req.ParkID
		if _, err := h.repo.FindParkByID(targetParkID); err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				errorResponse(c, http.StatusBadRequest, "所属园区不存在")
				return
			}
			errorResponse(c, http.StatusInternalServerError, "查询园区失败")
			return
		}
	}

	updates := map[string]interface{}{}
	if req.ParkID != nil && *req.ParkID != tree.ParkID {
		updates["park_id"] = *req.ParkID
	}

	treeCode := strings.TrimSpace(req.TreeCode)
	if treeCode != "" && (treeCode != tree.TreeCode || targetParkID != tree.ParkID) {
		if _, err := h.repo.FindTreeByParkAndCode(targetParkID, treeCode); err == nil {
			errorResponse(c, http.StatusConflict, "该园区下树木编号已存在")
			return
		} else if !errors.Is(err, gorm.ErrRecordNotFound) {
			errorResponse(c, http.StatusInternalServerError, "查询树木失败")
			return
		}
		updates["tree_code"] = treeCode
	}

	if req.Species != "" {
		updates["species"] = strings.TrimSpace(req.Species)
	}
	if req.Status != nil {
		updates["status"] = *req.Status
	}
	if len(updates) == 0 {
		successResponse(c, tree)
		return
	}

	if err := h.repo.Update(&models.Tree{}, updates, dao.WithWhere("id = ?", id)); err != nil {
		errorResponse(c, http.StatusInternalServerError, "更新树木失败")
		return
	}

	updated, err := h.repo.FindTreeByID(id)
	if err != nil {
		errorResponse(c, http.StatusInternalServerError, "查询更新后的树木失败")
		return
	}
	successResponse(c, updated)
}

func (h *TreeHandler) HandleDeleteTree(c *gin.Context) {
	if !requireAdmin(c, h.repo) {
		return
	}

	id, ok := parseUintIDParam(c, "id")
	if !ok {
		return
	}

	if _, err := h.repo.FindTreeByID(id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			errorResponse(c, http.StatusNotFound, "树木不存在")
			return
		}
		errorResponse(c, http.StatusInternalServerError, "查询树木失败")
		return
	}

	if err := h.repo.DeleteByID(&models.Tree{}, id); err != nil {
		errorResponse(c, http.StatusInternalServerError, "删除树木失败")
		return
	}
	successResponse(c, gin.H{"id": id, "deleted": true})
}

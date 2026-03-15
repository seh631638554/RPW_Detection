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

type DeviceHandler struct {
	repo *dao.Repo
}

type CreateDeviceRequest struct {
	DeviceID   string `json:"device_id"`
	DeviceCode string `json:"device_code"`
	DeviceName string `json:"device_name"`
	Name       string `json:"name"`
	ParkID     uint   `json:"park_id"`
	TreeID     *uint  `json:"tree_id"`
	Status     *int   `json:"status"`
}

type UpdateDeviceRequest struct {
	DeviceID   string `json:"device_id"`
	DeviceCode string `json:"device_code"`
	DeviceName string `json:"device_name"`
	Name       string `json:"name"`
	ParkID     *uint  `json:"park_id"`
	TreeID     *uint  `json:"tree_id"`
	Status     *int   `json:"status"`
}

func NewDeviceHandler(repo *dao.Repo) *DeviceHandler {
	return &DeviceHandler{repo: repo}
}

func (h *DeviceHandler) HandleListDevices(c *gin.Context) {
	opts := []dao.Option{
		dao.WithOrder("id DESC"),
		dao.WithPreload("Park", "Tree"),
	}

	if parkID := strings.TrimSpace(c.Query("park_id")); parkID != "" {
		opts = append(opts, dao.WithWhere("park_id = ?", parkID))
	}
	if treeID := strings.TrimSpace(c.Query("tree_id")); treeID != "" {
		opts = append(opts, dao.WithWhere("tree_id = ?", treeID))
	}
	if status := strings.TrimSpace(c.Query("status")); status != "" {
		opts = append(opts, dao.WithWhere("status = ?", status))
	}

	devices, err := h.repo.ListDevices(opts...)
	if err != nil {
		errorResponse(c, http.StatusInternalServerError, "查询设备列表失败")
		return
	}

	items := make([]gin.H, 0, len(devices))
	for _, device := range devices {
		items = append(items, formatDeviceResponse(device))
	}

	successResponse(c, gin.H{
		"total":   len(items),
		"devices": items,
	})
}

func (h *DeviceHandler) HandleGetDevice(c *gin.Context) {
	device, err := h.findDeviceFromParam(c.Param("id"))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			errorResponse(c, http.StatusNotFound, "设备不存在")
			return
		}
		errorResponse(c, http.StatusInternalServerError, "查询设备失败")
		return
	}
	successResponse(c, formatDeviceResponse(*device))
}

func (h *DeviceHandler) HandleCreateDevice(c *gin.Context) {
	if !requireAdmin(c, h.repo) {
		return
	}

	var req CreateDeviceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		errorResponse(c, http.StatusBadRequest, "请求参数错误: "+err.Error())
		return
	}

	deviceCode := firstNonEmpty(req.DeviceCode, req.DeviceID)
	name := firstNonEmpty(req.Name, req.DeviceName)
	deviceCode = strings.TrimSpace(deviceCode)
	name = strings.TrimSpace(name)
	if deviceCode == "" || name == "" {
		errorResponse(c, http.StatusBadRequest, "设备编码和设备名称不能为空")
		return
	}
	if req.ParkID == 0 {
		errorResponse(c, http.StatusBadRequest, "park_id不能为空")
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

	if req.TreeID != nil {
		if err := h.validateTreeBelongsToPark(*req.TreeID, req.ParkID); err != nil {
			handleDeviceValidationError(c, err)
			return
		}
	}

	if _, err := h.repo.FindDeviceByCode(deviceCode); err == nil {
		errorResponse(c, http.StatusConflict, "设备编码已存在")
		return
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		errorResponse(c, http.StatusInternalServerError, "查询设备失败")
		return
	}

	status := 1
	if req.Status != nil {
		status = *req.Status
	}

	device := &models.Device{
		DeviceCode: deviceCode,
		Name:       name,
		ParkID:     req.ParkID,
		TreeID:     req.TreeID,
		Status:     status,
	}
	if err := h.repo.Create(device); err != nil {
		errorResponse(c, http.StatusInternalServerError, "创建设备失败")
		return
	}

	created, err := h.repo.FindDeviceByID(device.ID)
	if err != nil {
		errorResponse(c, http.StatusInternalServerError, "查询新建设备失败")
		return
	}
	successResponse(c, formatDeviceResponse(*created))
}

func (h *DeviceHandler) HandleUpdateDevice(c *gin.Context) {
	if !requireAdmin(c, h.repo) {
		return
	}

	device, err := h.findDeviceFromParam(c.Param("id"))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			errorResponse(c, http.StatusNotFound, "设备不存在")
			return
		}
		errorResponse(c, http.StatusInternalServerError, "查询设备失败")
		return
	}

	var req UpdateDeviceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		errorResponse(c, http.StatusBadRequest, "请求参数错误: "+err.Error())
		return
	}

	targetParkID := device.ParkID
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

	targetTreeID := device.TreeID
	if req.TreeID != nil {
		targetTreeID = req.TreeID
		if req.TreeID != nil {
			if err := h.validateTreeBelongsToPark(*req.TreeID, targetParkID); err != nil {
				handleDeviceValidationError(c, err)
				return
			}
		}
	}

	updates := map[string]interface{}{}
	deviceCode := strings.TrimSpace(firstNonEmpty(req.DeviceCode, req.DeviceID))
	if deviceCode != "" && deviceCode != device.DeviceCode {
		if _, err := h.repo.FindDeviceByCode(deviceCode); err == nil {
			errorResponse(c, http.StatusConflict, "设备编码已存在")
			return
		} else if !errors.Is(err, gorm.ErrRecordNotFound) {
			errorResponse(c, http.StatusInternalServerError, "查询设备失败")
			return
		}
		updates["device_code"] = deviceCode
	}

	name := strings.TrimSpace(firstNonEmpty(req.Name, req.DeviceName))
	if name != "" && name != device.Name {
		updates["name"] = name
	}
	if req.ParkID != nil && *req.ParkID != device.ParkID {
		updates["park_id"] = *req.ParkID
	}
	if req.TreeID != nil {
		if targetTreeID == nil {
			updates["tree_id"] = nil
		} else {
			updates["tree_id"] = *targetTreeID
		}
	}
	if req.Status != nil {
		updates["status"] = *req.Status
	}
	if len(updates) == 0 {
		successResponse(c, formatDeviceResponse(*device))
		return
	}

	if err := h.repo.Update(&models.Device{}, updates, dao.WithWhere("id = ?", device.ID)); err != nil {
		errorResponse(c, http.StatusInternalServerError, "更新设备失败")
		return
	}

	updated, err := h.repo.FindDeviceByID(device.ID)
	if err != nil {
		errorResponse(c, http.StatusInternalServerError, "查询更新后的设备失败")
		return
	}
	successResponse(c, formatDeviceResponse(*updated))
}

func (h *DeviceHandler) HandleDeleteDevice(c *gin.Context) {
	if !requireAdmin(c, h.repo) {
		return
	}

	device, err := h.findDeviceFromParam(c.Param("id"))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			errorResponse(c, http.StatusNotFound, "设备不存在")
			return
		}
		errorResponse(c, http.StatusInternalServerError, "查询设备失败")
		return
	}

	if err := h.repo.DeleteByID(&models.Device{}, device.ID); err != nil {
		errorResponse(c, http.StatusInternalServerError, "删除设备失败")
		return
	}
	successResponse(c, gin.H{"id": device.ID, "deleted": true})
}

func (h *DeviceHandler) findDeviceFromParam(raw string) (*models.Device, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil, gorm.ErrRecordNotFound
	}
	if id, err := strconv.ParseUint(raw, 10, 64); err == nil {
		return h.repo.FindDeviceByID(uint(id))
	}
	return h.repo.FindDeviceByCode(raw)
}

func (h *DeviceHandler) validateTreeBelongsToPark(treeID uint, parkID uint) error {
	tree, err := h.repo.FindTreeByID(treeID)
	if err != nil {
		return err
	}
	if tree.ParkID != parkID {
		return errors.New("树木不属于当前园区")
	}
	return nil
}

func handleDeviceValidationError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, gorm.ErrRecordNotFound):
		errorResponse(c, http.StatusBadRequest, "所属树木不存在")
	default:
		if err.Error() == "树木不属于当前园区" {
			errorResponse(c, http.StatusBadRequest, err.Error())
			return
		}
		errorResponse(c, http.StatusInternalServerError, "校验设备关联关系失败")
	}
}

func formatDeviceResponse(device models.Device) gin.H {
	var treeID interface{}
	var treeCode string
	if device.TreeID != nil {
		treeID = *device.TreeID
	}
	if device.Tree != nil {
		treeCode = device.Tree.TreeCode
	}

	return gin.H{
		"id":          device.ID,
		"device_id":   device.DeviceCode,
		"device_code": device.DeviceCode,
		"device_name": device.Name,
		"name":        device.Name,
		"park_id":     device.ParkID,
		"park_name":   device.Park.Name,
		"tree_id":     treeID,
		"tree_code":   treeCode,
		"location":    device.Park.Name,
		"status":      device.Status,
		"created_at":  device.CreatedAt,
		"updated_at":  device.UpdatedAt,
	}
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}

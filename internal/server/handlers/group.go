package handlers

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/dlclark/regexp2"
	"github.com/gin-gonic/gin"
	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/model"
	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/observability"
	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/op"
	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/server/middleware"
	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/server/resp"
	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/server/router"
)

func respondGroupOpError(c *gin.Context, err error) {
	if err == nil {
		return
	}
	message := err.Error()
	switch {
	case errors.Is(err, context.Canceled):
		resp.Error(c, http.StatusRequestTimeout, message)
	case strings.Contains(message, "not found"):
		resp.Error(c, http.StatusNotFound, message)
	case strings.Contains(message, "invalid"), strings.Contains(message, "must be"):
		resp.Error(c, http.StatusBadRequest, message)
	default:
		resp.Error(c, http.StatusInternalServerError, "internal server error")
	}
}

func validateGroupName(name string) (string, error) {
	trimmed := strings.TrimSpace(name)
	if trimmed == "" {
		return "", fmt.Errorf("group name is required")
	}
	if strings.ContainsAny(trimmed, " :?\t\n\r") {
		return "", fmt.Errorf("group name cannot contain spaces or colon(:/?)")
	}
	return trimmed, nil
}

func init() {
	router.NewGroupRouter("/api/v1/group").
		Use(middleware.Auth()).
		Use(middleware.RequireJSON()).
		AddRoute(
			router.NewRoute("/list", http.MethodGet).
				Handle(getGroupList),
		).
		AddRoute(
			router.NewRoute("/create", http.MethodPost).
				Handle(createGroup),
		).
		AddRoute(
			router.NewRoute("/update", http.MethodPost).
				Handle(updateGroup),
		).
		AddRoute(
			router.NewRoute("/delete/:id", http.MethodDelete).
				Handle(deleteGroup),
		).
		AddRoute(
			router.NewRoute("/batch-delete", http.MethodPost).
				Handle(batchDeleteGroup),
		)
	// AddRoute(
	// 	router.NewRoute("/auto-add-item", http.MethodPost).
	// 		Handle(autoAddGroupItem),
	// )
}

func getGroupList(c *gin.Context) {
	groups, err := op.GroupList(c.Request.Context())
	if err != nil {
		resp.Error(c, http.StatusInternalServerError, "failed to list groups")
		return
	}
	resp.Success(c, groups)
}

func createGroup(c *gin.Context) {
	var group model.Group
	if err := c.ShouldBindJSON(&group); err != nil {
		resp.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	trimmedName, err := validateGroupName(group.Name)
	if err != nil {
		resp.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	group.Name = trimmedName
	if group.MatchRegex != "" {
		_, err := regexp2.Compile(group.MatchRegex, regexp2.ECMAScript)
		if err != nil {
			resp.Error(c, http.StatusBadRequest, err.Error())
			return
		}
	}
	if err := op.GroupCreate(&group, c.Request.Context()); err != nil {
		respondGroupOpError(c, err)
		return
	}
	resp.Success(c, group)
}

func updateGroup(c *gin.Context) {
	var req model.GroupUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		resp.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	if req.Name != nil {
		trimmedName, err := validateGroupName(*req.Name)
		if err != nil {
			resp.Error(c, http.StatusBadRequest, err.Error())
			return
		}
		req.Name = &trimmedName
	}
	if req.MatchRegex != nil {
		_, err := regexp2.Compile(*req.MatchRegex, regexp2.ECMAScript)
		if err != nil {
			resp.Error(c, http.StatusBadRequest, err.Error())
			return
		}
	}
	group, err := op.GroupUpdate(&req, c.Request.Context())
	if err != nil {
		respondGroupOpError(c, err)
		return
	}
	resp.Success(c, group)
}

func deleteGroup(c *gin.Context) {
	idNum, ok := parsePositivePathIDValue(c, "id")
	if !ok {
		resp.Error(c, http.StatusBadRequest, resp.ErrInvalidParam)
		return
	}
	if err := op.GroupDel(idNum, c.Request.Context()); err != nil {
		respondGroupOpError(c, err)
		return
	}
	resp.Success(c, "group deleted successfully")
}

// batchDeleteGroup 批量删除分组（分组无 enabled 字段，仅支持批量删除）
func batchDeleteGroup(c *gin.Context) {
	req, ok := parseBatchRequest(c)
	if !ok {
		return
	}
	result := runBatchOperation(c.Request.Context(), req, func(ctx context.Context, id int) error {
		var groupName string
		if g, err := op.GroupGet(id, ctx); err == nil {
			groupName = g.Name
		}
		if err := op.GroupDel(id, ctx); err != nil {
			return err
		}
		recordBatchAudit(c, observability.AuditActionDelete, observability.ResourceTypeGroup, id, groupName)
		return nil
	})
	resp.Success(c, result)
}

// func autoAddGroupItem(c *gin.Context) {
// 	var req struct {
// 		ID int `json:"id"`
// 	}
// 	if err := c.ShouldBindJSON(&req); err != nil {
// 		resp.Error(c, http.StatusBadRequest, err.Error())
// 		return
// 	}
// 	if req.ID <= 0 {
// 		resp.Error(c, http.StatusBadRequest, "invalid id")
// 		return
// 	}
// 	err := worker.AutoAddGroupItem(req.ID, c.Request.Context())
// 	if err != nil {
// 		resp.Error(c, http.StatusInternalServerError, err.Error())
// 		return
// 	}
// 	resp.Success(c, nil)
// }

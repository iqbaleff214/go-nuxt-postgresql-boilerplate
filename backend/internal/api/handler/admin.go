package handler

import (
	"net/http"
	"strconv"

	"github.com/404nfidv2/go-nuxt-starter-kit/backend/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type AdminHandler struct {
	users *service.UserService
}

func NewAdminHandler(users *service.UserService) *AdminHandler {
	return &AdminHandler{users: users}
}

// GET /api/v1/admin/users
func (h *AdminHandler) ListUsers(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	f := service.ListUsersFilter{}
	if r := c.Query("role"); r != "" {
		f.Role = &r
	}
	if s := c.Query("status"); s != "" {
		f.Status = &s
	}
	if v := c.Query("is_email_verified"); v != "" {
		b := v == "true"
		f.IsEmailVerified = &b
	}
	if q := c.Query("search"); q != "" {
		f.Search = &q
	}

	users, total, err := h.users.ListUsers(c.Request.Context(), f, page, pageSize)
	if err != nil {
		fail(c, http.StatusInternalServerError, err.Error())
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "ok",
		"data":    users,
		"meta": gin.H{
			"page":      page,
			"page_size": pageSize,
			"total":     total,
		},
	})
}

// GET /api/v1/admin/users/:id
func (h *AdminHandler) GetUser(c *gin.Context) {
	targetID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		fail(c, http.StatusBadRequest, "invalid user id")
		return
	}
	user, err := h.users.GetUserByID(c.Request.Context(), targetID)
	if err != nil {
		fail(c, http.StatusNotFound, err.Error())
		return
	}
	ok(c, http.StatusOK, "ok", user)
}

// POST /api/v1/admin/users
func (h *AdminHandler) CreateUser(c *gin.Context) {
	var req struct {
		Email     string `json:"email"      binding:"required"`
		FirstName string `json:"first_name" binding:"required"`
		LastName  string `json:"last_name"  binding:"required"`
		Role      string `json:"role"       binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, http.StatusBadRequest, err.Error())
		return
	}
	if err := h.users.AdminCreateUser(c.Request.Context(), req.Email, req.FirstName, req.LastName, req.Role); err != nil {
		fail(c, http.StatusUnprocessableEntity, err.Error())
		return
	}
	ok(c, http.StatusCreated, "User created. A set-password email has been sent.", nil)
}

// PATCH /api/v1/admin/users/:id
func (h *AdminHandler) UpdateUser(c *gin.Context) {
	callerID := mustUserID(c)
	targetID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		fail(c, http.StatusBadRequest, "invalid user id")
		return
	}
	var req struct {
		FirstName   *string `json:"first_name"`
		LastName    *string `json:"last_name"`
		DisplayName *string `json:"display_name"`
		Bio         *string `json:"bio"`
		Role        *string `json:"role"`
		Status      *string `json:"status"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, http.StatusBadRequest, err.Error())
		return
	}
	user, err := h.users.AdminUpdateUser(c.Request.Context(), callerID, targetID, service.AdminUpdateFields{
		FirstName:   req.FirstName,
		LastName:    req.LastName,
		DisplayName: req.DisplayName,
		Bio:         req.Bio,
		Role:        req.Role,
		Status:      req.Status,
	})
	if err != nil {
		fail(c, http.StatusUnprocessableEntity, err.Error())
		return
	}
	ok(c, http.StatusOK, "User updated.", user)
}

// DELETE /api/v1/admin/users/:id
func (h *AdminHandler) DeleteUser(c *gin.Context) {
	callerID := mustUserID(c)
	targetID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		fail(c, http.StatusBadRequest, "invalid user id")
		return
	}
	if err := h.users.AdminDeleteUser(c.Request.Context(), callerID, targetID); err != nil {
		fail(c, http.StatusUnprocessableEntity, err.Error())
		return
	}
	ok(c, http.StatusOK, "User deleted.", nil)
}

// POST /api/v1/admin/users/:id/activate
func (h *AdminHandler) ActivateUser(c *gin.Context) {
	h.setStatus(c, "active")
}

// POST /api/v1/admin/users/:id/deactivate
func (h *AdminHandler) DeactivateUser(c *gin.Context) {
	h.setStatus(c, "inactive")
}

// POST /api/v1/admin/users/:id/ban
func (h *AdminHandler) BanUser(c *gin.Context) {
	h.setStatus(c, "banned")
}

// POST /api/v1/admin/users/:id/unban
func (h *AdminHandler) UnbanUser(c *gin.Context) {
	h.setStatus(c, "active")
}

func (h *AdminHandler) setStatus(c *gin.Context, status string) {
	callerID := mustUserID(c)
	targetID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		fail(c, http.StatusBadRequest, "invalid user id")
		return
	}
	user, err := h.users.AdminUpdateUser(c.Request.Context(), callerID, targetID, service.AdminUpdateFields{
		Status: &status,
	})
	if err != nil {
		fail(c, http.StatusUnprocessableEntity, err.Error())
		return
	}
	ok(c, http.StatusOK, "User status updated.", user)
}

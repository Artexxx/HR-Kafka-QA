package api

import (
	"encoding/json"
	"strings"

	"github.com/Artexxx/github.com/Artexxx/HR-Kafka-QA/internal/dto"

	"github.com/valyala/fasthttp"
)

type createProfileRequest struct {
	EmployeeID    string  `json:"employee_id" example:"e-1024"`
	FirstName     *string `json:"first_name,omitempty" example:"Anna"`
	LastName      *string `json:"last_name,omitempty" example:"Ivanova"`
	BirthDate     *string `json:"birth_date,omitempty" example:"1994-06-12"`
	Email         *string `json:"email,omitempty" example:"anna@example.com"`
	Phone         *string `json:"phone,omitempty" example:"+33-123-45-67"`
	Title         *string `json:"title,omitempty" example:"QA Engineer"`
	Department    *string `json:"department,omitempty" example:"Quality Assurance"`
	Grade         *string `json:"grade,omitempty" example:"Middle"`
	EffectiveFrom *string `json:"effective_from,omitempty" example:"2025-10-01"`
}

type updateProfileRequest struct {
	FirstName     *string `json:"first_name,omitempty"`
	LastName      *string `json:"last_name,omitempty"`
	BirthDate     *string `json:"birth_date,omitempty"`
	Email         *string `json:"email,omitempty"`
	Phone         *string `json:"phone,omitempty"`
	Title         *string `json:"title,omitempty"`
	Department    *string `json:"department,omitempty"`
	Grade         *string `json:"grade,omitempty"`
	EffectiveFrom *string `json:"effective_from,omitempty"`
}

// @Summary Список профилей сотрудников
// @Tags    CRUD-Profiles
// @Produce json
// @Param   limit  query int false "Лимит"   default(50)
// @Param   offset query int false "Смещение" default(0)
// @Success 200 {object} listResponse
// @Failure 500 {object} errorResponse
// @Router  /profiles [get]
func (s *Service) listProfiles(ctx *fasthttp.RequestCtx) {
	limit, offset := parseLO(ctx)
	rows, err := s.profiles.ListProfiles(ctx, limit, offset)
	if err != nil {
		serverError(ctx, err)
		return
	}
	writeJSON(ctx, fasthttp.StatusOK, listResponse{Items: rows, Limit: limit, Offset: offset})
}

// @Summary Получить профиль по employee_id
// @Tags    CRUD-Profiles
// @Produce json
// @Param   employee_id path string true "Идентификатор сотрудника"
// @Success 200 {object} objectResponse
// @Failure 404 {object} errorResponse
// @Failure 500 {object} errorResponse
// @Router  /profiles/{employee_id} [get]
func (s *Service) getProfile(ctx *fasthttp.RequestCtx) {
	employeeID := ctx.UserValue("employee_id").(string)
	row, err := s.profiles.GetProfile(ctx, employeeID)
	if err != nil || row == nil {
		notFound(ctx, "profile_not_found", "Профиль сотрудника не найден")
		return
	}
	writeJSON(ctx, fasthttp.StatusOK, objectResponse{Data: row})
}

// @Summary Создать профиль
// @Tags    CRUD-Profiles
// @Accept  json
// @Produce json
// @Param   request body createProfileRequest true "Профиль"
// @Success 200 {object} okResponse
// @Failure 400 {object} errorResponse
// @Failure 500 {object} errorResponse
// @Router  /profiles [post]
func (s *Service) createProfile(ctx *fasthttp.RequestCtx) {
	var req createProfileRequest
	if err := json.Unmarshal(ctx.PostBody(), &req); err != nil {
		badRequest(ctx, "invalid_json", "Некорректный JSON")
		return
	}
	if strings.TrimSpace(req.EmployeeID) == "" {
		badRequest(ctx, "missing_required_field", "Отсутствует employee_id")
		return
	}

	row := dto.EmployeeProfile{
		EmployeeID:    req.EmployeeID,
		FirstName:     req.FirstName,
		LastName:      req.LastName,
		BirthDate:     req.BirthDate,
		Email:         req.Email,
		Phone:         req.Phone,
		Title:         req.Title,
		Department:    req.Department,
		Grade:         req.Grade,
		EffectiveFrom: req.EffectiveFrom,
	}
	if err := s.profiles.Create(ctx, row); err != nil {
		serverError(ctx, err)
		return
	}
	ok(ctx, "Профиль создан")
}

// @Summary Обновить профиль
// @Tags    CRUD-Profiles
// @Accept  json
// @Produce json
// @Param   employee_id path string true "Идентификатор сотрудника"
// @Param   request body updateProfileRequest true "Изменяемые поля"
// @Success 200 {object} okResponse
// @Failure 400 {object} errorResponse
// @Failure 404 {object} errorResponse
// @Failure 500 {object} errorResponse
// @Router  /profiles/{employee_id} [put]
func (s *Service) updateProfile(ctx *fasthttp.RequestCtx) {
	employeeID := ctx.UserValue("employee_id").(string)

	var req updateProfileRequest
	if err := json.Unmarshal(ctx.PostBody(), &req); err != nil {
		badRequest(ctx, "invalid_json", "Некорректный JSON")
		return
	}

	// проверим, что профиль существует
	exists, err := s.profiles.GetProfile(ctx, employeeID)
	if err != nil {
		serverError(ctx, err)
		return
	}
	if exists == nil {
		notFound(ctx, "profile_not_found", "Профиль сотрудника не найден")
		return
	}

	row := dto.EmployeeProfile{
		EmployeeID:    employeeID,
		FirstName:     req.FirstName,
		LastName:      req.LastName,
		BirthDate:     req.BirthDate,
		Email:         req.Email,
		Phone:         req.Phone,
		Title:         req.Title,
		Department:    req.Department,
		Grade:         req.Grade,
		EffectiveFrom: req.EffectiveFrom,
	}
	if err := s.profiles.Update(ctx, row); err != nil {
		serverError(ctx, err)
		return
	}
	ok(ctx, "Профиль обновлён")
}

// @Summary Удалить профиль
// @Tags    CRUD-Profiles
// @Produce json
// @Param   employee_id path string true "Идентификатор сотрудника"
// @Success 200 {object} okResponse
// @Failure 404 {object} errorResponse
// @Failure 500 {object} errorResponse
// @Router  /profiles/{employee_id} [delete]
func (s *Service) deleteProfile(ctx *fasthttp.RequestCtx) {
	employeeID := ctx.UserValue("employee_id").(string)

	// проверим существование
	exists, err := s.profiles.GetProfile(ctx, employeeID)
	if err != nil {
		serverError(ctx, err)
		return
	}
	if exists == nil {
		notFound(ctx, "profile_not_found", "Профиль сотрудника не найден")
		return
	}
	if err := s.profiles.Delete(ctx, employeeID); err != nil {
		serverError(ctx, err)
		return
	}
	ok(ctx, "Профиль удалён")
}

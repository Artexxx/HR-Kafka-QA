package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/Artexxx/HR-Kafka-QA/internal/dto"
	"github.com/valyala/fasthttp"
)

type employeeProfileReq struct {
	FirstName     string  `json:"first_name" example:"Анна"`                         // Имя сотрудника
	LastName      string  `json:"last_name" example:"Иванова"`                       // Фамилия сотрудника
	BirthDate     string  `json:"birth_date" example:"1994-06-12"`                   // Дата рождения в формате YYYY-MM-DD
	Email         string  `json:"email" example:"anna@mail.ru"`                      // Корпоративная/личная почта сотрудника
	Phone         string  `json:"phone" example:"+7 916 123-45-67"`                  // Телефон
	Title         *string `json:"title,omitempty" example:"Инженер по тестированию"` // Должность
	Department    *string `json:"department,omitempty" example:"Отдел качества"`     // Подразделение/отдел
	Grade         *string `json:"grade,omitempty" example:"Middle"`                  // Грейд (Junior, Middle, Senior или другой)
	EffectiveFrom *string `json:"effective_from,omitempty" example:"2025-10-01"`     // Дата вступления изменений в силу (YYYY-MM-DD)
}

// @Summary Список профилей сотрудников
// @Tags    CRUD-Profiles
// @Produce json
// @Success 200 {array} dto.EmployeeProfile
// @Failure 500 {object} errorResponse "Внутренняя ошибка"
// @Router  /profiles [get]
func (s *Service) listProfiles(ctx *fasthttp.RequestCtx) {
	rows, err := s.profiles.ListProfiles(ctx)

	if err != nil {
		writeError(ctx, fasthttp.StatusInternalServerError, fmt.Errorf("profileRepository.ListProfiles: %w", err))
		return
	}

	writeJSON(ctx, fasthttp.StatusOK, rows)
}

// @Summary Получить профиль по employee_id
// @Tags CRUD-Profiles
// @Produce json
// @Param employee_id path string true "Идентификатор сотрудника"
// @Success 200 {object} dto.EmployeeProfile
// @Failure 404 {object} errorResponse "employee not found"
// @Failure 500 {object} errorResponse "Внутренняя ошибка"
// @Router /profiles/{employee_id} [get]
func (s *Service) getProfile(ctx *fasthttp.RequestCtx) {
	employeeID := ctx.UserValue("employee_id").(string)
	if strings.TrimSpace(employeeID) == "" {
		writeError(ctx, fasthttp.StatusBadRequest, ErrEmployeeIDRequired)
		return
	}

	row, err := s.profiles.GetProfile(ctx, employeeID)

	if err != nil {
		if errors.Is(err, dto.ErrNotFound) {
			writeError(ctx, fasthttp.StatusNotFound, ErrProfileNotFound)
			return
		}

		writeError(ctx, fasthttp.StatusInternalServerError, fmt.Errorf("profileRepository.GetProfile: %w", err))
		return
	}

	writeJSON(ctx, fasthttp.StatusOK, row)
}

// @Summary Создать профиль
// @Tags    CRUD-Profiles
// @Accept  json
// @Produce json
// @Param   request body dto.EmployeeProfile true "Профиль"
// @Success 200 {object} okResponse
// @Failure 400 {object} errorResponse "VALIDATION ERROR — ошибки валидации входных данных"
// @description Варианты 400 (VALIDATION ERROR):
// @description - required: employee_id, first_name, birth_date, email, phone
// @description - required (если присутствует): title, department, grade, effective_from
// @description - invalid value: email, birth_date, effective_from (если присутствует)
// @description - invalid enum (если присутствует): grade in {Junior, Middle, Senior, Lead, Head}
// @Failure 409 {object} errorResponse "employee already exists"
// @Failure 500 {object} errorResponse "Внутренняя ошибка"
// @Router  /profiles [post]
func (s *Service) createProfile(ctx *fasthttp.RequestCtx) {
	var req dto.EmployeeProfile

	err := json.Unmarshal(ctx.PostBody(), &req)
	if err != nil {
		writeError(ctx, fasthttp.StatusBadRequest, fmt.Errorf("json.Unmarshal: %w", err))
		return
	}

	if msg := validateEmployeeProfile(req); msg != "" {
		writeError(ctx, fasthttp.StatusBadRequest, errors.New(msg))
		return
	}

	if err := s.profiles.Create(ctx, req); err != nil {
		if errors.Is(err, dto.ErrAlreadyExists) {
			writeError(ctx, fasthttp.StatusConflict, ErrProfileAlreadyExists)
			return
		}

		writeError(ctx, fasthttp.StatusInternalServerError, fmt.Errorf("profileRepository.Create: %w", err))
		return
	}

	ok(ctx, "Профиль создан")
}

// @Summary Обновить профиль
// @Tags    CRUD-Profiles
// @Accept  json
// @Produce json
// @Param   employee_id path string true "Идентификатор сотрудника"
// @Param   request body employeeProfileReq true "Профиль"
// @Success 200 {object} okResponse
// @Failure 400 {object} errorResponse "VALIDATION ERROR — ошибки валидации входных данных"
// @description Варианты 400 (VALIDATION ERROR):
// @description - required: employee_id, first_name, birth_date, email, phone
// @description - required (если присутствует): title, department, grade, effective_from
// @description - invalid value: email, birth_date, effective_from (если присутствует)
// @description - invalid enum (если присутствует): grade in {Junior, Middle, Senior, Lead, Head}
// @Failure 404 {object} errorResponse "employee not found"
// @Failure 500 {object} errorResponse "Внутренняя ошибка"
// @Router  /profiles/{employee_id} [put]
func (s *Service) updateProfile(ctx *fasthttp.RequestCtx) {
	var req employeeProfileReq

	err := json.Unmarshal(ctx.PostBody(), &req)
	if err != nil {
		writeError(ctx, fasthttp.StatusBadRequest, fmt.Errorf("json.Unmarshal: %w", err))
		return
	}

	employeeID := ctx.UserValue("employee_id").(string)
	if strings.TrimSpace(employeeID) == "" {
		writeError(ctx, fasthttp.StatusBadRequest, ErrEmployeeIDRequired)
		return
	}

	var employee = dto.EmployeeProfile{
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

	if msg := validateEmployeeProfile(employee); msg != "" {
		writeError(ctx, fasthttp.StatusBadRequest, errors.New(msg))
		return
	}

	if err := s.profiles.Update(ctx, employee); err != nil {
		if errors.Is(err, dto.ErrNotFound) {
			writeError(ctx, fasthttp.StatusNotFound, ErrProfileNotFound)

			return
		}

		writeError(ctx, fasthttp.StatusInternalServerError, fmt.Errorf("profileRepository.Update: %w", err))
		return
	}

	ok(ctx, "Профиль обновлён")
}

// @Summary Удалить профиль
// @Tags    CRUD-Profiles
// @Produce json
// @Param   employee_id path string true "Идентификатор сотрудника"
// @Success 200 {object} okResponse
// @Failure 400 {object} errorResponse "required field 'employee_id'"
// @Failure 404 {object} errorResponse "employee not found"
// @Failure 500 {object} errorResponse "Внутренняя ошибка"
// @Router  /profiles/{employee_id} [delete]
func (s *Service) deleteProfile(ctx *fasthttp.RequestCtx) {
	employeeID := ctx.UserValue("employee_id").(string)

	if strings.TrimSpace(employeeID) == "" {
		writeError(ctx, fasthttp.StatusBadRequest, ErrEmployeeIDRequired)
		return
	}

	if err := s.profiles.Delete(ctx, employeeID); err != nil {
		if errors.Is(err, dto.ErrNotFound) {
			writeError(ctx, fasthttp.StatusNotFound, ErrProfileNotFound)

			return
		}

		writeError(ctx, fasthttp.StatusInternalServerError, fmt.Errorf("profileRepository.Delete: %w", err))
		return
	}

	ok(ctx, "Профиль удалён")
}

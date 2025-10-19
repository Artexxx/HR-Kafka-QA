package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/Artexxx/HR-Kafka-QA/internal/dto"
	"github.com/valyala/fasthttp"
)

// @Summary Список профилей сотрудников
// @Tags    CRUD-Profiles
// @Produce json
// @Success 200 {array} dto.EmployeeProfile
// @Failure 500 {object} errorResponse "InternalError — внутренняя ошибка"
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
// @Failure 400 {object} errorResponse "Отсутствует employee_id"
// @Failure 404 {object} errorResponse "Сотрудник не найден"
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
// @Failure 400 {object} errorResponse "Отсутствует employee_id"
// @Failure 409 {object} errorResponse "Профиль уже существует"
// @Failure 500 {object} errorResponse "Внутренняя ошибка"
// @Router  /profiles [post]
func (s *Service) createProfile(ctx *fasthttp.RequestCtx) {
	var req dto.EmployeeProfile

	err := json.Unmarshal(ctx.PostBody(), &req)
	if err != nil {
		writeError(ctx, fasthttp.StatusBadRequest, fmt.Errorf("json.Unmarshal: %w", err))
		return
	}

	if strings.TrimSpace(req.EmployeeID) == "" {
		writeError(ctx, fasthttp.StatusBadRequest, ErrEmployeeIDRequired)
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
// @Param   request body dto.EmployeeProfile true "Профиль"
// @Success 200 {object} okResponse
// @Failure 400 {object} errorResponse "Отсутствует employee_id"
// @Failure 404 {object} errorResponse "Сотрудник не найден"
// @Failure 500 {object} errorResponse "Внутренняя ошибка"
// @Router  /profiles/{employee_id} [put]
func (s *Service) updateProfile(ctx *fasthttp.RequestCtx) {
	var req dto.EmployeeProfile

	err := json.Unmarshal(ctx.PostBody(), &req)
	if err != nil {
		writeError(ctx, fasthttp.StatusBadRequest, fmt.Errorf("json.Unmarshal: %w", err))
		return
	}

	req.EmployeeID = ctx.UserValue("employee_id").(string)

	if strings.TrimSpace(req.EmployeeID) == "" {
		writeError(ctx, fasthttp.StatusBadRequest, ErrEmployeeIDRequired)
		return
	}

	if err := s.profiles.Update(ctx, req); err != nil {
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
// @Failure 400 {object} errorResponse "Отсутствует employee_id"
// @Failure 404 {object} errorResponse "Сотрудник не найден"
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

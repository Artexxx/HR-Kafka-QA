package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"github.com/Artexxx/HR-Kafka-QA/internal/dto"
	"github.com/valyala/fasthttp"
)

// employmentHistoryRequest — запись истории работы сотрудника.
type employmentHistoryRequest struct {
	EmployeeID string   `json:"employee_id" example:"e-1024"`             // Идентификатор сотрудника
	Company    string   `json:"company" example:"ООО Ромашка"`            // Компания
	Position   string   `json:"position,omitempty" example:"Инженер QA"`  // Должность
	PeriodFrom string   `json:"period_from" example:"2022-07-01"`         // Дата начала периода занятости (YYYY-MM-DD)
	PeriodTo   string   `json:"period_to" example:"2025-09-30"`           // Дата окончания периода занятости (YYYY-MM-DD)
	Stack      []string `json:"stack" example:"Python,Pytest,PostgreSQL"` // Технологический стек (список строк)
}

// @Summary История работы сотрудника
// @Tags    CRUD-History
// @Produce json
// @Param   employee_id path string true "Идентификатор сотрудника"
// @Success 200 {object} dto.EmploymentHistory
// @Failure 400 {object} errorResponse "required field 'employee_id'"
// @Failure 500 {object} errorResponse
// @Router  /history/{employee_id} [get]
func (s *Service) listHistoryByEmployee(ctx *fasthttp.RequestCtx) {
	employeeID := ctx.UserValue("employee_id").(string)

	employeeID, err := url.PathUnescape(employeeID)
	if err != nil {
		writeError(ctx, fasthttp.StatusBadRequest, ErrEmployeeIDRequired)
		return
	}

	if strings.TrimSpace(employeeID) == "" {
		writeError(ctx, fasthttp.StatusBadRequest, ErrEmployeeIDRequired)
		return
	}

	rows, err := s.history.ListByEmployee(ctx, employeeID)

	if err != nil {
		writeError(ctx, fasthttp.StatusInternalServerError, fmt.Errorf("history.ListByEmployee: %w", err))
		return
	}

	writeJSON(ctx, fasthttp.StatusOK, rows)
}

// @Summary Добавить запись в историю работы
// @Tags    CRUD-History
// @Accept  json
// @Produce json
// @Param   request body employmentHistoryRequest true "История"
// @Failure 400 {object} errorResponse "VALIDATION ERROR — ошибки валидации входных данных"
// @description Варианты 400 (VALIDATION ERROR):
// @description - required: employee_id, company, period_from, period_to
// @description - invalid value: period_from, period_to, period (to < from)
// @Failure 412 {object} errorResponse "Внутренняя ошибка"
// @Failure 500 {object} errorResponse "Внутренняя ошибка"
// @Router  /history [post]
func (s *Service) createHistory(ctx *fasthttp.RequestCtx) {
	var req employmentHistoryRequest

	err := json.Unmarshal(ctx.PostBody(), &req)
	if err != nil {
		writeError(ctx, fasthttp.StatusBadRequest, fmt.Errorf("json.Unmarshal: %w", err))
		return
	}

	if strings.TrimSpace(req.EmployeeID) == "" {
		writeError(ctx, fasthttp.StatusBadRequest, ErrEmployeeIDRequired)
		return
	}

	row := dto.EmploymentHistory{
		EmployeeID: req.EmployeeID,
		Company:    req.Company,
		Position:   req.Position,
		PeriodFrom: req.PeriodFrom,
		PeriodTo:   req.PeriodTo,
		Stack:      req.Stack,
	}

	if msg := validateEmploymentHistory(row); msg != "" {
		writeError(ctx, fasthttp.StatusBadRequest, errors.New(msg))
		return
	}

	if _, err := s.profiles.GetProfile(ctx, row.EmployeeID); err != nil {
		if errors.Is(err, dto.ErrNotFound) {
			writeError(ctx, fasthttp.StatusPreconditionFailed, fmt.Errorf("employee_id=%s not found: create employee profile first", row.EmployeeID))
			return
		}

		writeError(ctx, fasthttp.StatusInternalServerError, fmt.Errorf("profiles.GetProfile employee_id=%s: %w", row.EmployeeID, err))
		return
	}

	if err := s.history.Insert(ctx, row); err != nil {
		writeError(ctx, fasthttp.StatusInternalServerError, fmt.Errorf("historyRepository.Create: %w", err))
		return
	}

	ok(ctx, "История добавлена")
}

// @Summary Обновить запись истории
// @Tags    CRUD-History
// @Accept  json
// @Produce json
// @Param   id path int true "ID записи истории"
// @Param   request body dto.EmploymentHistory true "Изменяемые поля"
// @Success 200 {object} okResponse
// @Failure 400 {object} errorResponse "VALIDATION ERROR — ошибки валидации входных данных"
// @description Варианты 400 (VALIDATION ERROR):
// @description - required: id, employee_id, company, period_from, period_to
// @description - invalid value: id, period_from, period_to, period (to < from)
// @Failure 404 {object} errorResponse "Запись не найдена"
// @Failure 500 {object} errorResponse
// @Router  /history/{id} [put]
func (s *Service) updateHistory(ctx *fasthttp.RequestCtx) {
	var req employmentHistoryRequest

	err := json.Unmarshal(ctx.PostBody(), &req)
	if err != nil {
		writeError(ctx, fasthttp.StatusBadRequest, fmt.Errorf("json.Unmarshal: %w", err))
		return
	}

	idStr := ctx.UserValue("id").(string)
	if idStr == "" {
		writeError(ctx, fasthttp.StatusBadRequest, ErrHistoryIDRequired)
	}

	row := dto.EmploymentHistory{
		EmployeeID: req.EmployeeID,
		Company:    req.Company,
		Position:   req.Position,
		PeriodFrom: req.PeriodFrom,
		PeriodTo:   req.PeriodTo,
		Stack:      req.Stack,
	}

	row.ID, err = strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		writeError(ctx, fasthttp.StatusBadRequest, fmt.Errorf("invalid value in field 'id'=%s", idStr))
		return
	}

	if row.ID <= 0 {
		writeError(ctx, fasthttp.StatusBadRequest, errors.New("required field 'id'"))
		return
	}

	if msg := validateEmploymentHistory(row); msg != "" {
		writeError(ctx, fasthttp.StatusBadRequest, errors.New(msg))
		return
	}

	if err := s.history.Update(ctx, row); err != nil {
		if errors.Is(err, dto.ErrNotFound) {
			writeError(ctx, fasthttp.StatusNotFound, ErrHistoryNotFound)

			return
		}

		writeError(ctx, fasthttp.StatusInternalServerError, fmt.Errorf("historyRepository.Update: %w", err))
		return
	}

	ok(ctx, "Запись истории обновлена")
}

// @Summary Удалить запись истории
// @Tags    CRUD-History
// @Produce json
// @Param   id path int true "ID записи истории"
// @Failure 400 {object} errorResponse "required field 'id'"
// @Failure 404 {object} errorResponse "history not found"
// @Failure 500 {object} errorResponse "Внутренняя ошибка"
// @Router  /history/{id} [delete]
func (s *Service) deleteHistory(ctx *fasthttp.RequestCtx) {
	idStr := ctx.UserValue("id").(string)
	if idStr == "" {
		writeError(ctx, fasthttp.StatusBadRequest, ErrHistoryIDRequired)
	}

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		writeError(ctx, fasthttp.StatusBadRequest, fmt.Errorf("invalid value in field 'id'=%s", idStr))
		return
	}

	if id <= 0 {
		writeError(ctx, fasthttp.StatusBadRequest, errors.New("required field 'id'"))
		return
	}

	if err := s.history.Delete(ctx, id); err != nil {
		if errors.Is(err, dto.ErrNotFound) {
			writeError(ctx, fasthttp.StatusNotFound, ErrHistoryNotFound)

			return
		}

		writeError(ctx, fasthttp.StatusInternalServerError, fmt.Errorf("historyRepository.Delete: %w", err))
		return
	}

	ok(ctx, "Запись истории удалена")
}

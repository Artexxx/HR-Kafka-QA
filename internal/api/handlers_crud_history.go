package api

import (
	"encoding/json"
	"strconv"
	"strings"

	"github.com/Artexxx/HR-Kafka-QA/internal/dto"

	"github.com/valyala/fasthttp"
)

type createHistoryRequest struct {
	EmployeeID string   `json:"employee_id" example:"e-1024"`
	Company    string   `json:"company" example:"Acme Corp"`
	Position   *string  `json:"position,omitempty" example:"QA"`
	PeriodFrom string   `json:"period_from" example:"2022-07-01"`
	PeriodTo   string   `json:"period_to" example:"2025-09-30"`
	Stack      []string `json:"stack" example:"Python,Pytest"`
}

type updateHistoryRequest struct {
	Company    *string  `json:"company,omitempty"`
	Position   *string  `json:"position,omitempty"`
	PeriodFrom *string  `json:"period_from,omitempty"`
	PeriodTo   *string  `json:"period_to,omitempty"`
	Stack      []string `json:"stack,omitempty"`
}

// @Summary История работы сотрудника
// @Tags    CRUD-History
// @Produce json
// @Param   employee_id path string true "Идентификатор сотрудника"
// @Param   limit  query int false "Лимит"   default(50)
// @Param   offset query int false "Смещение" default(0)
// @Success 200 {object} listResponse
// @Failure 500 {object} errorResponse
// @Router  /history/{employee_id} [get]
func (s *Service) listHistoryByEmployee(ctx *fasthttp.RequestCtx) {
	employeeID := ctx.UserValue("employee_id").(string)
	limit, offset := parseLO(ctx)
	rows, err := s.history.ListByEmployee(ctx, employeeID, limit, offset)
	if err != nil {
		serverError(ctx, err)
		return
	}
	writeJSON(ctx, fasthttp.StatusOK, listResponse{Items: rows, Limit: limit, Offset: offset})
}

// @Summary Добавить запись в историю работы
// @Tags    CRUD-History
// @Accept  json
// @Produce json
// @Param   request body createHistoryRequest true "История"
// @Success 200 {object} okResponse
// @Failure 400 {object} errorResponse
// @Failure 500 {object} errorResponse
// @Router  /history [post]
func (s *Service) createHistory(ctx *fasthttp.RequestCtx) {
	var req createHistoryRequest
	if err := json.Unmarshal(ctx.PostBody(), &req); err != nil {
		badRequest(ctx, "invalid_json", "Некорректный JSON")
		return
	}
	if strings.TrimSpace(req.EmployeeID) == "" || strings.TrimSpace(req.Company) == "" {
		badRequest(ctx, "missing_required_field", "Отсутствует employee_id или company")
		return
	}
	if req.PeriodTo < req.PeriodFrom {
		badRequest(ctx, "invalid_period", "Дата окончания раньше даты начала")
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
	if err := s.history.Insert(ctx, row); err != nil {
		serverError(ctx, err)
		return
	}
	ok(ctx, "История добавлена")
}

// @Summary Обновить запись истории
// @Tags    CRUD-History
// @Accept  json
// @Produce json
// @Param   id path int true "ID записи истории"
// @Param   request body updateHistoryRequest true "Изменяемые поля"
// @Success 200 {object} okResponse
// @Failure 400 {object} errorResponse
// @Failure 404 {object} errorResponse
// @Failure 500 {object} errorResponse
// @Router  /history/{id} [put]
func (s *Service) updateHistory(ctx *fasthttp.RequestCtx) {
	idStr := ctx.UserValue("id").(string)
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id <= 0 {
		badRequest(ctx, "invalid_id", "Некорректный id записи")
		return
	}

	exists, err := s.history.GetByID(ctx, id)
	if err != nil {
		serverError(ctx, err)
		return
	}
	if exists == nil {
		notFound(ctx, "history_not_found", "Запись истории не найдена")
		return
	}

	var req updateHistoryRequest
	if err := json.Unmarshal(ctx.PostBody(), &req); err != nil {
		badRequest(ctx, "invalid_json", "Некорректный JSON")
		return
	}

	row := dto.EmploymentHistory{
		ID:         id,
		EmployeeID: exists.EmployeeID,
		Company:    exists.Company,
		Position:   exists.Position,
		PeriodFrom: exists.PeriodFrom,
		PeriodTo:   exists.PeriodTo,
		Stack:      exists.Stack,
	}

	if req.Company != nil {
		row.Company = *req.Company
	}
	if req.Position != nil {
		row.Position = req.Position
	}
	if req.PeriodFrom != nil {
		row.PeriodFrom = *req.PeriodFrom
	}
	if req.PeriodTo != nil {
		row.PeriodTo = *req.PeriodTo
	}
	if req.Stack != nil && len(req.Stack) > 0 {
		row.Stack = req.Stack
	}

	if row.PeriodTo < row.PeriodFrom {
		badRequest(ctx, "invalid_period", "Дата окончания раньше даты начала")
		return
	}

	if err := s.history.Update(ctx, row); err != nil {
		serverError(ctx, err)
		return
	}
	ok(ctx, "Запись истории обновлена")
}

// @Summary Удалить запись истории
// @Tags    CRUD-History
// @Produce json
// @Param   id path int true "ID записи истории"
// @Success 200 {object} okResponse
// @Failure 404 {object} errorResponse
// @Failure 500 {object} errorResponse
// @Router  /history/{id} [delete]
func (s *Service) deleteHistory(ctx *fasthttp.RequestCtx) {
	idStr := ctx.UserValue("id").(string)
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id <= 0 {
		badRequest(ctx, "invalid_id", "Некорректный id записи")
		return
	}

	exists, err := s.history.GetByID(ctx, id)
	if err != nil {
		serverError(ctx, err)
		return
	}
	if exists == nil {
		notFound(ctx, "history_not_found", "Запись истории не найдена")
		return
	}

	if err := s.history.Delete(ctx, id); err != nil {
		serverError(ctx, err)
		return
	}
	ok(ctx, "Запись истории удалена")
}

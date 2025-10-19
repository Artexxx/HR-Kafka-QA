package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/Artexxx/HR-Kafka-QA/internal/dto"
	"github.com/valyala/fasthttp"
)

// @Summary История работы сотрудника
// @Tags    CRUD-History
// @Produce json
// @Param   employee_id path string true "Идентификатор сотрудника"
// @Success 200 {object} dto.EmploymentHistory
// @Failure 400 {object} errorResponse "Отсутствует employee_id"
// @Failure 404 {object} errorResponse "Сотрудник не найден"
// @Failure 500 {object} errorResponse
// @Router  /history/{employee_id} [get]
func (s *Service) listHistoryByEmployee(ctx *fasthttp.RequestCtx) {
	employeeID := ctx.UserValue("employee_id").(string)
	if strings.TrimSpace(employeeID) == "" {
		writeError(ctx, fasthttp.StatusBadRequest, ErrEmployeeIDRequired)
		return
	}

	rows, err := s.history.ListByEmployee(ctx, employeeID)

	if err != nil {
		if errors.Is(err, dto.ErrNotFound) {
			writeError(ctx, fasthttp.StatusNotFound, ErrHistoryNotFound)
			return
		}

		writeError(ctx, fasthttp.StatusInternalServerError, fmt.Errorf("history.ListByEmployee: %w", err))
		return
	}

	writeJSON(ctx, fasthttp.StatusOK, rows)
}

// @Summary Добавить запись в историю работы
// @Tags    CRUD-History
// @Accept  json
// @Produce json
// @Param   request body dto.EmploymentHistory true "История"
// @Failure 400 {object} errorResponse "Отсутствует employee_id"
// @Failure 500 {object} errorResponse "Внутренняя ошибка"
// @Router  /history [post]
func (s *Service) createHistory(ctx *fasthttp.RequestCtx) {
	var req dto.EmploymentHistory

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
// @Param   request body dto.EmploymentHistory true "Изменяемые поля"
// @Success 200 {object} okResponse
// @Failure 400 {object} errorResponse "Отсутствует id"
// @Failure 404 {object} errorResponse "Запись не найдена"
// @Failure 500 {object} errorResponse
// @Router  /history/{id} [put]
func (s *Service) updateHistory(ctx *fasthttp.RequestCtx) {
	var req dto.EmploymentHistory

	err := json.Unmarshal(ctx.PostBody(), &req)
	if err != nil {
		writeError(ctx, fasthttp.StatusBadRequest, fmt.Errorf("json.Unmarshal: %w", err))
		return
	}

	idStr := ctx.UserValue("id").(string)
	if idStr == "" {
		writeError(ctx, fasthttp.StatusBadRequest, ErrHistoryIDRequired)
	}

	req.ID, err = strconv.ParseInt(idStr, 10, 64)
	if err != nil || req.ID <= 0 {
		writeError(ctx, fasthttp.StatusBadRequest, errors.New("некорректный id"))
		return
	}

	if err := s.history.Update(ctx, req); err != nil {
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
// @Failure 400 {object} errorResponse "Отсутствует id"
// @Failure 404 {object} errorResponse "История не найдена"
// @Failure 500 {object} errorResponse "Внутренняя ошибка"
// @Router  /history/{id} [delete]
func (s *Service) deleteHistory(ctx *fasthttp.RequestCtx) {
	idStr := ctx.UserValue("id").(string)
	if idStr == "" {
		writeError(ctx, fasthttp.StatusBadRequest, ErrHistoryIDRequired)
	}

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id <= 0 {
		writeError(ctx, fasthttp.StatusBadRequest, errors.New("некорректный id"))
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

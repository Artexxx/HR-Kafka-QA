package api

import (
	"encoding/json"
	"strconv"
	"strings"

	"github.com/Artexxx/HR-Kafka-QA/internal/dto"
	"github.com/valyala/fasthttp"
)

type personalProduceRequest struct {
	MessageID  string  `json:"message_id"`
	EmployeeID string  `json:"employee_id"`
	FirstName  *string `json:"first_name"`
	LastName   *string `json:"last_name"`
	BirthDate  *string `json:"birth_date"`
	Email      *string `json:"email"`
	Phone      *string `json:"phone"`
}

type positionProduceRequest struct {
	MessageID     string  `json:"message_id"`
	EmployeeID    string  `json:"employee_id"`
	Title         *string `json:"title"`
	Department    *string `json:"department"`
	Grade         *string `json:"grade"`
	EffectiveFrom *string `json:"effective_from"`
}

type historyProduceRequest struct {
	MessageID  string   `json:"message_id"`
	EmployeeID string   `json:"employee_id"`
	Company    string   `json:"company"`
	Position   *string  `json:"position"`
	PeriodFrom string   `json:"period_from"`
	PeriodTo   string   `json:"period_to"`
	Stack      []string `json:"stack"`
}

// @Summary Публикация события в hr.personal
// @Tags    Producer
// @Accept  json
// @Produce json
// @Param   request body personalProduceRequest true "payload"
// @Success 200 {object} okResponse
// @Failure 400 {object} errorResponse
// @Failure 501 {object} errorResponse
// @Failure 500 {object} errorResponse
// @Router  /producer/personal [post]
func (s *Service) producerPersonal(ctx *fasthttp.RequestCtx) {
	if s.producer == nil {
		notImplemented(ctx, "producer_not_configured", "Продюсер не настроен. Обратитесь к администратору стенда.")
		return
	}

	var req personalProduceRequest
	if err := json.Unmarshal(ctx.PostBody(), &req); err != nil {
		badRequest(ctx, "invalid_json", "Некорректный JSON")
		return
	}
	if strings.TrimSpace(req.MessageID) == "" || strings.TrimSpace(req.EmployeeID) == "" {
		badRequest(ctx, "missing_required_field", "Отсутствует message_id или employee_id")
		return
	}

	prof := dto.EmployeeProfile{
		EmployeeID:    req.EmployeeID,
		FirstName:     req.FirstName,
		LastName:      req.LastName,
		BirthDate:     req.BirthDate,
		Email:         req.Email,
		Phone:         req.Phone,
		Title:         nil, // не используется в personal
		Department:    nil,
		Grade:         nil,
		EffectiveFrom: nil,
	}

	if err := s.producer.ProducePersonal(ctx, req.MessageID, prof); err != nil {
		serverError(ctx, err)
		return
	}
	ok(ctx, "Событие отправлено в hr.personal")
}

// @Summary Публикация события в hr.positions
// @Tags    Producer
// @Accept  json
// @Produce json
// @Param   request body positionProduceRequest true "payload"
// @Success 200 {object} okResponse
// @Failure 400 {object} errorResponse
// @Failure 501 {object} errorResponse
// @Failure 500 {object} errorResponse
// @Router  /producer/position [post]
func (s *Service) producerPosition(ctx *fasthttp.RequestCtx) {
	if s.producer == nil {
		notImplemented(ctx, "producer_not_configured", "Продюсер не настроен. Обратитесь к администратору стенда.")
		return
	}

	var req positionProduceRequest
	if err := json.Unmarshal(ctx.PostBody(), &req); err != nil {
		badRequest(ctx, "invalid_json", "Некорректный JSON")
		return
	}
	if strings.TrimSpace(req.MessageID) == "" || strings.TrimSpace(req.EmployeeID) == "" {
		badRequest(ctx, "missing_required_field", "Отсутствует message_id или employee_id")
		return
	}

	prof := dto.EmployeeProfile{
		EmployeeID:    req.EmployeeID,
		Title:         req.Title,
		Department:    req.Department,
		Grade:         req.Grade,
		EffectiveFrom: req.EffectiveFrom,
	}

	if err := s.producer.ProducePosition(ctx, req.MessageID, prof); err != nil {
		serverError(ctx, err)
		return
	}
	ok(ctx, "Событие отправлено в hr.positions")
}

// @Summary Публикация события в hr.history
// @Tags    Producer
// @Accept  json
// @Produce json
// @Param   request body historyProduceRequest true "payload"
// @Success 200 {object} okResponse
// @Failure 400 {object} errorResponse
// @Failure 501 {object} errorResponse
// @Failure 500 {object} errorResponse
// @Router  /producer/history [post]
func (s *Service) producerHistory(ctx *fasthttp.RequestCtx) {
	if s.producer == nil {
		notImplemented(ctx, "producer_not_configured", "Продюсер не настроен. Обратитесь к администратору стенда.")
		return
	}

	var req historyProduceRequest
	if err := json.Unmarshal(ctx.PostBody(), &req); err != nil {
		badRequest(ctx, "invalid_json", "Некорректный JSON")
		return
	}
	if strings.TrimSpace(req.MessageID) == "" || strings.TrimSpace(req.EmployeeID) == "" {
		badRequest(ctx, "missing_required_field", "Отсутствует message_id или employee_id")
		return
	}
	if req.PeriodTo < req.PeriodFrom {
		badRequest(ctx, "invalid_period", "Дата окончания раньше даты начала")
		return
	}

	h := dto.EmploymentHistory{
		EmployeeID: req.EmployeeID,
		Company:    req.Company,
		Position:   req.Position,
		PeriodFrom: req.PeriodFrom,
		PeriodTo:   req.PeriodTo,
		Stack:      req.Stack,
	}

	if err := s.producer.ProduceHistory(ctx, req.MessageID, h); err != nil {
		serverError(ctx, err)
		return
	}
	ok(ctx, "Событие отправлено в hr.history")
}

// @Summary Сырые события (эмуляция kafka_events)
// @Tags    Producer
// @Produce json
// @Param   limit  query int false "Лимит"   default(50)
// @Param   offset query int false "Смещение" default(0)
// @Success 200 {object} listResponse
// @Failure 500 {object} errorResponse
// @Router  /events [get]
func (s *Service) listEvents(ctx *fasthttp.RequestCtx) {
	limit, offset := parseLO(ctx)
	rows, err := s.events.ListEvents(ctx, limit, offset)
	if err != nil {
		serverError(ctx, err)
		return
	}

	writeJSON(ctx, fasthttp.StatusOK, listResponse{Items: rows, Limit: limit, Offset: offset})
}

// @Summary Сообщения DLQ
// @Tags    Producer
// @Produce json
// @Param   limit  query int false "Лимит"   default(50)
// @Param   offset query int false "Смещение" default(0)
// @Success 200 {object} listResponse
// @Failure 500 {object} errorResponse
// @Router  /dlq [get]
func (s *Service) listDLQ(ctx *fasthttp.RequestCtx) {
	limit, offset := parseLO(ctx)
	rows, err := s.events.ListDLQ(ctx, limit, offset)
	if err != nil {
		serverError(ctx, err)
		return
	}
	writeJSON(ctx, fasthttp.StatusOK, listResponse{Items: rows, Limit: limit, Offset: offset})
}

func parseLO(ctx *fasthttp.RequestCtx) (int, int) {
	q := ctx.URI().QueryArgs()
	limit := 50
	offset := 0

	if v := q.GetUfloatOrZero("limit"); v > 0 && v <= 500 {
		limit = int(v)
	}
	if s := string(q.Peek("offset")); s != "" {
		if x, err := strconv.Atoi(s); err == nil && x >= 0 {
			offset = x
		}
	}

	return limit, offset
}

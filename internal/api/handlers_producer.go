package api

import (
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"strings"

	"github.com/Artexxx/HR-Kafka-QA/internal/dto"
	"github.com/valyala/fasthttp"
)

// personalProduceRequest — payload для топика hr.personal
type personalProduceRequest struct {
	MessageID  uuid.UUID `json:"message_id" example:"6b6f9c38-3e2a-4b3d-9a9a-9f1c0f8b2a10"` // Идентификатор события (UUIDv4)
	EmployeeID string    `json:"employee_id" example:"e-1024"`                              // Идентификатор сотрудника
	FirstName  *string   `json:"first_name,omitempty" example:"Анна"`                       // Имя
	LastName   *string   `json:"last_name,omitempty" example:"Иванова"`                     // Фамилия
	BirthDate  *string   `json:"birth_date,omitempty" example:"1994-06-12"`                 // Дата рождения (YYYY-MM-DD)
	Email      *string   `json:"email,omitempty" example:"anna@mail.ru"`                    // Email
	Phone      *string   `json:"phone,omitempty" example:"+7 916 123-45-67"`                // Телефон
}

// positionProduceRequest — payload для топика hr.positions
type positionProduceRequest struct {
	MessageID     uuid.UUID `json:"message_id" example:"a1d2f3c4-5678-4abc-9def-0123456789ab"` // Идентификатор события (UUIDv4)
	EmployeeID    string    `json:"employee_id" example:"e-1024"`                              // Идентификатор сотрудника
	Title         *string   `json:"title,omitempty" example:"Инженер по тестированию"`         // Должность
	Department    *string   `json:"department,omitempty" example:"Отдел качества"`             // Подразделение
	Grade         *string   `json:"grade,omitempty" example:"Middle"`                          // Грейд
	EffectiveFrom *string   `json:"effective_from,omitempty" example:"2025-10-01"`             // Дата вступления в силу (YYYY-MM-DD)
}

// historyProduceRequest — payload для топика hr.history
type historyProduceRequest struct {
	MessageID  uuid.UUID `json:"message_id" example:"0f2eb2b1-6a25-4d2a-8a7e-2c642e00e5ed"` // Идентификатор события (UUIDv4)
	EmployeeID string    `json:"employee_id" example:"e-1024"`                              // Идентификатор сотрудника
	Company    string    `json:"company" example:"ООО Ромашка"`                             // Компания
	Position   *string   `json:"position,omitempty"  example:"Инженер QA"`                  // Должность (опционально)
	PeriodFrom string    `json:"period_from" example:"2022-07-01"`                          // Начало периода (YYYY-MM-DD)
	PeriodTo   string    `json:"period_to" example:"2025-09-30"`                            // Окончание периода (YYYY-MM-DD)
	Stack      []string  `json:"stack" example:"Python,Pytest,PostgreSQL"`                  // Стек (список строк)
}

// @Summary Публикация события в hr.personal
// @Tags    Producer
// @Accept  json
// @Produce json
// @Param   request body personalProduceRequest true "payload"
// @Success 200 {object} okResponse
// @Failure 400 {object} errorResponse "Отсутствует message_id/employee_id"
// @Failure 500 {object} errorResponse "Внутренняя ошибка"
// @Router  /producer/personal [post]
func (s *Service) producerPersonal(ctx *fasthttp.RequestCtx) {
	var req personalProduceRequest

	err := json.Unmarshal(ctx.PostBody(), &req)
	if err != nil {
		writeError(ctx, fasthttp.StatusBadRequest, fmt.Errorf("json.Unmarshal: %w", err))
		return
	}

	if req.MessageID == uuid.Nil {
		writeError(ctx, fasthttp.StatusBadRequest, ErrMessageIDRequired)
		return
	}

	if strings.TrimSpace(req.EmployeeID) == "" {
		writeError(ctx, fasthttp.StatusBadRequest, ErrEmployeeIDRequired)
		return
	}

	employee := dto.EmployeeProfile{
		EmployeeID: req.EmployeeID,
		FirstName:  req.FirstName,
		LastName:   req.LastName,
		BirthDate:  req.BirthDate,
		Email:      req.Email,
		Phone:      req.Phone,
	}

	if err := s.producer.ProducePersonal(ctx, req.MessageID, employee); err != nil {
		writeError(ctx, fasthttp.StatusInternalServerError, fmt.Errorf("producer.ProducePersonal: %w", err))
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
// @Failure 400 {object} errorResponse "Отсутствует message_id/employee_id"
// @Failure 500 {object} errorResponse "Внутренняя ошибка"
// @Router  /producer/position [post]
func (s *Service) producerPosition(ctx *fasthttp.RequestCtx) {
	var req positionProduceRequest

	err := json.Unmarshal(ctx.PostBody(), &req)
	if err != nil {
		writeError(ctx, fasthttp.StatusBadRequest, fmt.Errorf("json.Unmarshal: %w", err))
		return
	}

	if req.MessageID == uuid.Nil {
		writeError(ctx, fasthttp.StatusBadRequest, ErrMessageIDRequired)
		return
	}

	if strings.TrimSpace(req.EmployeeID) == "" {
		writeError(ctx, fasthttp.StatusBadRequest, ErrEmployeeIDRequired)
		return
	}

	employee := dto.EmployeeProfile{
		EmployeeID:    req.EmployeeID,
		Title:         req.Title,
		Department:    req.Department,
		Grade:         req.Grade,
		EffectiveFrom: req.EffectiveFrom,
	}

	if err := s.producer.ProducePosition(ctx, req.MessageID, employee); err != nil {
		writeError(ctx, fasthttp.StatusInternalServerError, fmt.Errorf("producer.ProducePosition: %w", err))
		return
	}

	ok(ctx, "Событие отправлено в hr.positions")
}

// @Summary Публикация события в hr.history
// @Tags    Producer
// @Accept  json
// @Produce json
// @Param   request body historyProduceRequest true "payload"
// @Failure 400 {object} errorResponse "Отсутствует message_id/employee_id"
// @Failure 500 {object} errorResponse "Внутренняя ошибка"
// @Router  /producer/history [post]
func (s *Service) producerHistory(ctx *fasthttp.RequestCtx) {
	var req historyProduceRequest

	err := json.Unmarshal(ctx.PostBody(), &req)
	if err != nil {
		writeError(ctx, fasthttp.StatusBadRequest, fmt.Errorf("json.Unmarshal: %w", err))
		return
	}

	if req.MessageID == uuid.Nil {
		writeError(ctx, fasthttp.StatusBadRequest, ErrMessageIDRequired)
		return
	}

	if strings.TrimSpace(req.EmployeeID) == "" {
		writeError(ctx, fasthttp.StatusBadRequest, ErrEmployeeIDRequired)
		return
	}

	history := dto.EmploymentHistory{
		EmployeeID: req.EmployeeID,
		Company:    &req.Company,
		Position:   req.Position,
		PeriodFrom: &req.PeriodFrom,
		PeriodTo:   &req.PeriodTo,
		Stack:      req.Stack,
	}

	if err := s.producer.ProduceHistory(ctx, req.MessageID, history); err != nil {
		writeError(ctx, fasthttp.StatusInternalServerError, fmt.Errorf("producer.ProducePosition: %w", err))
		return
	}

	ok(ctx, "Событие отправлено в hr.history")
}

// @Summary Сырые события (эмуляция kafka_events)
// @Tags    Producer
// @Produce json
// @Success 200 {array} dto.KafkaEvent
// @Failure 500 {string} string "Внутренняя ошибка"
// @Router  /events [get]
func (s *Service) listEvents(ctx *fasthttp.RequestCtx) {
	rows, err := s.events.ListEvents(ctx)
	if err != nil {
		writeError(ctx, fasthttp.StatusInternalServerError, fmt.Errorf("events.ListEvents: %w", err))
		return
	}

	writeJSON(ctx, fasthttp.StatusOK, rows)
}

// @Summary Сообщения DLQ
// @Tags    Producer
// @Produce json
// @Success 200 {array} dto.KafkaDLQ
// @Failure 500 {string} string "Внутренняя ошибка"
// @Router  /dlq [get]
func (s *Service) listDLQ(ctx *fasthttp.RequestCtx) {
	rows, err := s.events.ListDLQ(ctx)
	if err != nil {
		writeError(ctx, fasthttp.StatusInternalServerError, fmt.Errorf("events.ListDLQ: %w", err))
		return
	}

	writeJSON(ctx, fasthttp.StatusOK, rows)
}

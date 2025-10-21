package api

import (
	"context"
	"fmt"
	"time"

	"github.com/rs/zerolog/log"

	"github.com/Artexxx/HR-Kafka-QA/internal/dto"

	"github.com/fasthttp/router"
	"github.com/google/uuid"
	"github.com/valyala/fasthttp"
)

// @title           QA Kafka — HR Profiles Trainer
// @version         1.0
// @description     Учебный стенд для QA по Kafka: отправка событий продюсером, обработка 3 консьюмерами, запись в Postgres, проверка сценариев (порядок, дубликаты, DLQ, лаги).
//
//@license.name  MIT
// @license.url   https://opensource.org/license/mit
//
// @BasePath  /
// @schemes   http
// @accept    json
// @produce   json

type EventsRepository interface {
	ExistsMessage(ctx context.Context, messageID uuid.UUID) (bool, error)
	InsertEvent(ctx context.Context, ev dto.KafkaEvent) error
	InsertDLQ(ctx context.Context, dlq dto.KafkaDLQ) error
	ListEvents(ctx context.Context) ([]dto.KafkaEvent, error)
	ListDLQ(ctx context.Context) ([]dto.KafkaDLQ, error)
	ResetAll(ctx context.Context) error
}

type ProfileRepository interface {
	Create(ctx context.Context, p dto.EmployeeProfile) error
	Update(ctx context.Context, p dto.EmployeeProfile) error
	Delete(ctx context.Context, employeeID string) error
	GetProfile(ctx context.Context, employeeID string) (*dto.EmployeeProfile, error)
	ListProfiles(ctx context.Context) ([]dto.EmployeeProfile, error)

	UpsertPersonal(ctx context.Context, p dto.EmployeeProfile) error
	UpsertPosition(ctx context.Context, p dto.EmployeeProfile) error
}

type HistoryRepository interface {
	Insert(ctx context.Context, h dto.EmploymentHistory) error
	Update(ctx context.Context, h dto.EmploymentHistory) error
	Delete(ctx context.Context, id int64) error
	ListByEmployee(ctx context.Context, employeeID string) ([]dto.EmploymentHistory, error)
	GetByID(ctx context.Context, id int64) (*dto.EmploymentHistory, error)
}

type Producer interface {
	ProducePersonal(ctx context.Context, messageID uuid.UUID, in dto.EmployeeProfile) error
	ProducePosition(ctx context.Context, messageID uuid.UUID, in dto.EmployeeProfile) error
	ProduceHistory(ctx context.Context, messageID uuid.UUID, in dto.EmploymentHistory) error
}

type ServiceDeps struct {
	Port int

	EventsRepo  EventsRepository
	ProfileRepo ProfileRepository
	HistoryRepo HistoryRepository

	Producer Producer
}

type Service struct {
	r      *router.Router
	server *fasthttp.Server
	port   int

	events   EventsRepository
	profiles ProfileRepository
	history  HistoryRepository
	producer Producer
}

func NewService(d ServiceDeps) *Service {
	rt := router.New()

	s := &Service{
		r:        rt,
		port:     d.Port,
		events:   d.EventsRepo,
		profiles: d.ProfileRepo,
		history:  d.HistoryRepo,
		producer: d.Producer,
	}

	s.mountRoutes()

	s.server = &fasthttp.Server{
		Handler:            RecoveryMiddleware(LoggingMiddleware(CORS(s.r.Handler))),
		Name:               "qa-kafka-api",
		ReadTimeout:        10 * time.Second,
		WriteTimeout:       15 * time.Second,
		MaxRequestBodySize: 2 << 20, // 2 MiB
	}

	return s
}

func (s *Service) Start(ctx context.Context) error {
	mainHandler := RecoveryMiddleware(LoggingMiddleware(CORS(s.r.Handler)))

	server := fasthttp.Server{
		Handler: mainHandler,
		Name:    "QA Kafka API",
	}

	log.Info().Int("port", s.port).Msg("Starting product API")

	emergencyShutdown := make(chan error)
	go func() {
		err := server.ListenAndServe(fmt.Sprintf(":%d", s.port))
		emergencyShutdown <- err
	}()

	select {
	case <-ctx.Done():
		return server.Shutdown()
	case e := <-emergencyShutdown:
		return e
	}
}

func (s *Service) mountRoutes() {
	s.r.POST("/producer/personal", s.producerPersonal)
	s.r.POST("/producer/position", s.producerPosition)
	s.r.POST("/producer/history", s.producerHistory)

	// Profiles
	s.r.POST("/profiles", s.createProfile)
	s.r.PUT("/profiles/{employee_id}", s.updateProfile)
	s.r.DELETE("/profiles/{employee_id}", s.deleteProfile)
	s.r.GET("/profiles", s.listProfiles)
	s.r.GET("/profiles/{employee_id}", s.getProfile)

	// History
	s.r.POST("/history", s.createHistory)
	s.r.PUT("/history/{id}", s.updateHistory)
	s.r.DELETE("/history/{id}", s.deleteHistory)
	s.r.GET("/history/{employee_id}", s.listHistoryByEmployee)

	// Events/DLQ
	s.r.GET("/events", s.listEvents)
	s.r.GET("/dlq", s.listDLQ)

	// Admin & Health
	s.r.GET("/health", s.healthHandler)
	s.r.POST("/admin/reset", s.resetHandler)
}

package producer

import (
	"time"

	"github.com/google/uuid"
)

// PersonalPayload — событие о персональных данных сотрудника
type PersonalPayload struct {
	MessageID  uuid.UUID `json:"message_id" example:"6b6f9c38-3e2a-4b3d-9a9a-9f1c0f8b2a10"` // Идентификатор события (UUID v4)
	EmployeeID string    `json:"employee_id" example:"e-1024"`                              // Внутренний идентификатор сотрудника
	FirstName  string    `json:"first_name" example:"Анна"`                                 // Имя
	LastName   string    `json:"last_name"  example:"Иванова"`                              // Фамилия
	BirthDate  string    `json:"birth_date" example:"1994-06-12"`                           // Дата рождения (YYYY-MM-DD)

	Contacts struct {
		Email string `json:"email" example:"anna.ivanova@company.ru"` // E-mail
		Phone string `json:"phone" example:"+7 916 123-45-67"`        // Телефон
	} `json:"contacts" swaggertype:"object"` // Контактные данные
}

// PositionPayload — событие о должности/позиции сотрудника
type PositionPayload struct {
	MessageID     uuid.UUID `json:"message_id"  example:"a1d2f3c4-5678-4abc-9def-0123456789ab"` // Идентификатор события (UUID v4)
	EmployeeID    string    `json:"employee_id" example:"e-1024"`                               // Внутренний идентификатор сотрудника
	Title         string    `json:"title"        example:"Инженер по тестированию"`             // Должность
	Department    string    `json:"department"   example:"Отдел качества"`                      // Подразделение/отдел
	Grade         string    `json:"grade"        example:"Middle"`                              // Грейд
	EffectiveFrom string    `json:"effective_from" example:"2025-10-01"`                        // Дата вступления в силу (YYYY-MM-DD)
}

// HistoryPayload — событие об изменении/добавлении записи в историю работы
type HistoryPayload struct {
	MessageID  uuid.UUID `json:"message_id" example:"0f2eb2b1-6a25-4d2a-8a7e-2c642e00e5ed"` // Идентификатор события (UUID v4)
	EmployeeID string    `json:"employee_id" example:"e-1024"`                              // Внутренний идентификатор сотрудника
	Company    string    `json:"company"     example:"ООО Ромашка"`                         // Компания (РФ)
	Position   string    `json:"position"    example:"Инженер QA"`                          // Должность
	Period     struct {
		From string `json:"from" example:"2022-07-01"` // Дата начала (YYYY-MM-DD)
		To   string `json:"to"   example:"2025-09-30"` // Дата окончания (YYYY-MM-DD)
	} `json:"period" swaggertype:"object"` // Период работы
	Stack []string `json:"stack" example:"Python,Pytest,PostgreSQL"` // Технологический стек
}

type Envelope[T any] struct {
	Kind       string    `json:"kind"        example:"personal"`                             // Тип события
	MessageID  uuid.UUID `json:"message_id"  example:"c7e06db5-4b71-4c54-9334-3f9a6e6c5d0e"` // Идентификатор события (UUID v4)
	EmployeeID string    `json:"employee_id" example:"e-1024"`                               // Внутренний идентификатор сотрудника
	Payload    T         `json:"payload"`                                                    // Полезная нагрузка (структура зависит от kind)
	Timestamp  time.Time `json:"timestamp"   example:"2025-10-19T12:34:56Z"`                 // Время формирования события
	Source     string    `json:"source"      example:"hr-api-gateway"`                       // Сервис-источник
}

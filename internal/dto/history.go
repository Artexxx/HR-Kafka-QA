package dto

import (
	"time"
)

// EmploymentHistory — запись истории работы сотрудника.
type EmploymentHistory struct {
	ID         int64     `json:"id" example:"42"`                           // Идентификатор записи (БД)
	EmployeeID string    `json:"employee_id" example:"e-1024"`              // Идентификатор сотрудника
	Company    string    `json:"company" example:"ООО Ромашка"`             // Компания
	Position   *string   `json:"position,omitempty" example:"Инженер QA"`   // Должность
	PeriodFrom string    `json:"period_from" example:"2022-07-01"`          // Дата начала периода занятости (YYYY-MM-DD)
	PeriodTo   string    `json:"period_to" example:"2025-09-30"`            // Дата окончания периода занятости (YYYY-MM-DD)
	Stack      []string  `json:"stack" example:"Python,Pytest,PostgreSQL"`  // Технологический стек (список строк)
	CreatedAt  time.Time `json:"created_at" example:"2025-10-19T10:15:30Z"` // Время создания записи
}

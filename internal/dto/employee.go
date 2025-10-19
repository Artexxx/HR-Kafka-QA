package dto

import (
	"time"
)

// EmployeeProfile — данные профиля сотрудника
type EmployeeProfile struct {
	EmployeeID    string    `json:"employee_id" example:"e-1024"`                      // Идентификатор сотрудника
	FirstName     *string   `json:"first_name" example:"Анна"`                         // Имя сотрудника
	LastName      *string   `json:"last_name,omitempty" example:"Иванова"`             // Фамилия сотрудника
	BirthDate     *string   `json:"birth_date,omitempty" example:"1994-06-12"`         // Дата рождения в формате YYYY-MM-DD
	Email         *string   `json:"email,omitempty" example:"anna@mail.ru"`            // Корпоративная/личная почта сотрудника
	Phone         *string   `json:"phone,omitempty" example:"+7 916 123-45-67"`        // Телефон
	Title         *string   `json:"title,omitempty" example:"Инженер по тестированию"` // Должность
	Department    *string   `json:"department,omitempty" example:"Отдел качества"`     // Подразделение/отдел
	Grade         *string   `json:"grade,omitempty" example:"Middle"`                  // Грейд (Junior, Middle, Senior или другой)
	EffectiveFrom *string   `json:"effective_from,omitempty" example:"2025-10-01"`     // Дата вступления изменений в силу (YYYY-MM-DD)
	UpdatedAt     time.Time `json:"updated_at"`
}

package producer

// PersonalPayload — событие о персональных данных сотрудника
type PersonalPayload struct {
	EmployeeID string `json:"employee_id" example:"e-1024"`    // Внутренний идентификатор сотрудника
	FirstName  string `json:"first_name" example:"Анна"`       // Имя
	LastName   string `json:"last_name"  example:"Иванова"`    // Фамилия
	BirthDate  string `json:"birth_date" example:"1994-06-12"` // Дата рождения (YYYY-MM-DD)

	Contacts struct {
		Email string `json:"email" example:"anna.ivanova@company.ru"` // E-mail
		Phone string `json:"phone" example:"+7 916 123-45-67"`        // Телефон
	} `json:"contacts" swaggertype:"object"` // Контактные данные
}

// PositionPayload — событие о должности/позиции сотрудника
type PositionPayload struct {
	EmployeeID    string `json:"employee_id" example:"e-1024"`                   // Внутренний идентификатор сотрудника
	Title         string `json:"title"        example:"Инженер по тестированию"` // Должность
	Department    string `json:"department"   example:"Отдел качества"`          // Подразделение/отдел
	Grade         string `json:"grade"        example:"Middle"`                  // Грейд
	EffectiveFrom string `json:"effective_from" example:"2025-10-01"`            // Дата вступления в силу (YYYY-MM-DD)
}

// HistoryPayload — событие об изменении/добавлении записи в историю работы
type HistoryPayload struct {
	EmployeeID string `json:"employee_id" example:"e-1024"`      // Внутренний идентификатор сотрудника
	Company    string `json:"company"     example:"ООО Ромашка"` // Компания (РФ)
	Position   string `json:"position"    example:"Инженер QA"`  // Должность
	Period     struct {
		From string `json:"from" example:"2022-07-01"` // Дата начала (YYYY-MM-DD)
		To   string `json:"to"   example:"2025-09-30"` // Дата окончания (YYYY-MM-DD)
	} `json:"period" swaggertype:"object"` // Период работы
	Stack []string `json:"stack" example:"Python,Pytest,PostgreSQL"` // Технологический стек
}

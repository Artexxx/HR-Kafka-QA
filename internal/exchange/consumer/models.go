package consumer

type PersonalPayload struct {
	EmployeeID string `json:"employee_id"`
	FirstName  string `json:"first_name"`
	LastName   string `json:"last_name"`
	BirthDate  string `json:"birth_date"`
	Contacts   struct {
		Email string `json:"email"`
		Phone string `json:"phone"`
	} `json:"contacts"`
}

type PositionPayload struct {
	EmployeeID    string `json:"employee_id"`
	Title         string `json:"title"`
	Department    string `json:"department"`
	Grade         string `json:"grade"`
	EffectiveFrom string `json:"effective_from"`
}

type HistoryPayload struct {
	EmployeeID string `json:"employee_id"`
	Company    string `json:"company"`
	Position   string `json:"position"`
	Period     struct {
		From string `json:"from"`
		To   string `json:"to"`
	} `json:"period"`
	Stack []string `json:"stack"`
}

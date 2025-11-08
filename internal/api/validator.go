package api

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/Artexxx/HR-Kafka-QA/internal/dto"
)

var regexDate = regexp.MustCompile(`^\d{4}-\d{2}-\d{2}$`)
var allowedGrades = map[string]struct{}{
	"Junior": {}, "Middle": {}, "Senior": {}, "Lead": {}, "Head": {},
}

func validDate(s string) bool {
	_, err := time.Parse("2006-01-02", s)
	return err == nil
}

func isEmptyPtr(p *string) bool {
	return p == nil || strings.TrimSpace(*p) == ""
}

func checkDate(field, value string) string {
	if !regexDate.MatchString(value) || !validDate(value) {
		return fmt.Sprintf("invalid value in field '%s'=%s", field, value)
	}

	return ""
}

func checkEmail(field, value string) string {
	if !strings.Contains(value, "@") {
		return fmt.Sprintf("invalid value in field '%s'=%s", field, value)
	}

	return ""
}

func validateEmployeeProfile(p dto.EmployeeProfile) string {
	if strings.TrimSpace(p.EmployeeID) == "" {
		return "required field 'employee_id'"
	}

	if strings.TrimSpace(p.FirstName) == "" {
		return "required field 'first_name'"
	}

	if strings.TrimSpace(p.BirthDate) == "" {
		return "required field 'birth_date'"
	}

	if msg := checkDate("birth_date", strings.TrimSpace(p.BirthDate)); msg != "" {
		return msg
	}

	if strings.TrimSpace(p.Email) == "" {
		return "required field 'email'"
	}

	if msg := checkEmail("email", strings.TrimSpace(p.Email)); msg != "" {
		return msg
	}

	if strings.TrimSpace(p.Phone) == "" {
		return "required field 'phone'"
	}

	if p.Title != nil && strings.TrimSpace(*p.Title) == "" {
		return "required field 'title'"
	}

	if p.Department != nil && strings.TrimSpace(*p.Department) == "" {
		return "required field 'department'"
	}

	if p.Grade != nil {
		val := strings.TrimSpace(*p.Grade)

		if val == "" {
			return "required field 'grade'"
		}

		if _, ok := allowedGrades[strings.TrimSpace(*p.Grade)]; !ok {
			return fmt.Sprintf("invalid enum value: grade %s not in allowed grades %v", *p.Grade, allowedGrades)
		}
	}

	if p.EffectiveFrom != nil {
		val := strings.TrimSpace(*p.EffectiveFrom)

		if val == "" {
			return "required field 'effective_from'"
		}

		if msg := checkDate("effective_from", val); msg != "" {
			return msg
		}
	}

	return ""
}

func validateEmploymentHistory(history dto.EmploymentHistory) string {
	if strings.TrimSpace(history.EmployeeID) == "" {
		return "required field 'employee_id'"
	}

	if strings.TrimSpace(history.Company) == "" {
		return "required field 'company'"
	}

	if strings.TrimSpace(history.PeriodFrom) == "" {
		return "required field 'period_from'"
	}

	if strings.TrimSpace(history.PeriodTo) == "" {
		return "required field 'period_to'"
	}

	if msg := checkDate("period_from", strings.TrimSpace(history.PeriodFrom)); msg != "" {
		return msg
	}

	if msg := checkDate("period_to", strings.TrimSpace(history.PeriodTo)); msg != "" {
		return msg
	}

	fromT, _ := time.Parse("2006-01-02", strings.TrimSpace(history.PeriodFrom))
	toT, _ := time.Parse("2006-01-02", strings.TrimSpace(history.PeriodTo))
	if toT.Before(fromT) {
		return fmt.Sprintf("invalid value in field 'period'={from:%s to:%s}", history.PeriodFrom, history.PeriodTo)
	}

	return ""
}

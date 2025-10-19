package consumer

import (
	"fmt"
	"regexp"
	"strings"
	"time"
)

var regexDate = regexp.MustCompile(`^\d{4}-\d{2}-\d{2}$`)
var allowedGrades = map[string]struct{}{
	"Junior": {}, "Middle": {}, "Senior": {}, "Lead": {}, "Head": {},
}

func validatePersonal(payload PersonalPayload) string {
	if strings.TrimSpace(payload.FirstName) == "" {
		return "required field 'first_name'"
	}

	if strings.TrimSpace(payload.BirthDate) == "" {
		return "required field 'birth_date'"
	}

	if !regexDate.MatchString(payload.BirthDate) || !validDate(payload.BirthDate) {
		return fmt.Sprintf("ivalid value in field 'birth_date'=%s", payload.BirthDate)
	}

	if strings.TrimSpace(payload.Contacts.Email) == "" {
		return "required field 'contacts.email'"
	}

	if !strings.Contains(payload.Contacts.Email, "@") {
		return fmt.Sprintf("ivalid value in field 'email'=%s", payload.Contacts.Email)
	}

	if strings.TrimSpace(payload.Contacts.Phone) == "" {
		return "required field 'contacts.phone'"
	}

	return ""
}

func validatePosition(payload PositionPayload) string {
	if strings.TrimSpace(payload.Title) == "" {
		return "required field 'title'"
	}

	if strings.TrimSpace(payload.Department) == "" {
		return "required field 'department'"
	}

	if payload.Grade != "" {
		if _, ok := allowedGrades[payload.Grade]; !ok {
			return fmt.Sprintf("invalid enum value: grade %s not in allowed grades %v", payload.Grade, allowedGrades)
		}
	}

	if strings.TrimSpace(payload.EffectiveFrom) == "" {
		return "required field 'effective_from'"
	}

	if !regexDate.MatchString(payload.EffectiveFrom) || !validDate(payload.EffectiveFrom) {
		return fmt.Sprintf("ivalid value in field 'date'=%s", payload.EffectiveFrom)
	}

	return ""
}

func validateHistory(payload HistoryPayload) string {
	if strings.TrimSpace(payload.Company) == "" {
		return "required field 'company'"
	}

	if strings.TrimSpace(payload.Period.From) == "" {
		return "required field 'period.to'"
	}

	if strings.TrimSpace(payload.Period.To) == "" {
		return "required field 'period.to'"
	}

	if !regexDate.MatchString(payload.Period.From) || !validDate(payload.Period.From) {
		return fmt.Sprintf("ivalid value in field 'period.from'=%s", payload.Period.From)
	}

	if !regexDate.MatchString(payload.Period.To) || !validDate(payload.Period.To) {
		return fmt.Sprintf("ivalid value in field 'period.to'=%s", payload.Period.To)
	}

	fromT, _ := time.Parse("2006-01-02", payload.Period.From)
	toT, _ := time.Parse("2006-01-02", payload.Period.To)
	if toT.Before(fromT) {
		return fmt.Sprintf("ivalid value in field 'period'=%s", payload.Period)
	}

	return ""
}

func validDate(s string) bool {
	_, err := time.Parse("2006-01-02", s)
	return err == nil
}

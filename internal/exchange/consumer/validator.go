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

func validatePersonal(p PersonalPayload) string {
	if strings.TrimSpace(p.FirstName) == "" || strings.TrimSpace(p.LastName) == "" {
		return fmt.Sprintf("invalid name: fist_name=%s last_name=%s", p.FirstName, p.LastName)
	}

	if !regexDate.MatchString(p.BirthDate) || !validDate(p.BirthDate) {
		return fmt.Sprintf("invalid date format: birth_date=%s", p.BirthDate)
	}

	if p.Contacts.Email != "" && !strings.Contains(p.Contacts.Email, "@") {
		return fmt.Sprintf("invalid email: email=%s", p.Contacts.Email)
	}

	return ""
}

func validatePosition(p PositionPayload) string {
	if strings.TrimSpace(p.Title) == "" {
		return "missing position title"
	}

	if strings.TrimSpace(p.Department) == "" {
		return "missing department"

	}
	if p.Grade != "" {
		if _, ok := allowedGrades[p.Grade]; !ok {
			return fmt.Sprintf("invalid enum value: grade %s not in allowed grades %v", p.Grade, allowedGrades)
		}
	}
	if !regexDate.MatchString(p.EffectiveFrom) || !validDate(p.EffectiveFrom) {
		return fmt.Sprintf("invalid date: date=%s", p.EffectiveFrom)
	}
	return ""
}

func validateHistory(p HistoryPayload) string {
	if strings.TrimSpace(p.Company) == "" {
		return "missing required field company"
	}
	if p.Period.From == "" || p.Period.To == "" {
		return "missing required field period"
	}
	if !regexDate.MatchString(p.Period.From) || !regexDate.MatchString(p.Period.To) || !validDate(p.Period.From) || !validDate(p.Period.To) {
		return fmt.Sprintf("invalid date: period=%s", p.Period)
	}

	fromT, _ := time.Parse("2006-01-02", p.Period.From)
	toT, _ := time.Parse("2006-01-02", p.Period.To)
	if toT.Before(fromT) {
		return fmt.Sprintf("invalid period: period=%s", p.Period)
	}

	return ""
}

func validDate(s string) bool {
	_, err := time.Parse("2006-01-02", s)
	return err == nil
}

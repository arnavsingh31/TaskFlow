package helpers

import (
	"net/mail"
)

func ValidStatus(s *string) bool {
	if s == nil {
		return true
	}
	switch *s {
	case "todo", "in_progress", "done":
		return true
	}
	return false
}

func ValidPriority(s *string) bool {
	if s == nil {
		return true
	}
	switch *s {
	case "low", "medium", "high":
		return true
	}
	return false
}

func ValidEmail(email *string) bool {
	if email == nil {
		return false
	}
	_, err := mail.ParseAddress(*email)
	return err == nil
}

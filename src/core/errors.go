package core

import "fmt"

type UserError struct {
	Msg string
}

func (e UserError) Error() string { return e.Msg }

func NewUserError(format string, args ...any) error {
	return UserError{Msg: fmt.Sprintf(format, args...)}
}

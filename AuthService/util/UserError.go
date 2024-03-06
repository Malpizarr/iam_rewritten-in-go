package util

type UserError struct {
	Username string
	Msg      string
}

func (e *UserError) Error() string {
	return e.Msg
}

func NewUserError(username, msg string) error {
	return &UserError{
		Username: username,
		Msg:      msg,
	}
}

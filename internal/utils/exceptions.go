package utils

type Exception interface {
	Error() string
	Message() string
}

type exceptions struct {
	err error
	msg string
	val interface{}
}

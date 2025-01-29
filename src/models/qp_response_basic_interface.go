package models

type QpResponseBasicInterface interface {
	IsSuccess() bool
	GetStatusMessage() string
}

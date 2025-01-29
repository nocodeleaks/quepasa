package models

type QpResponseInterface interface {
	QpResponseBasicInterface
	ParseSuccess(string)
}

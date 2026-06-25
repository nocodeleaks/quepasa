package models

// Interface para recuperar o estado do objeto em questão
type QPIStateRecovery interface {
	GetState() (int, string)
}

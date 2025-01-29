package models

type QPBot struct {
	QpServer

	UserID string `db:"user_id" json:"user_id"`
}

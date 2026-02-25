package models

import (
	whatsapp "github.com/nocodeleaks/quepasa/whatsapp"
)

type QpReceiveResponse struct {
	QpResponse
	Total    uint64                     `json:"total"`
	Page     int                        `json:"page,omitempty"`
	Limit    int                        `json:"limit,omitempty"`
	TotalPages int                      `json:"total_pages,omitempty"`
	Messages []whatsapp.WhatsappMessage `json:"messages,omitempty"`
	Server   *QpServer                  `json:"server,omitempty"`
}

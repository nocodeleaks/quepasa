package models

import (
	"sync"

	whatsapp "github.com/nocodeleaks/quepasa/whatsapp"
)

type QpWhatsappServer struct {
	*QpServer
	QpDataDispatching // new dispatching system

	// should auto reconnect, false for qrcode scanner
	Reconnect bool `json:"reconnect"`

	connection     whatsapp.IWhatsappConnection `json:"-"`
	syncConnection *sync.Mutex                  `json:"-"` // Objeto de sinaleiro para evitar chamadas simultâneas a este objeto
	syncMessages   *sync.Mutex                  `json:"-"` // Objeto de sinaleiro para evitar chamadas simultâneas a este objeto

	Timestamps QpTimestamps `json:"timestamps"`

	Handler        *DispatchingHandler `json:"-"`
	GroupManager   *QpGroupManager     `json:"-"` // composition for group operations
	StatusManager  *QpStatusManager    `json:"-"` // composition for status operations
	ContactManager *QpContactManager   `json:"-"` // composition for contact operations

	// Intent tracks the current application-level lifecycle request for this session.
	// Use IsStopRequested() / IsDeleteRequested() instead of reading the field directly.
	Intent SessionIntent          `json:"-"`
	db     QpDataServersInterface `json:"-"`
}

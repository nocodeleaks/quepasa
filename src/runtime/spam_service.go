package runtime

import (
	"fmt"

	"github.com/nocodeleaks/quepasa/models"
	"github.com/nocodeleaks/quepasa/whatsapp"
)

// GetSpamSession returns the section that should send a /spam request.
// When spam_sections has configured rows, only those rows are eligible and the
// stored order is respected. When the table is empty, the legacy behavior is
// preserved by returning the first ready live session.
func GetSpamSession() (*models.QpWhatsappSession, error) {
	db := models.GetDatabase()
	if db == nil || db.SpamSections == nil {
		return GetFirstReadySession()
	}

	sections, err := db.SpamSections.ListAll()
	if err != nil {
		return nil, err
	}

	if len(sections) == 0 {
		return GetFirstReadySession()
	}

	for _, item := range sections {
		if item == nil || !item.Enabled {
			continue
		}

		session, ok := FindLiveSessionByToken(item.Token)
		if !ok || session == nil {
			continue
		}
		if session.GetStatus() == whatsapp.Ready {
			return session, nil
		}
	}

	return nil, fmt.Errorf("no configured spam section is ready")
}

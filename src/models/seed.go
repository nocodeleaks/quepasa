package models

import (
	"fmt"
	"os"
	"strings"

	log "github.com/sirupsen/logrus"
)

const DEFAULTEMAIL string = "default@quepasa.io"

func InitialSeed() (err error) {
    // Ensure DB services are initialized
    if WhatsappService == nil || WhatsappService.DB == nil || WhatsappService.DB.Users == nil {
        return fmt.Errorf("WhatsappService or DB.Users not initialized")
    }

    // Read env vars and apply fallbacks
    envEMAIL := strings.TrimSpace(os.Getenv("QUEPASA_BASIC_AUTH_USER"))
    envPASSWORD := os.Getenv("QUEPASA_BASIC_AUTH_PASSWORD")

    if envEMAIL != "" {
		// Check if user exists
		exists2, err := WhatsappService.DB.Users.Exists(envEMAIL)
		if err != nil {
			return fmt.Errorf("failed to check existence of user '%s': %w", envEMAIL, err)
		}

		if !exists2 {
			if envPASSWORD == "" {
				log.Warn("QUEPASA_BASIC_AUTH_PASSWORD not set; creating user with empty password")
			}
			_, err = WhatsappService.DB.Users.Create(envEMAIL, envPASSWORD)
			if err != nil {
				return fmt.Errorf("failed to create seed user '%s': %w", envEMAIL, err)
			}
			log.Infof("Created seed user: %s", envEMAIL)
		}
    }

    return
}
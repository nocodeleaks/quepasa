package models

import (
	"fmt"

	environment "github.com/nocodeleaks/quepasa/environment"
	log "github.com/sirupsen/logrus"
)

const DEFAULTEMAIL string = "default@quepasa.io"

func InitialSeed() (err error) {
	// Ensure DB services are initialized
	if WhatsappService == nil || WhatsappService.DB == nil || WhatsappService.DB.Users == nil {
		return fmt.Errorf("WhatsappService or DB.Users not initialized")
	}

	// Read environment variables for default user from environment module
	envEMAIL := environment.Settings.API.User
	envPASSWORD := environment.Settings.API.Password

	// Log loaded values for debugging
	log.Debugf("InitialSeed: Loaded USER='%s' PASSWORD='%s' (length: %d)",
		envEMAIL,
		func() string {
			if envPASSWORD != "" {
				return "***"
			} else {
				return ""
			}
		}(),
		len(envPASSWORD))

	// Use environment variable if set, otherwise fallback to DEFAULTEMAIL
	if envEMAIL != "" {
		// Check if user exists
		exists, err := WhatsappService.DB.Users.Exists(envEMAIL)
		if err != nil {
			return fmt.Errorf("failed to check if user exists: %w", err)
		}

		if !exists {
			// Validate password is not empty for security
			if envPASSWORD == "" {
				return fmt.Errorf("PASSWORD not set; refusing to create user '%s' with empty password", envEMAIL)
			}

			log.Infof("Creating default user from environment: %s", envEMAIL)
			_, err = WhatsappService.DB.Users.Create(envEMAIL, envPASSWORD)
			if err != nil {
				return fmt.Errorf("failed to create user '%s': %w", envEMAIL, err)
			}
			log.Infof("Successfully created default user: %s", envEMAIL)
		} else {
			log.Infof("User '%s' already exists, skipping creation", envEMAIL)
		}
	} else {
		// Fallback to default email with empty password (legacy behavior)
		log.Warnf("USER not set, using default: %s", DEFAULTEMAIL)
		exists, err := WhatsappService.DB.Users.Exists(DEFAULTEMAIL)
		if err != nil {
			return fmt.Errorf("failed to check if default user exists: %w", err)
		}

		if !exists {
			log.Warn("Creating default user with EMPTY password - CHANGE THIS IMMEDIATELY!")
			_, err = WhatsappService.DB.Users.Create(DEFAULTEMAIL, "")
			if err != nil {
				return fmt.Errorf("failed to create default user: %w", err)
			}
		}
	}

	return nil
}

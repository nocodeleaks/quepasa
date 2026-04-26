package models

import (
	"context"
	"fmt"

	library "github.com/nocodeleaks/quepasa/library"
	whatsmeow "github.com/nocodeleaks/quepasa/whatsmeow"
)

// WhatsmeowOrphanDevice represents a device registered in the whatsmeow database
// that does not have a matching active server entry in the QuePasa database.
//
// An "orphan" occurs when:
//   - The QuePasa server record was deleted while the whatsmeow session remained.
//   - The instance was migrated from another machine without copying quepasa.sqlite.
//   - A crash or hard reset cleared quepasa.sqlite but left whatsmeow.sqlite intact.
//
// All fields are read from whatsmeow_device via the whatsmeow container and
// are never persisted directly in quepasa.sqlite by this struct.
type WhatsmeowOrphanDevice struct {
	// JID is the full WhatsApp identifier of the device session,
	// e.g. "553176011595:18@s.whatsapp.net".
	JID string `json:"jid"`

	// Phone is the international phone number extracted from the JID,
	// e.g. "553176011595". Useful for matching against existing tokens.
	Phone string `json:"phone"`

	// PushName is the display name the user set on WhatsApp (may be empty).
	PushName string `json:"push_name,omitempty"`

	// Platform is the device platform string reported by WhatsApp,
	// e.g. "android", "ios". May be empty for desktop sessions.
	Platform string `json:"platform,omitempty"`
}

// RestoreReport summarises the result of a restore operation, regardless of
// whether it was a dry-run diagnostic or an actual restore attempt.
type RestoreReport struct {
	// OrphanDevices holds whatsmeow device sessions that have no matching
	// QuePasa server with a non-empty wid pointing to them.
	OrphanDevices []WhatsmeowOrphanDevice `json:"orphan_devices"`

	// UnlinkedServers holds QuePasa server tokens whose wid field is empty,
	// meaning the session has never been paired or was cleared.
	UnlinkedServers []string `json:"unlinked_servers"`

	// Restored contains the token→JID pairs that were successfully linked
	// during an actual restore (empty during a dry-run).
	Restored []RestoreResult `json:"restored,omitempty"`

	// Errors contains token→error pairs for any linking attempt that failed.
	Errors []RestoreError `json:"errors,omitempty"`
}

// RestoreResult records a successful token↔JID link performed by RestoreOrphaned.
type RestoreResult struct {
	// Token is the QuePasa server token that was linked.
	Token string `json:"token"`

	// JID is the whatsmeow device JID that was associated with the token.
	JID string `json:"jid"`
}

// RestoreError records a failed restore attempt with the reason.
type RestoreError struct {
	Token  string `json:"token,omitempty"`
	JID    string `json:"jid,omitempty"`
	Reason string `json:"reason"`
}

// DiagnoseOrphaned queries both the whatsmeow and QuePasa databases to identify:
//  1. Whatsmeow device sessions that have no matching QuePasa server (orphan devices).
//  2. QuePasa server records that have no wid set (unlinked servers).
//
// This method is read-only — it never modifies any data. Use it to inspect the
// current state before deciding to call RestoreOrphaned.
//
// Returns a RestoreReport with OrphanDevices and UnlinkedServers populated.
// Returns an error only when the whatsmeow container cannot be read.
func (source *QPWhatsappService) DiagnoseOrphaned() (*RestoreReport, error) {
	logentry := source.GetLogger()
	logentry.Info("starting orphaned account diagnosis")

	report := &RestoreReport{
		OrphanDevices:   make([]WhatsmeowOrphanDevice, 0),
		UnlinkedServers: make([]string, 0),
	}

	// ── Step 1: Read all device sessions from whatsmeow ──────────────────────
	// The whatsmeow container holds all WhatsApp session keys. Each entry in
	// whatsmeow_device corresponds to one phone-number+device-slot pair that
	// was once successfully paired via QR code or pair code.
	wmService := whatsmeow.WhatsmeowService
	if wmService == nil {
		return nil, fmt.Errorf("whatsmeow service is not initialised; cannot diagnose orphans")
	}

	wmDevices, err := wmService.Container.GetAllDevices(context.TODO())
	if err != nil {
		return nil, fmt.Errorf("failed to list whatsmeow devices: %w", err)
	}
	logentry.Infof("found %d device(s) in whatsmeow database", len(wmDevices))

	// Build a set of wids that are already claimed by a QuePasa server so that
	// the inner loop below runs in O(n) instead of O(n²).
	dbServers := source.DB.Servers.FindAll()
	claimedWids := make(map[string]struct{}, len(dbServers))
	for _, srv := range dbServers {
		wid := srv.GetWId()
		if len(wid) > 0 {
			claimedWids[wid] = struct{}{}
		} else {
			// Servers without a wid are candidates for re-linking.
			report.UnlinkedServers = append(report.UnlinkedServers, srv.Token)
		}
	}
	logentry.Infof("found %d unlinked server(s) in quepasa database", len(report.UnlinkedServers))

	// ── Step 2: Find whatsmeow devices not claimed by any server ─────────────
	for _, dev := range wmDevices {
		if dev == nil || dev.ID == nil {
			continue
		}
		jid := dev.ID.String()
		if _, claimed := claimedWids[jid]; claimed {
			// This device is already linked to a QuePasa server — skip it.
			continue
		}

		phone := library.GetPhoneByWId(jid)
		orphan := WhatsmeowOrphanDevice{
			JID:      jid,
			Phone:    phone,
			PushName: dev.PushName,
			Platform: dev.Platform,
		}
		report.OrphanDevices = append(report.OrphanDevices, orphan)
		logentry.Infof("orphan device found: jid=%s phone=%s push_name=%s", jid, phone, dev.PushName)
	}

	logentry.Infof("diagnosis complete: %d orphan device(s), %d unlinked server(s)",
		len(report.OrphanDevices), len(report.UnlinkedServers))
	return report, nil
}

// RestoreOrphaned attempts to automatically re-link orphaned whatsmeow device
// sessions to unlinked QuePasa servers by matching phone numbers.
//
// Matching strategy:
//  1. If there is exactly one orphan device and exactly one unlinked server,
//     they are linked automatically.
//  2. Any ambiguous cases (multiple orphan devices and/or multiple unlinked
//     servers) are listed in the report but left unchanged — never force-linked.
//
// After linking, each matched server is saved to the database and re-initialised
// in the runtime cache so it can immediately resume receiving messages.
//
// Returns a RestoreReport with the Restored and Errors fields populated.
// Always returns a report even when errors occurred — partial restores are
// preserved so the caller can inspect what worked and retry failures manually.
func (source *QPWhatsappService) RestoreOrphaned() (*RestoreReport, error) {
	logentry := source.GetLogger()
	logentry.Info("starting orphaned account restore")

	// Obtain current diagnosis first so we know what needs to be fixed.
	report, err := source.DiagnoseOrphaned()
	if err != nil {
		return nil, err
	}

	report.Restored = make([]RestoreResult, 0)
	report.Errors = make([]RestoreError, 0)

	// Nothing to do when either list is empty.
	if len(report.OrphanDevices) == 0 {
		logentry.Info("no orphan devices found; nothing to restore")
		return report, nil
	}
	if len(report.UnlinkedServers) == 0 {
		logentry.Info("no unlinked servers found; nothing to restore")
		return report, nil
	}

	// When exactly one orphan device and one unlinked server exist, the
	// association is unambiguous and we can apply it automatically.
	if len(report.OrphanDevices) == 1 && len(report.UnlinkedServers) == 1 {
		orphan := report.OrphanDevices[0]
		token := report.UnlinkedServers[0]
		if restErr := source.linkAndRestore(token, orphan.JID); restErr != nil {
			logentry.Errorf("one-to-one restore failed token=%s jid=%s: %v", token, orphan.JID, restErr)
			report.Errors = append(report.Errors, RestoreError{Token: token, JID: orphan.JID, Reason: restErr.Error()})
		} else {
			logentry.Infof("restored via one-to-one fallback: token=%s jid=%s", token, orphan.JID)
			report.Restored = append(report.Restored, RestoreResult{Token: token, JID: orphan.JID})
		}
	} else {
		for _, orphan := range report.OrphanDevices {
			logentry.Warnf("ambiguous orphan not auto-linked: jid=%s phone=%s", orphan.JID, orphan.Phone)
		}
	}

	logentry.Infof("restore complete: %d restored, %d errors", len(report.Restored), len(report.Errors))
	return report, nil
}

// RestoreManual performs a direct token↔JID link without any automatic
// matching heuristics. Use this when the automatic restore cannot disambiguate
// and you know exactly which server token should be paired to which whatsmeow
// device JID.
//
// The token must already exist in the QuePasa database. If the token does not
// exist, an error is returned and no changes are made.
//
// The JID must correspond to an active device session in whatsmeow.sqlite.
// If the JID is not found, an error is returned and the server is left unchanged.
//
// On success the server record is saved and the runtime cache is updated, so
// the restored connection is immediately usable without a service restart.
func (source *QPWhatsappService) RestoreManual(token string, jid string) error {
	logentry := source.GetLogger()
	logentry.Infof("manual restore requested: token=%s jid=%s", token, jid)

	// Validate that the token exists in the QuePasa database before touching anything.
	exists, err := source.DB.Servers.Exists(token)
	if err != nil {
		return fmt.Errorf("database error while checking token %s: %w", token, err)
	}
	if !exists {
		return fmt.Errorf("token %s not found in quepasa database", token)
	}

	// Validate that the JID is registered in whatsmeow.
	if err := source.validateWhatsmeowJID(jid); err != nil {
		return fmt.Errorf("whatsmeow validation failed for jid %s: %w", jid, err)
	}

	return source.linkAndRestore(token, jid)
}

// ── Private helpers ───────────────────────────────────────────────────────────

// linkAndRestore updates the server record with the given wid, persists the
// change to the database, and re-initialises the server in the runtime cache.
//
// This is the single point of mutation used by all restore paths so the logic
// stays consistent regardless of how the token↔JID pair was found.
func (source *QPWhatsappService) linkAndRestore(token string, jid string) error {
	logentry := source.GetLogger()
	logentry.Infof("linking token=%s to jid=%s", token, jid)

	// Load the server info from the database to get the full record
	// (including groups, broadcasts, user, etc.) so we do not overwrite those
	// fields with empty defaults.
	info, err := source.DB.Servers.FindByToken(token)
	if err != nil {
		return fmt.Errorf("failed to load server %s from database: %w", token, err)
	}

	// Persist the newly discovered JID into the server record.
	// This is the primary change that fixes the "NULL wid" problem described
	// in the session history: sqlx.Select cannot scan NULL into a plain string
	// field, so any server that never had a wid set fails to load on restart.
	info.SetWId(jid)
	info.Verified = true // mark as verified since we found a live session

	// Persist to the database first so the change survives a restart even if
	// the in-memory cache update below encounters a transient error.
	if err := source.DB.Servers.Update(info); err != nil {
		return fmt.Errorf("failed to save restored server %s: %w", token, err)
	}
	logentry.Infof("server %s saved with wid=%s", token, jid)

	// Update or add the server in the runtime cache so messages can flow
	// immediately without waiting for a full service restart.
	server, err := source.AppendNewServer(info)
	if err != nil {
		// Cache update failure is not fatal — the DB record is already correct.
		// Log the error but do not return it so the caller sees a partial success
		// and can inspect the runtime state separately.
		logentry.Warnf("server %s saved to DB but cache update failed: %v", token, err)
		return nil
	}

	// Trigger an async connection attempt so the server starts receiving
	// messages right after restore without needing a manual restart.
	go server.Initialize()

	logentry.Infof("server %s successfully restored and initialised", token)
	return nil
}

// validateWhatsmeowJID confirms that the given JID is present in the whatsmeow
// device table. Returns an error if the JID is not found or if the whatsmeow
// service is unavailable.
//
// This check prevents linking a server to a non-existent session, which would
// result in a broken connection that can never authenticate.
func (source *QPWhatsappService) validateWhatsmeowJID(jid string) error {
	wmService := whatsmeow.WhatsmeowService
	if wmService == nil {
		return fmt.Errorf("whatsmeow service is not initialised")
	}

	devices, err := wmService.Container.GetAllDevices(context.TODO())
	if err != nil {
		return fmt.Errorf("failed to query whatsmeow devices: %w", err)
	}

	for _, dev := range devices {
		if dev != nil && dev.ID != nil && dev.ID.String() == jid {
			return nil // JID found and valid
		}
	}

	return fmt.Errorf("jid %s not found in whatsmeow database", jid)
}

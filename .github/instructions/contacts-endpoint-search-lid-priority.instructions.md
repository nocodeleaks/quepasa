# Contacts Endpoint Instruction

## Scope
- API scope: contacts listing and search behavior.
- Main files:
  - `src/api/api_handlers+ContactsController.go`
  - `src/api/api_handlers+ContactSearchController.go`
  - `src/whatsmeow/whatsmeow_contact_manager.go`

## Data Source Rules
- Contacts must come from local whatsmeow contact store/cache.
- Do not assume direct online fetch from WhatsApp for listing.
- Preserve merge behavior between `@lid` and `@s.whatsapp.net` identities.

## LID Rules
- LID is opaque and privacy-oriented.
- Never derive phone numbers from LID strings.
- Phone mapping may be absent and must be handled safely.

## Name Extraction Rules
- Use contact name priority:
  1. `FullName`
  2. `BusinessName`
  3. `PushName`
  4. `FirstName`
- Preserve fallback behavior when fields are missing.

## Search Rules
- Keep case-insensitive matching for text query.
- Support query matching by name and phone.
- Keep filters for name/LID presence when implemented.
- Return empty result set for no matches.

## API Rules
- Keep controller naming pattern `api_handlers+*Controller.go`.
- Keep swagger annotations synchronized with endpoint behavior.
- Regenerate swagger after endpoint/model/annotation changes.

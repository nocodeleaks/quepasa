package whatsmeow

import (
	"regexp"
	"strings"

	_ "github.com/mattn/go-sqlite3"
)

var BUTTONSMSGREGEX regexp.Regexp = *regexp.MustCompile(`(?i)(?P<content>.*)\s?[\$#]buttons:\[(?P<buttons>.*)\]\s?(?P<footer>.*)`)
var BUTTONSREGEXCONTENTINDEX int = BUTTONSMSGREGEX.SubexpIndex("content")
var BUTTONSREGEXFOOTERINDEX int = BUTTONSMSGREGEX.SubexpIndex("footer")
var BUTTONSREGEXBUTTONSINDEX int = BUTTONSMSGREGEX.SubexpIndex("buttons")

var RegexButton regexp.Regexp = *regexp.MustCompile(`\((?P<value>.*)\)(?P<display>.*)`)
var RegexButtonValue int = RegexButton.SubexpIndex("value")
var RegexButtonDisplay int = RegexButton.SubexpIndex("display")

// IsValidForButtons reports whether the text contains a valid $buttons:[...] or #buttons:[...] block.
func IsValidForButtons(text string) bool {
	lowerText := strings.ToLower(text)
	if strings.Contains(lowerText, "buttons:") {
		matches := BUTTONSMSGREGEX.FindStringSubmatch(text)
		if len(matches) > 0 {
			if len(strings.TrimSpace(matches[0])) > 0 {
				return true
			}
		}
	}
	return false
}

// ConvertButtonsToText parses the $buttons:[...] syntax and renders the message as plain
// formatted text. WhatsApp blocks all interactive button proto types (ButtonsMessage,
// InteractiveMessage/NativeFlowMessage) for accounts not connected via the official
// Business Cloud API — the server accepts the send and returns success but silently
// discards the message. Plain text is the only format that reliably arrives on all clients.
//
// Example input:  "*Bom dia!* $buttons:[ (1)Suporte, (2)Financeiro ]"
// Example output: "*Bom dia!*\n*(1)* Suporte\n*(2)* Financeiro"
func ConvertButtonsToText(messageText string) string {
	matches := BUTTONSMSGREGEX.FindStringSubmatch(messageText)
	content := strings.TrimSpace(matches[BUTTONSREGEXCONTENTINDEX])
	footer := strings.TrimSpace(matches[BUTTONSREGEXFOOTERINDEX])
	buttonsRaw := matches[BUTTONSREGEXBUTTONSINDEX]

	var sb strings.Builder
	if content != "" {
		sb.WriteString(content)
	}

	for _, s := range strings.Split(buttonsRaw, ",") {
		normalized := strings.TrimSpace(s)
		if normalized == "" {
			continue
		}

		displayText := normalized
		buttonID := normalized

		btnMatches := RegexButton.FindStringSubmatch(normalized)
		if len(btnMatches) > 0 {
			if v := btnMatches[RegexButtonValue]; v != "" {
				buttonID = v
			}
			if d := strings.TrimSpace(btnMatches[RegexButtonDisplay]); d != "" {
				displayText = d
			}
		}

		sb.WriteString("\n*" + buttonID + ")* " + displayText)
	}

	if footer != "" {
		sb.WriteString("\n" + footer)
	}

	return sb.String()
}

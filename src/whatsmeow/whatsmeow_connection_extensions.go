package whatsmeow

import (
	"regexp"

	types "go.mau.fi/whatsmeow/types"
)

func GetMentions(text string) (mentions []string) {

	re := regexp.MustCompile(`\@(\d{7,15})(?:[ \r\n]?)`)
	matches := re.FindAllStringSubmatch(text, -1)

	for row := 0; row < len(matches); row++ {
		if len(matches[row]) > 0 {
			jid := types.NewJID(matches[row][1], types.DefaultUserServer)
			mentions = append(mentions, jid.ToNonAD().String())
		}
	}

	return
}

// returns a valid chat title from local memory store
func GetChatTitleFromWId(source *WhatsmeowConnection, wid string) string {
	jid, err := types.ParseJID(wid)
	if err == nil {
		return GetChatTitle(source.Client, jid)
	}

	return ""
}

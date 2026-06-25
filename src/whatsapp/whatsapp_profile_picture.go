package whatsapp

import (
	"io"
	"net/http"
)

type WhatsappProfilePicture struct {

	/*
		<summary>
			Unique id of this picture
			Save this Id if u have to consult multiple times, it will ensure to download only if it have changed
		</summary>
	*/
	Id string `json:"id,omitempty"`

	// Dont know the difference yet
	Type string `json:"type,omitempty"`

	// Public Url to download, dont know for how long its valid
	Url string `json:"url,omitempty"`

	// Id of whatsapp contact or group of this picture
	ChatId string `json:"chatid,omitempty"`

	// Whatsapp id that was used to retrieve that info
	Wid string `json:"wid,omitempty"`
}

func (source *WhatsappProfilePicture) Download() (content []byte, err error) {
	resp, err := http.Get(source.Url)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	content, err = io.ReadAll(resp.Body)
	return
}

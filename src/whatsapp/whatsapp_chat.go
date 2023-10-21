package whatsapp

type WhatsappChat struct {
	Id    string `json:"id"`
	Title string `json:"title,omitempty"`
}

var WASYSTEMCHAT = WhatsappChat{Id: "system", Title: "Internal System Message"}

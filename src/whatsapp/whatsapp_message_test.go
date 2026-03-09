package whatsapp

import "testing"

func TestWhatsappMessageFromDirect(t *testing.T) {
	tests := []struct {
		name   string
		chatID string
		want   bool
	}{
		{name: "user suffix", chatID: "123@s.whatsapp.net", want: true},
		{name: "lid suffix", chatID: "abc@lid", want: true},
		{name: "group suffix", chatID: "120363000000000000@g.us", want: false},
		{name: "status broadcast", chatID: "status@broadcast", want: false},
		{name: "newsletter", chatID: "chan@newsletter", want: false},
		{name: "system", chatID: "system", want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msg := &WhatsappMessage{
				Chat: WhatsappChat{Id: tt.chatID},
			}
			if got := msg.FromDirect(); got != tt.want {
				t.Fatalf("FromDirect() = %v, want %v", got, tt.want)
			}
		})
	}
}

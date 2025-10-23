package whatsapp

type WhatsappLocation struct {
	Latitude  float64 `json:"latitude"`          // Required: Location latitude in degrees
	Longitude float64 `json:"longitude"`         // Required: Location longitude in degrees
	Name      string  `json:"name,omitempty"`    // Optional: Location name/description
	Address   string  `json:"address,omitempty"` // Optional: Location full address
}

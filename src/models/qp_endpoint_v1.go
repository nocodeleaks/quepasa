package models

// Destino de msg whatsapp
type QPEndpointV1 struct {
	ID     string `json:"id"`
	Phone  string `json:"phone,omitempty"`
	Title  string `json:"title,omitempty"`
	Status string `json:"status,omitempty"`
}

func (source QPEndpointV1) GetQPEndPointV2() QPEndpointV2 {
	ob2 := QPEndpointV2{ID: source.ID, UserName: source.Phone, FirstName: source.Title, LastName: source.Status}
	return ob2
}

func (source QPEndpointV1) ToQpUserV2() QpUserV2 {
	result := QpUserV2{
		ID: source.ID,
	}
	return result
}

func (source QPEndpointV1) ToQPChatV2() QPChatV2 {
	result := QPChatV2{
		ID: source.ID,
	}
	return result
}

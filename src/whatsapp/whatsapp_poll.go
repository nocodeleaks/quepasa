package whatsapp

import "time"

type WhatsappPoll struct {
	Question   string   `json:"question"`             // Required: Poll question/title
	Options    []string `json:"options"`              // Required: Array of poll options
	Selections uint     `json:"selections,omitempty"` // Optional: Maximum number of options a user can select (default: 1)
	
	// Additional fields for poll information
	CreatedAt  *time.Time `json:"created_at,omitempty"`  // When the poll was created
	ExpiresAt  *time.Time `json:"expires_at,omitempty"`  // When the poll expires
	IsClosed   bool       `json:"is_closed,omitempty"`   // Whether the poll is closed
	IsSecret   bool       `json:"is_secret,omitempty"`   // Whether votes are secret/anonymous
	
	// Vote tracking
	TotalVotes int                    `json:"total_votes,omitempty"` // Total number of votes cast
	VoteCounts map[string]int         `json:"vote_counts,omitempty"` // Vote count per option
	Voters     map[string][]string    `json:"voters,omitempty"`      // List of voters per option (if not secret)
	
	// Metadata
	CreatorId string `json:"creator_id,omitempty"` // JID of poll creator
	MessageId string `json:"message_id,omitempty"` // Message ID of the poll creation message
}

// WhatsappPollVote represents a vote on a poll
type WhatsappPollVote struct {
	PollId      string    `json:"poll_id"`                // ID of the poll message
	VoterId     string    `json:"voter_id"`               // JID of the voter
	VoterName   string    `json:"voter_name,omitempty"`   // Display name of the voter
	VotedAt     time.Time `json:"voted_at"`               // When the vote was cast
	
	// Vote data
	SelectedOptions []string `json:"selected_options"` // The options that were voted for
	
	// Encrypted vote data (from WhatsApp)
	EncryptedPayload string `json:"encrypted_payload,omitempty"` // Base64 encoded encrypted vote
	EncryptedIV      string `json:"encrypted_iv,omitempty"`      // Base64 encoded encryption IV
}

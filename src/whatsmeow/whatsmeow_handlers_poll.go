package whatsmeow

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"time"

	whatsapp "github.com/nocodeleaks/quepasa/whatsapp"
	log "github.com/sirupsen/logrus"
	"go.mau.fi/whatsmeow/proto/waE2E"
	"go.mau.fi/whatsmeow/types"
	"go.mau.fi/whatsmeow/types/events"
)

// HandlePollCreationMessageAdvanced processes poll creation with enhanced features
func HandlePollCreationMessageAdvanced(log *log.Entry, out *whatsapp.WhatsappMessage, in *waE2E.PollCreationMessage) {
	log.Debug("processing poll creation message with advanced features")
	out.Type = whatsapp.PollMessageType

	// Convert options from protobuf to string slice
	options := make([]string, len(in.GetOptions()))
	for i, option := range in.GetOptions() {
		options[i] = option.GetOptionName()
	}

	// Create poll with enhanced information
	now := time.Now()
	out.Poll = &whatsapp.WhatsappPoll{
		Question:    in.GetName(),
		Options:     options,
		Selections:  uint(in.GetSelectableOptionsCount()),
		CreatedAt:   &now,
		MessageId:   out.Id,
		TotalVotes:  0,
		VoteCounts:  make(map[string]int),
		Voters:      make(map[string][]string),
	}

	// Build formatted text representation
	text := fmt.Sprintf("📊 *%s*\n\n", in.GetName())
	for i, option := range options {
		text += fmt.Sprintf("%d. %s\n", i+1, option)
	}
	
	if in.GetSelectableOptionsCount() > 1 {
		text += fmt.Sprintf("\n_Você pode selecionar até %d opções_", in.GetSelectableOptionsCount())
	} else {
		text += "\n_Selecione uma opção_"
	}
	
	text += "\n\n📈 *Resultados:* 0 votos"
	out.Text = text

	// Handle context info if available
	info := in.GetContextInfo()
	if info != nil {
		out.ForwardingScore = info.GetForwardingScore()
		out.InReply = info.GetStanzaID()
	}

	log.Infof("poll created: %s with %d options", in.GetName(), len(options))
}

// HandlePollUpdateMessageAdvanced processes poll votes with decryption attempts
func HandlePollUpdateMessageAdvanced(log *log.Entry, handler *WhatsmeowHandlers, out *whatsapp.WhatsappMessage, in *waE2E.PollUpdateMessage) {
	log.Debug("processing poll vote with advanced features")
	out.Type = whatsapp.PollMessageType

	// Get the original poll creation message to understand what was voted
	pollCreationKey := in.GetPollCreationMessageKey()
	var originalPoll *whatsapp.WhatsappPoll
	
	if pollCreationKey != nil {
		pollId := pollCreationKey.GetID()
		log.Debugf("looking for original poll message: %s", pollId)
		
		// Try to get the original poll message from cache
		if handler.WAHandlers != nil {
			originalMsg, err := handler.WAHandlers.GetById(pollId)
			if err == nil && originalMsg.Poll != nil {
				originalPoll = originalMsg.Poll
				log.Debugf("found original poll: %s", originalPoll.Question)
			} else {
				log.Warnf("could not find original poll message: %v", err)
			}
		}
	}

	// Create poll vote information
	vote := in.GetVote()
	pollVote := &whatsapp.WhatsappPollVote{
		VotedAt: time.Now(),
	}

	if pollCreationKey != nil {
		pollVote.PollId = pollCreationKey.GetID()
	}

	// Extract voter information from message info
	if out.Chat.Id != "" {
		pollVote.VoterId = out.Chat.Id
		pollVote.VoterName = out.Chat.Title
	}

	// Handle encrypted vote data
	if vote != nil {
		encPayload := vote.GetEncPayload()
		encIV := vote.GetEncIV()
		
		if len(encPayload) > 0 && len(encIV) > 0 {
			// Convert to base64 for storage/transmission
			pollVote.EncryptedPayload = base64.StdEncoding.EncodeToString(encPayload)
			pollVote.EncryptedIV = base64.StdEncoding.EncodeToString(encIV)
			
			log.Debugf("encrypted vote data - payload: %d bytes, IV: %d bytes", len(encPayload), len(encIV))
			
			// Create a temporary message event for decryption
			if out.InfoForHistory != nil {
				// Convert interface{} to types.MessageInfo
				if msgInfo, ok := out.InfoForHistory.(types.MessageInfo); ok {
					tempEvent := &events.Message{
						Info:    msgInfo,
						Message: &waE2E.Message{PollUpdateMessage: in},
					}
					
					// Attempt to decrypt the vote using whatsmeow's built-in decryption
					selectedOptions := attemptVoteDecryption(handler, in, tempEvent, originalPoll)
					pollVote.SelectedOptions = selectedOptions
				} else {
					log.Warn("InfoForHistory is not of type types.MessageInfo")
				}
			} else {
				log.Warn("cannot decrypt vote without message info")
			}
		}
	}

	// Build response message
	var text string
	if originalPoll != nil {
		// Copy poll information
		out.Poll = &whatsapp.WhatsappPoll{
			Question:   originalPoll.Question,
			Options:    originalPoll.Options,
			Selections: originalPoll.Selections,
			MessageId:  originalPoll.MessageId,
		}
		
		voterName := pollVote.VoterName
		if voterName == "" {
			voterName = "Alguém"
		}
		
		text = fmt.Sprintf("🗳️ *Voto registrado*\n\n📊 **%s**\n\n👤 %s votou", 
			originalPoll.Question, voterName)
			
		if len(pollVote.SelectedOptions) > 0 {
			text += "\n\n✅ *Opções selecionadas:*\n"
			for _, option := range pollVote.SelectedOptions {
				text += fmt.Sprintf("• %s\n", option)
			}
		} else {
			text += "\n\n🔒 _Voto criptografado (não foi possível descriptografar)_"
			
			// Include encrypted data for debugging
			if vote != nil {
				encPayloadStr := base64.StdEncoding.EncodeToString(vote.GetEncPayload())
				encIVStr := base64.StdEncoding.EncodeToString(vote.GetEncIV())
				text += fmt.Sprintf("\n\n_Dados criptografados:_\nPayload: %s\nIV: %s", encPayloadStr, encIVStr)
			}
		}
	} else {
		text = "🗳️ *Voto em enquete*\n\n_Enquete original não encontrada no cache_"
		
		// Still show encrypted data even without original poll
		if vote != nil {
			encPayloadStr := base64.StdEncoding.EncodeToString(vote.GetEncPayload())
			encIVStr := base64.StdEncoding.EncodeToString(vote.GetEncIV())
			text += fmt.Sprintf("\n\n_Dados criptografados:_\nPayload: %s\nIV: %s", encPayloadStr, encIVStr)
		}
	}

	out.Text = text

	// Store comprehensive vote data in debug info for webhook processing
	out.Debug = &whatsapp.WhatsappMessageDebug{
		Event:  "poll_vote",
		Reason: func() string {
			if len(pollVote.SelectedOptions) > 0 {
				return "vote_decrypted"
			}
			return "vote_encrypted"
		}(),
		Info: map[string]interface{}{
			"poll_vote":              pollVote,
			"poll_creation_key":      pollCreationKey,
			"vote_metadata":          in.GetMetadata(),
			"sender_timestamp_ms":    in.GetSenderTimestampMS(),
			"original_poll_found":    originalPoll != nil,
			"decryption_successful":  len(pollVote.SelectedOptions) > 0,
		},
	}

	log.Infof("poll vote processed for poll %s by %s - decrypted: %v", pollVote.PollId, pollVote.VoterName, len(pollVote.SelectedOptions) > 0)
}

// attemptVoteDecryption tries to decrypt the vote payload using whatsmeow's built-in DecryptPollVote
func attemptVoteDecryption(handler *WhatsmeowHandlers, pollUpdateMessage *waE2E.PollUpdateMessage, messageInfo *events.Message, poll *whatsapp.WhatsappPoll) []string {
	if handler == nil || handler.Client == nil {
		log.Debug("handler or client is nil, cannot decrypt vote")
		return []string{}
	}

	// Create a context for the decryption operation
	ctx := context.Background()
	
	// Use whatsmeow's built-in DecryptPollVote function
	pollVoteMessage, err := handler.Client.DecryptPollVote(ctx, messageInfo)
	if err != nil {
		log.Debugf("failed to decrypt poll vote: %v", err)
		return []string{}
	}

	if pollVoteMessage == nil {
		log.Debug("decrypted poll vote message is nil")
		return []string{}
	}

	// Get the selected option hashes from the decrypted vote
	selectedHashes := pollVoteMessage.GetSelectedOptions()
	if len(selectedHashes) == 0 {
		log.Debug("no selected options in decrypted vote")
		return []string{}
	}

	// If we don't have the original poll options, we can't map hashes back to option names
	if poll == nil || len(poll.Options) == 0 {
		log.Debug("no original poll options available to map hashes")
		return []string{}
	}

	// Create SHA-256 hashes for each poll option and match with selected hashes
	var selectedOptions []string
	for _, selectedHash := range selectedHashes {
		for _, option := range poll.Options {
			// Calculate SHA-256 hash of the option name
			optionHash := sha256.Sum256([]byte(option))
			
			// Compare the calculated hash with the selected hash
			if len(selectedHash) == len(optionHash) {
				match := true
				for i := 0; i < len(selectedHash); i++ {
					if selectedHash[i] != optionHash[i] {
						match = false
						break
					}
				}
				
				if match {
					selectedOptions = append(selectedOptions, option)
					break
				}
			}
		}
	}

	log.Debugf("successfully decrypted vote - %d options selected: %v", len(selectedOptions), selectedOptions)
	return selectedOptions
}

// ProcessPollForWebhook formats poll data for webhook delivery
func ProcessPollForWebhook(msg *whatsapp.WhatsappMessage) map[string]interface{} {
	webhook := map[string]interface{}{
		"type": "poll",
	}

	if msg.Poll != nil {
		webhook["poll"] = map[string]interface{}{
			"question":     msg.Poll.Question,
			"options":      msg.Poll.Options,
			"selections":   msg.Poll.Selections,
			"total_votes":  msg.Poll.TotalVotes,
			"vote_counts":  msg.Poll.VoteCounts,
			"is_closed":    msg.Poll.IsClosed,
			"is_secret":    msg.Poll.IsSecret,
			"creator_id":   msg.Poll.CreatorId,
			"message_id":   msg.Poll.MessageId,
		}

		if msg.Poll.CreatedAt != nil {
			webhook["poll"].(map[string]interface{})["created_at"] = msg.Poll.CreatedAt.Format(time.RFC3339)
		}

		if msg.Poll.ExpiresAt != nil {
			webhook["poll"].(map[string]interface{})["expires_at"] = msg.Poll.ExpiresAt.Format(time.RFC3339)
		}
	}

	// Add vote information if this is a poll vote
	if msg.Debug != nil && msg.Debug.Event == "poll_vote" {
		webhook["action"] = "vote"
		
		if info, ok := msg.Debug.Info.(map[string]interface{}); ok {
			if pollVote, exists := info["poll_vote"]; exists {
				webhook["vote"] = pollVote
				
				// Add decryption status
				if decrypted, exists := info["decryption_successful"]; exists {
					webhook["decryption_successful"] = decrypted
				}
			}
			
			// Add poll creation key info
			if pollKey, exists := info["poll_creation_key"]; exists {
				webhook["poll_creation_key"] = pollKey
			}
			
			// Add metadata and timestamp
			if metadata, exists := info["vote_metadata"]; exists {
				webhook["vote_metadata"] = metadata
			}
			
			if timestamp, exists := info["sender_timestamp_ms"]; exists {
				webhook["sender_timestamp_ms"] = timestamp
			}
		}
	} else {
		webhook["action"] = "create"
	}

	return webhook
}

package models

import (
	"context"
	"reflect"
	"strings"

	whatsapp "github.com/nocodeleaks/quepasa/whatsapp"
	log "github.com/sirupsen/logrus"
	"go.mau.fi/whatsmeow/proto/waE2E"
	"google.golang.org/protobuf/proto"
)

func ValidateItemBecauseUNOAPIConflict(item QpCacheItem, from string, previous any) bool {
	// debugging messages in cache
	if strings.HasPrefix(from, "message") {

		prevItem := previous.(QpCacheItem)

		logentry := log.New().WithContext(context.Background())
		logentry = logentry.WithField(LogFields.MessageId, item.Key)
		logentry = logentry.WithField("from", from)
		logentry.Level = log.DebugLevel

		logentry.Info("updating cache item ...")
		logentry.Infof("old type: %s, %v", reflect.TypeOf(prevItem.Value), prevItem.Value)
		logentry.Infof("new type: %s, %v", reflect.TypeOf(item.Value), item.Value)
		logentry.Infof("equals: %v, deep equals: %v", item.Value == prevItem.Value, reflect.DeepEqual(item.Value, prevItem.Value))

	var prevContent interface{}
	var prevOriginalMsg *waE2E.Message // Keep original Message for ads field comparison
	if prevWaMsg, ok := prevItem.Value.(*whatsapp.WhatsappMessage); ok {
		if nee, ok := prevWaMsg.Content.(*waE2E.Message); ok {
			prevOriginalMsg = nee // Save the full Message
			
			// Extract content for comparison (same as before)
			if nee.ExtendedTextMessage != nil {
				prevContent = nee.ExtendedTextMessage.GetText()
				logentry.Infof("old content from .ExtendedTextMessage as string: %s", prevContent)
			} else {
				conversation := nee.GetConversation()
				if len(conversation) > 0 {
					prevContent = conversation
					logentry.Infof("old content from .Message.Conversation: %s", prevContent)
				} else {
					prevContent = nee.String()
					logentry.Infof("old content as string: %s", prevContent)
				}
			}
		}
	}

	var newContent interface{}
	var newOriginalMsg *waE2E.Message // Keep original Message for ads field comparison
	if newWaMsg, ok := item.Value.(*whatsapp.WhatsappMessage); ok {
		if nee, ok := newWaMsg.Content.(*waE2E.Message); ok {
			newOriginalMsg = nee // Save the full Message
			
			// Extract content for comparison (same as before)
			if nee.ExtendedTextMessage != nil {
				newContent = nee.ExtendedTextMessage.GetText()
				logentry.Infof("new content from .ExtendedTextMessage as string: %s", newContent)
			} else {
				conversation := nee.GetConversation()
				if len(conversation) > 0 {
					newContent = conversation
					logentry.Infof("new content from .Message.Conversation: %s", newContent)
				} else {
					newContent = nee.String()
					logentry.Infof("new content as string: %s", newContent)
				}
			}
		}
	}

	if prevContent != nil && newContent != nil {
		logentry.Infof("content equals: %v, content deep equals: %v", prevContent == newContent, reflect.DeepEqual(prevContent, newContent))

		// CRITICAL FIX: For ads messages (ExtendedTextMessage), ignore volatile delay fields
		// These fields change from N to 0 on retries but don't affect actual message content:
		// - conversionDelaySeconds (5→0, 4→0, 3→0, etc)
		// - entryPointConversionDelaySeconds (same behavior)
		if prevOriginalMsg != nil && newOriginalMsg != nil {
			if prevOriginalMsg.ExtendedTextMessage != nil && newOriginalMsg.ExtendedTextMessage != nil {
				if prevOriginalMsg.ExtendedTextMessage.ContextInfo != nil && newOriginalMsg.ExtendedTextMessage.ContextInfo != nil {
					// Clone both messages to avoid modifying originals
					prevClone := proto.Clone(prevOriginalMsg).(*waE2E.Message)
					newClone := proto.Clone(newOriginalMsg).(*waE2E.Message)

					// Remove volatile delay fields before comparison
					prevClone.ExtendedTextMessage.ContextInfo.ConversionDelaySeconds = nil
					prevClone.ExtendedTextMessage.ContextInfo.EntryPointConversionDelaySeconds = nil
					newClone.ExtendedTextMessage.ContextInfo.ConversionDelaySeconds = nil
					newClone.ExtendedTextMessage.ContextInfo.EntryPointConversionDelaySeconds = nil

					// Compare the clones (without volatile delay fields)
					isEqual := reflect.DeepEqual(prevClone, newClone)

					if isEqual {
						logentry.Info("content is equal ignoring volatile delay fields, denying trigger - duplicate ads message detected")
						return false // Deny trigger
					}

					logentry.Debug("content differs even ignoring volatile delay fields, allowing trigger")
					return true // Allow trigger
				}
			}
		}

		// Default behavior: if equals, deny triggers
		return !reflect.DeepEqual(prevContent, newContent)
	}
	}

	return true
}

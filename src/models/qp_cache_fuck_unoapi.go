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
		if prevWaMsg, ok := prevItem.Value.(*whatsapp.WhatsappMessage); ok {
			prevContent = prevWaMsg.Content

			if nee, ok := prevContent.(*waE2E.Message); ok {
				if neeETM, ok := prevContent.(*waE2E.ExtendedTextMessage); ok {
					prevContent = neeETM.Text
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
		if newWaMsg, ok := item.Value.(*whatsapp.WhatsappMessage); ok {
			newContent = newWaMsg.Content

			if nee, ok := newContent.(*waE2E.Message); ok {
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

		if prevContent != nil && newContent != nil {
			logentry.Infof("content equals: %v, content deep equals: %v", prevContent == newContent, reflect.DeepEqual(prevContent, newContent))

			// CRITICAL FIX: For Message comparison, ignore conversionDelaySeconds field
			// This field changes from 4 to 0 on retries but doesn't affect actual message content
			prevMsg, prevIsMsg := prevContent.(*waE2E.Message)
			newMsg, newIsMsg := newContent.(*waE2E.Message)

			if prevIsMsg && newIsMsg {
				// Clone both messages to avoid modifying originals
				prevClone := proto.Clone(prevMsg).(*waE2E.Message)
				newClone := proto.Clone(newMsg).(*waE2E.Message)

				// Remove conversionDelaySeconds before comparison (this is the only difference on retry)
				if prevClone.ExtendedTextMessage != nil && prevClone.ExtendedTextMessage.ContextInfo != nil {
					prevClone.ExtendedTextMessage.ContextInfo.ConversionDelaySeconds = nil
				}
				if newClone.ExtendedTextMessage != nil && newClone.ExtendedTextMessage.ContextInfo != nil {
					newClone.ExtendedTextMessage.ContextInfo.ConversionDelaySeconds = nil
				}

				// Compare the clones (without conversionDelaySeconds)
				isEqual := reflect.DeepEqual(prevClone, newClone)

				if isEqual {
					logentry.Info("content is equal ignoring conversionDelaySeconds, denying trigger - duplicate detected")
					return false // Deny trigger
				}

				logentry.Debug("content differs even ignoring conversionDelaySeconds, allowing trigger")
				return true // Allow trigger
			}

			// if equals, deny triggers
			return !reflect.DeepEqual(prevContent, newContent)
		}
	}

	return true
}

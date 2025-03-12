package whatsmeow

import (
	"encoding/json"
	"regexp"
	"strings"

	_ "github.com/mattn/go-sqlite3"
	"go.mau.fi/whatsmeow/proto/waE2E"
	"google.golang.org/protobuf/proto"
)

var BUTTONSMSGREGEX regexp.Regexp = *regexp.MustCompile(`(?i)(?P<content>.*)\s?[\$#]buttons:\[(?P<buttons>.*)\]\s?(?P<footer>.*)`)
var BUTTONSREGEXCONTENTINDEX int = BUTTONSMSGREGEX.SubexpIndex("content")
var BUTTONSREGEXFOOTERINDEX int = BUTTONSMSGREGEX.SubexpIndex("footer")
var BUTTONSREGEXBUTTONSINDEX int = BUTTONSMSGREGEX.SubexpIndex("buttons")

var RegexButton regexp.Regexp = *regexp.MustCompile(`\((?P<value>.*)\)(?P<display>.*)`)
var RegexButtonValue int = RegexButton.SubexpIndex("value")
var RegexButtonDisplay int = RegexButton.SubexpIndex("display")

func GenerateButtonsMessage(messageText string) *waE2E.ButtonsMessage {
	var contentText *string
	var footerText *string

	// button list
	var buttons []*waE2E.ButtonsMessage_Button

	matches := BUTTONSMSGREGEX.FindStringSubmatch(messageText)
	contentMatched := matches[BUTTONSREGEXCONTENTINDEX]
	if len(contentMatched) > 0 {
		contentText = proto.String(contentMatched)
	}

	footerMatched := matches[BUTTONSREGEXFOOTERINDEX]
	if len(footerMatched) > 0 {
		footerText = proto.String(footerMatched)
	}

	buttonsText := matches[BUTTONSREGEXBUTTONSINDEX]
	buttonsSplited := strings.Split(buttonsText, ",")
	for _, s := range buttonsSplited {
		normalized := strings.TrimSpace(s)

		buttonText := &waE2E.ButtonsMessage_Button_ButtonText{}
		buttonText.DisplayText = proto.String(normalized)
		buttonId := buttonText.DisplayText

		matchesButton := RegexButton.FindStringSubmatch(normalized)
		if len(matchesButton) > 0 {
			buttonValueMatched := matchesButton[RegexButtonValue]
			if len(buttonValueMatched) > 0 {
				buttonId = &buttonValueMatched
			}

			buttonDisplayMatched := matchesButton[RegexButtonDisplay]
			if len(buttonDisplayMatched) > 0 {
				buttonText.DisplayText = &buttonDisplayMatched
			}
		}

		buttonType := waE2E.ButtonsMessage_Button_RESPONSE.Enum()
		buttons = append(buttons, &waE2E.ButtonsMessage_Button{ButtonText: buttonText, ButtonID: buttonId, Type: buttonType})
	}

	headerType := waE2E.ButtonsMessage_EMPTY.Enum()
	return &waE2E.ButtonsMessage{HeaderType: headerType, ContentText: contentText, Buttons: buttons, FooterText: footerText}
}

func IsValidForButtons(text string) bool {
	lowerText := strings.ToLower(text)
	if strings.Contains(lowerText, "buttons:") {
		matches := BUTTONSMSGREGEX.FindStringSubmatch(text)
		if len(matches) > 0 {
			if len(strings.TrimSpace(matches[0])) > 0 {
				return true
			}
		}
	}
	return false
}

func GenerateButtonsMessageNew(messageText string) *waE2E.ButtonsMessage {
	msg := &waE2E.ButtonsMessage{
		ContentText: proto.String("لدي إستفسار بخصوص:"),
		HeaderType:  waE2E.ButtonsMessage_EMPTY.Enum(),
		Buttons: []*waE2E.ButtonsMessage_Button{
			{
				ButtonID:       proto.String("bt1"),
				ButtonText:     &waE2E.ButtonsMessage_Button_ButtonText{DisplayText: proto.String("نعم")},
				Type:           waE2E.ButtonsMessage_Button_RESPONSE.Enum(),
				NativeFlowInfo: &waE2E.ButtonsMessage_Button_NativeFlowInfo{},
			},
			{
				ButtonID:       proto.String("bt2"),
				ButtonText:     &waE2E.ButtonsMessage_Button_ButtonText{DisplayText: proto.String("لا")},
				Type:           waE2E.ButtonsMessage_Button_RESPONSE.Enum(), //proto.ButtonsMessage_Button_Type.Enum,
				NativeFlowInfo: &waE2E.ButtonsMessage_Button_NativeFlowInfo{},
			},
		},
	}
	return msg
}

func GenerateListResponseMessage(messageText string) *waE2E.ListResponseMessage {

	ListResponseMessage := &waE2E.ListResponseMessage{
		Title:       proto.String("title"),
		Description: proto.String("Description"),
		SingleSelectReply: &waE2E.ListResponseMessage_SingleSelectReply{
			SelectedRowID: proto.String("SelectedRowId1"),
		},
		//	ContextInfo: ,
		ListType: waE2E.ListResponseMessage_SINGLE_SELECT.Enum(),
	}

	return ListResponseMessage
}

func GenerateListMessage(messageText string) *waE2E.ListMessage {

	ListMessage := &waE2E.ListMessage{
		Title:       proto.String("ListMessage title"),
		Description: proto.String("ListMessage Description"),
		FooterText:  proto.String("ListMessage footer"),
		ButtonText:  proto.String("ListMessage ButtonText"),
		ListType:    waE2E.ListMessage_SINGLE_SELECT.Enum(),
		Sections: []*waE2E.ListMessage_Section{
			{
				Title: proto.String("Section1 title"),
				Rows: []*waE2E.ListMessage_Row{
					{
						RowID:       proto.String("id1"),
						Title:       proto.String("ListMessage section row title"),
						Description: proto.String("ListMessage section row desc"),
					},
					{
						RowID:       proto.String("id2"),
						Title:       proto.String("title 2"),
						Description: proto.String("desc 2"),
					},
				},
			},
			{
				Title: proto.String("Section2 title"),
				Rows: []*waE2E.ListMessage_Row{
					{
						RowID:       proto.String("id1"),
						Title:       proto.String("ListMessage section row title"),
						Description: proto.String("ListMessage section row desc"),
					},
					{
						RowID:       proto.String("id2"),
						Title:       proto.String("title 2"),
						Description: proto.String("desc 2"),
					},
				},
			},
		},
	}

	return ListMessage
}

func GenerateTemplateMessage(messageText string) *waE2E.TemplateMessage {
	TemplateMessage := &waE2E.TemplateMessage{
		HydratedTemplate: &waE2E.TemplateMessage_HydratedFourRowTemplate{
			Title: &waE2E.TemplateMessage_HydratedFourRowTemplate_HydratedTitleText{
				HydratedTitleText: "The Title",
			},
			TemplateID:          proto.String("template-id"),
			HydratedContentText: proto.String("The Content"),
			HydratedFooterText:  proto.String("The Footer"),

			HydratedButtons: []*waE2E.HydratedTemplateButton{

				// This for URL button
				{
					Index: proto.Uint32(1),
					HydratedButton: &waE2E.HydratedTemplateButton_UrlButton{
						UrlButton: &waE2E.HydratedTemplateButton_HydratedURLButton{
							DisplayText: proto.String("The Link"),
							URL:         proto.String("https://fb.me/this"),
						},
					},
				},

				// This for call button
				{
					Index: proto.Uint32(2),
					HydratedButton: &waE2E.HydratedTemplateButton_CallButton{
						CallButton: &waE2E.HydratedTemplateButton_HydratedCallButton{
							DisplayText: proto.String("Call us"),
							PhoneNumber: proto.String("1234567890"),
						},
					},
				},

				// This is just a quick reply
				{
					Index: proto.Uint32(3),
					HydratedButton: &waE2E.HydratedTemplateButton_QuickReplyButton{
						QuickReplyButton: &waE2E.HydratedTemplateButton_HydratedQuickReplyButton{
							DisplayText: proto.String("Quick reply"),
							ID:          proto.String("quick-id"),
						},
					},
				},
			},
		},
	}
	return TemplateMessage
}

type ButtonParams struct {
	DisplayText string `json:"displayText"`
	ID          string `json:"buttonID"`
}

func GenerateInteractiveMessage(messageText string) *waE2E.InteractiveMessage {
	jsonDataFirst, _ := json.Marshal(ButtonParams{
		DisplayText: "button - 01",
		ID:          ".bt1",
	})

	jsonDataSecond, _ := json.Marshal(ButtonParams{
		DisplayText: "button - 02",
		ID:          ".bt2",
	})

	msgVersion := int32(1)
	InteractiveMessage := &waE2E.InteractiveMessage{
		Header: &waE2E.InteractiveMessage_Header{
			Title:              proto.String("title"),
			Subtitle:           proto.String("subtitle"),
			HasMediaAttachment: proto.Bool(false),
		},
		Footer: &waE2E.InteractiveMessage_Footer{
			Text: proto.String("footer"),
		},
		Body: &waE2E.InteractiveMessage_Body{
			Text: proto.String("body"),
		},
		InteractiveMessage: &waE2E.InteractiveMessage_NativeFlowMessage_{
			NativeFlowMessage: &waE2E.InteractiveMessage_NativeFlowMessage{
				MessageVersion: &msgVersion,
				Buttons: []*waE2E.InteractiveMessage_NativeFlowMessage_NativeFlowButton{
					{
						Name:             proto.String("Yes"),
						ButtonParamsJSON: proto.String(string(jsonDataFirst)),
					},
					{
						Name:             proto.String("No"),
						ButtonParamsJSON: proto.String(string(jsonDataSecond)),
					},
				},
			},
		},
		ContextInfo: &waE2E.ContextInfo{},
	}
	return InteractiveMessage
}

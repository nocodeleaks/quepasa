module github.com/nocodeleaks/quepasa/runtime

go 1.25

require (
	github.com/nocodeleaks/quepasa/dispatch v0.0.0
	github.com/nocodeleaks/quepasa/library v0.0.0
	github.com/nocodeleaks/quepasa/models v0.0.0
	github.com/nocodeleaks/quepasa/whatsapp v0.0.0
)

replace (
	github.com/nocodeleaks/quepasa/dispatch => ../dispatch
	github.com/nocodeleaks/quepasa/library => ../library
	github.com/nocodeleaks/quepasa/models => ../models
	github.com/nocodeleaks/quepasa/whatsapp => ../whatsapp
)

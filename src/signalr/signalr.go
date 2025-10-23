package signalr

import (
	"context"

	"github.com/go-chi/chi/v5"
	kitlog "github.com/go-kit/log"
	webserver "github.com/nocodeleaks/quepasa/webserver"
	signalr "github.com/philippseith/signalr"
	log "github.com/sirupsen/logrus"
)

func init() {
	// Automatically registers the SignalR configuration in the webserver
	// This allows SignalR to be configured without the webserver module
	// needing to know specifically about SignalR
	webserver.RegisterRouterConfigurator(Configure)
}

// Configure automatically configures SignalR routes in the router
func Configure(r chi.Router) {
	// SignalR
	ServeSignalR(r)
}

func ServeSignalR(r chi.Router) {
	// setting group
	r.Group(func(r chi.Router) {
		log.Debug("starting signalr service")

		factory := signalr.UseHub(SignalRHub)
		//keepalive := signalr.KeepAliveInterval(2 * time.Second)
		//timeout := signalr.ChanReceiveTimeout(1 * time.Hour)

		ctx := context.Background()
		logentry := log.New().WithContext(ctx)

		// setting signalr log level
		logentry.Level = log.InfoLevel

		// should generate debug logs
		debug := logentry.Level >= log.DebugLevel

		slogger := signalr.Logger(kitlog.NewLogfmtLogger(logentry.Writer()), debug)
		server, err := signalr.NewServer(ctx, factory, slogger)
		if err != nil {
			logentry.Errorf("error on set signalr server: %s", err.Error())
		}

		mappable := WithChiRouter(r)
		server.MapHTTP(mappable, "/signalr")
	})
}

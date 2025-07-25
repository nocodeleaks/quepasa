package webserver

import (
	"context"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	api "github.com/nocodeleaks/quepasa/api"
	form "github.com/nocodeleaks/quepasa/form"
	models "github.com/nocodeleaks/quepasa/models"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	log "github.com/sirupsen/logrus"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	kitlog "github.com/go-kit/log"
	signalr "github.com/philippseith/signalr"

	// swagger embed files
	httpSwagger "github.com/swaggo/http-swagger"
)

func WebServerStart(logentry *log.Entry) error {
	r := newRouter()
	webAPIPort := os.Getenv(models.ENV_WEBAPIPORT)
	webAPIHost := os.Getenv(models.ENV_WEBAPIHOST)
	if len(webAPIPort) == 0 {
		webAPIPort = "31000"
	}

	var timeout = 30 * time.Second
	server := http.Server{
		Addr:         webAPIHost + ":" + webAPIPort,
		ReadTimeout:  timeout,
		WriteTimeout: timeout,
		Handler:      r,
	}

	logentry.Infof("starting web server on port: %s", webAPIPort)
	err := server.ListenAndServe()
	if err != nil {
		log.Fatal(err)
	}

	return err
}

func newRouter() chi.Router {
	r := chi.NewRouter()

	r.Use(MiddlewareForNormalizePaths)
	r.Use(middleware.StripSlashes)
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)

	if models.ENV.HttpLogs() {
		r.Use(middleware.Logger)
	}

	r.Use(middleware.Recoverer)

	// API routes, main content
	ServeAPI(r)

	// Form routes, extra content
	ServeForms(r)

	// SignalR
	ServeSignalR(r)

	// Static content
	ServeStaticContent(r)

	// Swagger Ui
	ServeSwaggerUi(r)

	// Metrics
	ServeMetrics(r)

	return r
}

func ServeForms(r chi.Router) {

	// setting group
	r.Group(func(r chi.Router) {

		// setting timeout for the group
		r.Use(middleware.Timeout(30 * time.Second))

		// web routes
		// authenticated web routes
		r.Group(form.RegisterFormAuthenticatedControllers)

		// unauthenticated web routes
		r.Group(form.RegisterFormControllers)
	})
}

func ServeAPI(r chi.Router) {

	// setting group
	r.Group(func(r chi.Router) {

		// setting timeout for the group
		r.Use(middleware.Timeout(30 * time.Second))

		/* CORS TESTING
		r.Use(cors.Handler(cors.Options{
			//AllowedOrigins:   []string{"https://foo.com"}, // Use this to allow specific origin hosts
			//AllowedOrigins: []string{"https://*", "http://*"},
			AllowOriginFunc: func(r *http.Request, origin string) bool { return true },
			//AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
			//AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
			//ExposedHeaders:   []string{"Link"},
			//AllowCredentials: false,
			// MaxAge: 300, // Maximum value not ignored by any of major browsers
		}))
		*/

		r.Group(api.RegisterAPIControllers)
		r.Group(api.RegisterAPIV2Controllers)
		r.Group(api.RegisterAPIV3Controllers)
	})
}

func ServeStaticContent(r chi.Router) {

	// setting group
	r.Group(func(r chi.Router) {

		// static files
		workDir, _ := os.Getwd()
		assetsDir := filepath.Join(workDir, "assets")
		root := http.Dir(assetsDir)

		path := "/assets"

		if strings.ContainsAny(path, "{}*") {
			panic("FileServer does not permit URL parameters.")
		}

		fs := http.StripPrefix(path, http.FileServer(root))
		if path != "/" && path[len(path)-1] != '/' {
			r.Get(path, http.RedirectHandler(path+"/", http.StatusMovedPermanently).ServeHTTP)
			path += "/"
		}
		path += "*"
		r.Get(path, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			fs.ServeHTTP(w, r)
		}))

	})
}

func ServeSignalR(r chi.Router) {

	// setting group
	r.Group(func(r chi.Router) {
		log.Debug("starting signalr service")

		factory := signalr.UseHub(models.SignalRHub)
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

func ServeSwaggerUi(r chi.Router) {
	log.Debug("starting swaggerUi service")
	r.Mount("/swagger", httpSwagger.WrapHandler)
}

func ServeMetrics(r chi.Router) {
	log.Debug("starting metrics service")
	r.Handle("/metrics", promhttp.Handler())
}

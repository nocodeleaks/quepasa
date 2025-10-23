package webserver

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	environment "github.com/nocodeleaks/quepasa/environment"
	log "github.com/sirupsen/logrus"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

// RouterConfigurator é uma função que pode configurar rotas adicionais no router
type RouterConfigurator func(r chi.Router)

// configurators armazena as funções de configuração adicionais que serão executadas
// durante a criação do router. Módulos externos podem registrar suas configurações aqui.
var configurators []RouterConfigurator

// RegisterRouterConfigurator permite que módulos externos registrem funções
// para configurar rotas adicionais no router principal
func RegisterRouterConfigurator(configurator RouterConfigurator) {
	configurators = append(configurators, configurator)
}

func WebServerStart(logentry *log.Entry) error {
	r := newRouter()
	webAPIPort := environment.Settings.WebServer.Port
	webAPIHost := environment.Settings.WebServer.Host

	var timeout = 30 * time.Second
	server := http.Server{
		Addr:         fmt.Sprintf("%s:%d", webAPIHost, webAPIPort),
		ReadTimeout:  timeout,
		WriteTimeout: timeout,
		Handler:      r,
	}

	logentry.Infof("starting web server on port: %d", webAPIPort)
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

	if environment.Settings.WebServer.Logs {
		r.Use(middleware.Logger)
	}

	r.Use(middleware.Recoverer)

	// Form and SignalR routes are now configured automatically via configurators

	// Execute registered configurators (e.g., Swagger, custom modules)
	for _, configurator := range configurators {
		configurator(r)
	}

	return r
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

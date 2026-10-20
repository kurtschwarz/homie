// Package server wires the HTTP mux: the GraphQL endpoint, a plain
// healthz probe for the systemd watchdog, and the two embedded SPAs.
package server

import (
	"io/fs"
	"net/http"
	"strings"
	"time"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/handler/transport"
	"github.com/99designs/gqlgen/graphql/playground"

	"github.com/kurtschwarz/homie/services/api/internal/auth"
	"github.com/kurtschwarz/homie/services/api/internal/graph"
)

type Config struct {
	Resolver   *graph.Resolver
	Display    fs.FS
	Manager    fs.FS
	Photos     fs.FS
	Playground bool
	Token      string
	Signer     *auth.Signer
}

func New(cfg Config) http.Handler {
	return newHandler(cfg, false, true)
}

// NewLAN serves only network-facing routes. GraphQL requires a bearer token;
// photos use expiring signed URLs so native image loaders need no headers.
func NewLAN(cfg Config) http.Handler { return newHandler(cfg, true, false) }

// NewLocal serves the kiosk and manager UI on a loopback-only listener.
func NewLocal(cfg Config) http.Handler { return newHandler(cfg, false, true) }

func newHandler(cfg Config, protectGraphQL, includeLocalRoutes bool) http.Handler {
	mux := http.NewServeMux()

	gqlSrv := handler.New(graph.NewExecutableSchema(graph.Config{Resolvers: cfg.Resolver}))
	gqlSrv.AddTransport(transport.Options{})
	gqlSrv.AddTransport(transport.GET{})
	gqlSrv.AddTransport(transport.POST{})
	gqlSrv.AddTransport(transport.MultipartForm{MaxMemory: 10 << 20, MaxUploadSize: 20 << 20})
	graphqlHandler := http.Handler(gqlSrv)
	if protectGraphQL {
		graphqlHandler = auth.RequireBearer(cfg.Token, graphqlHandler)
	}
	mux.Handle("/graphql", graphqlHandler)

	if cfg.Playground && includeLocalRoutes {
		mux.Handle("/playground", playground.Handler("homie", "/graphql"))
	}

	// Plain HTTP, independent of the GraphQL stack: the systemd watchdog
	// probe must keep working even if resolvers are failing.
	if includeLocalRoutes {
		mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("ok"))
		})
	}

	// Strip "/manage" and not "/manage/": the handlers below expect a
	// rooted path, and "/manage/" alone strips down to "".
	if cfg.Manager != nil {
		mux.Handle("/manage/", http.StripPrefix("/manage", spaHandler(cfg.Manager)))
	}
	mux.Handle("/photos/", http.StripPrefix("/photos/", photoHandler(cfg.Photos, cfg.Signer)))
	if includeLocalRoutes {
		mux.Handle("/", spaHandler(cfg.Display))
	}

	return mux
}

// photoHandler serves stored photo files, rejecting anything that isn't a
// plain .jpg name so temp files and nested paths can't be fetched.
func photoHandler(fsys fs.FS, signer *auth.Signer) http.Handler {
	var fileServer http.Handler
	if fsys != nil {
		fileServer = http.FileServerFS(fsys)
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if fileServer == nil {
			http.NotFound(w, r)
			return
		}
		name := r.URL.Path
		if strings.Contains(name, "/") || !strings.HasSuffix(name, ".jpg") {
			http.NotFound(w, r)
			return
		}
		if signer != nil {
			expiry, err := auth.ParseExpiry(r.URL.Query().Get("exp"))
			if err != nil || !signer.Verify(name, expiry, r.URL.Query().Get("sig"), time.Now()) {
				http.NotFound(w, r)
				return
			}
		}
		w.Header().Set("Cache-Control", "no-cache")
		fileServer.ServeHTTP(w, r)
	})
}

// spaHandler serves files from fsys, falling back to index.html for any
// path that doesn't match a real file so client-side routing works.
func spaHandler(fsys fs.FS) http.Handler {
	fileServer := http.FileServerFS(fsys)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		if path == "" {
			path = "/"
		}
		trimmed := strings.TrimPrefix(path, "/")
		if trimmed != "" {
			if _, err := fs.Stat(fsys, trimmed); err != nil {
				r = r.Clone(r.Context())
				r.URL.Path = "/"
			}
		}
		fileServer.ServeHTTP(w, r)
	})
}

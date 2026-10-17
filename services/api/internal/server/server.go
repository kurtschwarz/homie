// Package server wires the HTTP mux: the GraphQL endpoint, a plain
// healthz probe for the systemd watchdog, and the two embedded SPAs.
package server

import (
	"io/fs"
	"net/http"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/handler/transport"
	"github.com/99designs/gqlgen/graphql/playground"

	"github.com/kurtschwarz/homie/services/api/internal/graph"
)

type Config struct {
	Resolver   *graph.Resolver
	Display    fs.FS
	Manager    fs.FS
	Playground bool
}

func New(cfg Config) http.Handler {
	mux := http.NewServeMux()

	gqlSrv := handler.New(graph.NewExecutableSchema(graph.Config{Resolvers: cfg.Resolver}))
	gqlSrv.AddTransport(transport.Options{})
	gqlSrv.AddTransport(transport.GET{})
	gqlSrv.AddTransport(transport.POST{})
	gqlSrv.AddTransport(transport.MultipartForm{MaxMemory: 10 << 20, MaxUploadSize: 20 << 20})
	gqlSrv.AddTransport(transport.Websocket{})
	mux.Handle("/graphql", gqlSrv)

	if cfg.Playground {
		mux.Handle("/playground", playground.Handler("homie", "/graphql"))
	}

	// Plain HTTP, independent of the GraphQL stack: the systemd watchdog
	// probe must keep working even if resolvers are failing.
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	mux.Handle("/manage/", http.StripPrefix("/manage/", spaHandler(cfg.Manager)))
	mux.Handle("/", spaHandler(cfg.Display))

	return mux
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
		trimmed := path[1:]
		if trimmed != "" {
			if _, err := fs.Stat(fsys, trimmed); err != nil {
				r = r.Clone(r.Context())
				r.URL.Path = "/"
			}
		}
		fileServer.ServeHTTP(w, r)
	})
}

// Package server wires the HTTP mux: the GraphQL endpoint, a plain
// healthz probe for the systemd watchdog, and the two embedded SPAs.
package server

import (
	"context"
	"crypto/subtle"
	"io/fs"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/99designs/gqlgen/graphql"
	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/handler/transport"
	"github.com/99designs/gqlgen/graphql/playground"

	"github.com/kurtschwarz/homie/services/api/internal/auth"
	"github.com/kurtschwarz/homie/services/api/internal/graph"
	"github.com/kurtschwarz/homie/services/api/internal/graph/model"
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

// New serves everything on one listener and always protects GraphQL.
func New(cfg Config) http.Handler {
	return newHandler(cfg, true, true)
}

// NewLAN serves only network-facing routes. GraphQL requires a bearer token;
// photos use expiring signed URLs so native image loaders need no headers.
func NewLAN(cfg Config) http.Handler { return newHandler(cfg, true, false) }

// NewLocal serves the kiosk on a loopback-only listener, with no
// credential required: the manager is network-facing and lives on NewLAN
// instead, so this never serves an unauthenticated admin UI.
func NewLocal(cfg Config) http.Handler { return newHandler(cfg, false, true) }

func newHandler(cfg Config, protectGraphQL, includeLocalRoutes bool) http.Handler {
	mux := http.NewServeMux()

	gqlSrv := handler.New(graph.NewExecutableSchema(graph.Config{Resolvers: cfg.Resolver}))
	gqlSrv.AddTransport(transport.Options{})
	gqlSrv.AddTransport(transport.POST{})
	gqlSrv.AddTransport(transport.MultipartForm{MaxMemory: 10 << 20, MaxUploadSize: 20 << 20})
	if protectGraphQL {
		// The whole security boundary: every root field an operation
		// selects must be public, or the caller must prove a credential.
		// GET is deliberately not registered above - with one endpoint for
		// everything, a SameSite=Lax cookie riding a cross-site top-level
		// GET navigation would be the only CSRF shape left.
		//
		// includeLocalRoutes doubles as "this listener serves the display",
		// which is what DISPLAY-kind fields are narrowed to.
		gqlSrv.AroundOperations(authGate(cfg.Resolver, cfg.Token, includeLocalRoutes))
	}
	graphqlHandler := withRequestContext(http.Handler(gqlSrv))
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
	// rooted path, and "/manage/" alone strips down to "". The manager is
	// reached from the network, never from the kiosk, so it is only
	// registered on the LAN listener - serving it on loopback too would
	// leave an unauthenticated admin UI on a port that needs no auth at all.
	if cfg.Manager != nil && protectGraphQL {
		mux.Handle("/manage/", http.StripPrefix("/manage", spaHandler(cfg.Manager)))
	}
	mux.Handle("/photos/", http.StripPrefix("/photos/", photoHandler(cfg.Photos, cfg.Signer)))
	if includeLocalRoutes {
		mux.Handle("/", spaHandler(cfg.Display))
	}

	return mux
}

// withRequestContext makes the client's credentials and address available
// to resolvers and the operation gate, and gives resolvers a narrow way to
// set or clear the session cookie - gqlgen does not hand them the response.
func withRequestContext(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var bearer string
		if header := r.Header.Get("Authorization"); strings.HasPrefix(header, "Bearer ") {
			bearer = strings.TrimSpace(strings.TrimPrefix(header, "Bearer "))
		}
		var sessionToken string
		if cookie, err := r.Cookie(auth.SessionCookieName); err == nil {
			sessionToken = cookie.Value
		}
		ctx := auth.WithRequestInfo(r.Context(), auth.RequestInfo{
			BearerToken:  bearer,
			SessionToken: sessionToken,
			ClientIP:     clientIP(r),
			UserAgent:    r.UserAgent(),
		})
		ctx = auth.WithSessionCookies(ctx, w, r.TLS != nil)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func clientIP(r *http.Request) string {
	if host, _, err := net.SplitHostPort(r.RemoteAddr); err == nil {
		return host
	}
	return r.RemoteAddr
}

// authGate is gqlgen's one security boundary. Three things can satisfy it:
// a public operation needs nothing; the shared bearer token proves the
// caller is the display; and a valid session cookie resolves into the
// current user, who must hold whatever role the operation's fields demand.
//
// servesDisplay reports whether this listener is one the display is served
// on. A field whose kinds exclude USER is refused everywhere else before
// any credential is considered, so holding the bearer token does not widen
// the surface.
func authGate(resolver *graph.Resolver, token string, servesDisplay bool) graphql.OperationMiddleware {
	deny := func(message string) graphql.ResponseHandler {
		return func(ctx context.Context) *graphql.Response {
			return graphql.ErrorResponse(ctx, "%s", message)
		}
	}
	return func(ctx context.Context, next graphql.OperationHandler) graphql.ResponseHandler {
		oc := graphql.GetOperationContext(ctx)
		if !servesDisplay && graph.SelectsDisplayOnlyField(oc.Operation) {
			return deny("unauthorized: only served to the display")
		}
		if graph.IsPublicOperation(oc.Operation) {
			return next(ctx)
		}
		// TODO(photo-sync-auth): Photo Sync and the iOS app still carry this
		// token rather than a session; once they move to user-based auth,
		// delete this branch and Config.Token along with it.
		if bearer := auth.BearerTokenFromContext(ctx); token != "" && len(bearer) == len(token) &&
			subtle.ConstantTimeCompare([]byte(bearer), []byte(token)) == 1 {
			if !graph.OperationAllowsKind(oc.Operation, string(model.AuthKindDisplay)) {
				return deny("unauthorized")
			}
			return next(ctx)
		}
		if resolver != nil && resolver.Accounts != nil {
			if user, session, err := resolver.Accounts.CurrentUser(ctx, auth.SessionTokenFromContext(ctx)); err == nil {
				if !graph.OperationAllowsKind(oc.Operation, string(model.AuthKindUser)) {
					return deny("unauthorized")
				}
				role := string(user.Role)
				if !graph.OperationAllowsRole(oc.Operation, role) {
					return deny("forbidden")
				}
				return next(auth.WithCurrentUser(ctx, auth.CurrentUser{UserID: user.ID, SessionID: session.ID, Role: role}))
			}
		}
		return deny("unauthorized")
	}
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

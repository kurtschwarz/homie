package auth

import (
	"context"
	"net/http"
	"time"

	"github.com/google/uuid"
)

// SessionCookieName is the cookie carrying the opaque session token.
const SessionCookieName = "homie_session"

type contextKey string

const (
	bearerTokenKey   contextKey = "bearer_token"
	sessionTokenKey  contextKey = "session_token"
	clientIPKey      contextKey = "client_ip"
	userAgentKey     contextKey = "user_agent"
	sessionCookieKey contextKey = "session_cookies"
	currentUserKey   contextKey = "current_user"
)

// RequestInfo is what the gate needs from the raw HTTP request: the two
// forms a caller can prove a credential, and the address rate limiting is
// keyed on.
type RequestInfo struct {
	BearerToken  string
	SessionToken string
	ClientIP     string
	UserAgent    string
}

func WithRequestInfo(ctx context.Context, info RequestInfo) context.Context {
	ctx = context.WithValue(ctx, bearerTokenKey, info.BearerToken)
	ctx = context.WithValue(ctx, sessionTokenKey, info.SessionToken)
	ctx = context.WithValue(ctx, clientIPKey, info.ClientIP)
	return context.WithValue(ctx, userAgentKey, info.UserAgent)
}

func BearerTokenFromContext(ctx context.Context) string {
	token, _ := ctx.Value(bearerTokenKey).(string)
	return token
}

func SessionTokenFromContext(ctx context.Context) string {
	token, _ := ctx.Value(sessionTokenKey).(string)
	return token
}

func ClientIPFromContext(ctx context.Context) string {
	ip, _ := ctx.Value(clientIPKey).(string)
	return ip
}

func UserAgentFromContext(ctx context.Context) string {
	userAgent, _ := ctx.Value(userAgentKey).(string)
	return userAgent
}

// SessionCookies lets a resolver set or clear the session cookie without
// otherwise touching the response: gqlgen does not hand resolvers the
// http.ResponseWriter, and this is the narrowest capability that covers
// what login, setup and logout need.
type SessionCookies struct {
	w      http.ResponseWriter
	secure bool
}

// Set writes the session cookie for token, expiring at expiresAt.
func (c *SessionCookies) Set(token string, expiresAt time.Time) {
	http.SetCookie(c.w, &http.Cookie{
		Name:     SessionCookieName,
		Value:    token,
		Path:     "/",
		Expires:  expiresAt,
		HttpOnly: true,
		Secure:   c.secure,
		SameSite: http.SameSiteLaxMode,
	})
}

// Clear removes the session cookie from the browser.
func (c *SessionCookies) Clear() {
	http.SetCookie(c.w, &http.Cookie{
		Name:     SessionCookieName,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   c.secure,
		SameSite: http.SameSiteLaxMode,
	})
}

func WithSessionCookies(ctx context.Context, w http.ResponseWriter, secure bool) context.Context {
	return context.WithValue(ctx, sessionCookieKey, &SessionCookies{w: w, secure: secure})
}

func CookiesFromContext(ctx context.Context) *SessionCookies {
	cookies, _ := ctx.Value(sessionCookieKey).(*SessionCookies)
	return cookies
}

// CurrentUser identifies the user a validated session belongs to, set by
// the operation gate once it has checked the session cookie.
type CurrentUser struct {
	UserID    uuid.UUID
	SessionID uuid.UUID
	Role      string
}

func WithCurrentUser(ctx context.Context, user CurrentUser) context.Context {
	return context.WithValue(ctx, currentUserKey, user)
}

func CurrentUserFromContext(ctx context.Context) (CurrentUser, bool) {
	user, ok := ctx.Value(currentUserKey).(CurrentUser)
	return user, ok
}

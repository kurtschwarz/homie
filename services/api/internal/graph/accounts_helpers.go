package graph

// Kept out of accounts.resolvers.go: gqlgen strips any function in a
// resolver file that it doesn't recognize as a resolver method on every
// regeneration, the same reason event_helpers.go and photo_helpers.go exist.

import (
	"context"
	"errors"

	"github.com/kurtschwarz/homie/services/api/internal/accounts"
	"github.com/kurtschwarz/homie/services/api/internal/auth"
	"github.com/kurtschwarz/homie/services/api/internal/graph/model"
)

var errAccountsNotConfigured = errors.New("accounts are not configured")

func userModel(user accounts.User) *model.User {
	return &model.User{ID: user.ID.String(), Username: user.Username, Role: model.Role(user.Role)}
}

func setSessionCookie(ctx context.Context, session accounts.Session) {
	if cookies := auth.CookiesFromContext(ctx); cookies != nil {
		cookies.Set(session.Token, session.ExpiresAt)
	}
}

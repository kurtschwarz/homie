package graph

// This file will not be regenerated automatically.
//
// It serves as dependency injection for your app, add any dependencies you require
// here.

import (
	"sync"
	"time"

	"github.com/kurtschwarz/homie/services/api/internal/accounts"
	"github.com/kurtschwarz/homie/services/api/internal/auth"
	"github.com/kurtschwarz/homie/services/api/internal/events"
	"github.com/kurtschwarz/homie/services/api/internal/photos"
	"github.com/kurtschwarz/homie/services/api/internal/weather"
)

type Resolver struct {
	TimerMu       sync.Mutex
	Timer         *TimerState
	Version       string
	StartedAt     time.Time
	Weather       *weather.Service
	Photos        *photos.Service
	PhotoRotation *photos.RotationService
	Events        *events.Service
	PhotoSigner   *auth.Signer
	Accounts      *accounts.Service
}

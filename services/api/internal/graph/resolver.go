package graph

// This file will not be regenerated automatically.
//
// It serves as dependency injection for your app, add any dependencies you require
// here.

import (
	"time"

	"github.com/kurtschwarz/homie/services/api/internal/events"
	"github.com/kurtschwarz/homie/services/api/internal/photos"
	"github.com/kurtschwarz/homie/services/api/internal/weather"
)

type Resolver struct {
	Version   string
	StartedAt time.Time
	Weather   *weather.Service
	Photos    *photos.Service
	Events    *events.Service
}

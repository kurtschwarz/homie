package graph

import (
	"errors"
	"strings"
	"time"

	"github.com/kurtschwarz/homie/services/api/internal/display"
	"github.com/kurtschwarz/homie/services/api/internal/graph/model"
)

var errDisplayNotConfigured = errors.New("display is not configured")

func displayModeToModel(mode display.Mode) model.DisplayMode {
	return model.DisplayMode(strings.ToUpper(string(mode)))
}

func displayModeFromModel(mode model.DisplayMode) display.Mode {
	return display.Mode(strings.ToLower(string(mode)))
}

func displayScheduleToModel(schedule display.Schedule) *model.DisplaySchedule {
	modes := make([]model.DisplayModeSetting, 0, len(schedule.Modes))
	for _, setting := range schedule.Modes {
		modes = append(modes, model.DisplayModeSetting{Mode: displayModeToModel(setting.Mode), StartsAtMinutes: setting.StartsAtMinutes, Brightness: setting.Brightness})
	}
	return &model.DisplaySchedule{Modes: modes, PowerSavingEnabled: schedule.PowerSavingEnabled, WakeForAlerts: schedule.WakeForAlerts, DdcBus: schedule.DDCBus}
}

func displayScheduleFromInput(input model.UpdateDisplayScheduleInput) display.Schedule {
	schedule := display.Schedule{Settings: display.Settings{PowerSavingEnabled: input.PowerSavingEnabled, WakeForAlerts: input.WakeForAlerts, DDCBus: input.DdcBus}}
	for _, setting := range input.Modes {
		schedule.Modes = append(schedule.Modes, display.ModeSetting{Mode: displayModeFromModel(setting.Mode), StartsAtMinutes: setting.StartsAtMinutes, Brightness: setting.Brightness})
	}
	return schedule
}

func displayStatusToModel(current display.Current) *model.DisplayStatus {
	status := &model.DisplayStatus{Mode: displayModeToModel(current.Mode), Brightness: current.Brightness, Powered: current.Power, OverrideActive: current.OverrideActive}
	if current.OverrideActive {
		ends := int(current.OverrideEndsAt.Unix())
		status.OverrideEndsAt = &ends
	}
	return status
}

func previewDuration(seconds int) time.Duration { return time.Duration(seconds) * time.Second }

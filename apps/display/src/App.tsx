import * as React from "react";
import { Clock } from "@homie/ui/components/clock";
import { ComingUp, ComingUpHandle, ComingUpSheet, ComingUpSwipeArea } from "@homie/ui/components/coming-up";
import { EventList, EventRow } from "@homie/ui/components/event-row";
import { HomeEvents, HomeLayout } from "@homie/ui/components/layout";
import { NextUp } from "@homie/ui/components/next-up";
import { PhotoPanel } from "@homie/ui/components/photo-panel";
import { WeatherGlance } from "@homie/ui/components/weather-glance";
import type { WeatherCondition } from "@homie/ui/components/weather-icon";
import { formatShortDate } from "@homie/ui/lib/time";
import { useNow } from "@homie/ui/hooks/use-now";
import { useWeather } from "./useWeather";
import { usePhotos } from "./usePhotos";
import { useEvents } from "./useEvents";
import { splitEvents, toRows } from "./events";

const SWIPE_AREA_ROWS = 4;

export function App() {
  const now = useNow();
  const weather = useWeather();
  const photos = usePhotos();
  const events = useEvents();
  const [sheetOpen, setSheetOpen] = React.useState(false);
  const eventsColumnRef = React.useRef<HTMLDivElement>(null);

  const allEvents = events.kind === "ready" ? events.events : [];
  const { headline, upcoming } = splitEvents(allEvents, now);
  const rows = toRows(upcoming, now);

  return (
    <HomeLayout>
      {photos.kind === "ready" && photos.photos.length > 0 && <PhotoPanel photos={photos.photos} />}
      <HomeEvents ref={eventsColumnRef}>
        <div className="flex flex-col gap-hm-8">
          <Clock date={now} />
          <div className="text-hm-lead text-hm-ink-2">{formatShortDate(now)}</div>
        </div>
        <NextUp label={headline?.label} title={headline?.event.title} people={headline?.event.people ?? []} detail={headline?.detail} />
        <ComingUp open={sheetOpen} onOpenChange={setSheetOpen}>
          <ComingUpSwipeArea>
            <ComingUpHandle />
            <EventList>
              {rows.slice(0, SWIPE_AREA_ROWS).map((row) => (
                <EventRow key={row.id} time={row.time} day={row.day} people={row.people} title={row.title} />
              ))}
            </EventList>
          </ComingUpSwipeArea>
          <ComingUpSheet container={eventsColumnRef}>
            <EventList>
              {rows.map((row) => (
                <EventRow key={row.id} time={row.time} day={row.day} people={row.people} title={row.title} compact />
              ))}
            </EventList>
          </ComingUpSheet>
        </ComingUp>
      </HomeEvents>
      {weather.kind === "ready" && (
        <WeatherGlance
          className="fixed right-hm-screen top-hm-screen z-10"
          condition={weatherIconCondition(weather.weather.condition)}
          temperature={weather.weather.temperature}
          stale={weather.weather.stale}
          aria-label={weatherAriaLabel(weather.weather.conditionCode, weather.weather.temperature, weather.weather.fetchedAt, weather.weather.stale)}
        />
      )}
    </HomeLayout>
  );
}

function weatherAriaLabel(conditionCode: string, temperature: number, fetchedAt: number, stale: boolean): string {
  const ageSeconds = Math.max(0, Math.floor(Date.now() / 1000) - fetchedAt);
  const age = ageSeconds < 60
    ? "less than a minute ago"
    : ageSeconds < 3600
      ? `${Math.floor(ageSeconds / 60)} minutes ago`
      : `${Math.floor(ageSeconds / 3600)} hours ago`;
  return `${conditionCode}, ${temperature} degrees, updated ${age}${stale ? ", stale weather data" : ""}`;
}

function weatherIconCondition(condition: "SUN" | "CLOUD_SUN" | "CLOUD" | "RAIN" | "MOON"): WeatherCondition {
  switch (condition) {
    case "SUN":
      return "sun";
    case "CLOUD_SUN":
      return "cloud-sun";
    case "CLOUD":
      return "cloud";
    case "RAIN":
      return "rain";
    case "MOON":
      return "moon";
  }
}

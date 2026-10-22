import * as React from "react";
import { Clock } from "@homie/ui/components/clock";
import { ComingUp, ComingUpHandle, ComingUpSheet, ComingUpSwipeArea } from "@homie/ui/components/coming-up";
import { EventList, EventRow } from "@homie/ui/components/event-row";
import { HomeEvents, HomeLayout } from "@homie/ui/components/layout";
import { NextUp } from "@homie/ui/components/next-up";
import { PhotoPanel } from "@homie/ui/components/photo-panel";
import { WeatherGlance } from "@homie/ui/components/weather-glance";
import { formatShortDate } from "@homie/ui/lib/time";
import { useNow } from "@homie/ui/hooks/use-now";
import { useWeather } from "./features/weather/use-weather";
import { weatherAriaLabel, weatherIconCondition } from "./features/weather/weather-labels";
import { usePhotos } from "./features/photos/use-photos";
import { usePhotoShowing } from "./features/photos/use-photo-showing";
import { useEvents } from "./features/events/use-events";
import { splitEvents, toRows } from "./features/events/calendar-event";
import { useTimer } from "./features/timer/use-timer";
import { TimerPage } from "./features/timer/timer-page";
import { useSetup } from "./features/setup/use-setup";
import { SetupPage } from "./features/setup/setup-page";

const SWIPE_AREA_ROWS = 4;

export function App() {
  const setup = useSetup();
  const timer = useTimer();
  const now = useNow();
  const weather = useWeather();
  const photos = usePhotos();
  const photoShowing = usePhotoShowing();
  const events = useEvents();
  const [sheetOpen, setSheetOpen] = React.useState(false);
  const eventsColumnRef = React.useRef<HTMLDivElement>(null);

  const allEvents = events.kind === "ready" ? events.events : [];
  const { headline, upcoming } = splitEvents(allEvents, now);
  const rows = toRows(upcoming, now);

  if (setup.kind === "setup") return <SetupPage code={setup.code} />;
  if (timer.timer) return <TimerPage timer={timer.timer} addMinute={timer.addMinute} togglePause={timer.togglePause} cancel={timer.cancel} />;

  return (
    <HomeLayout>
      {photos.kind === "ready" && photos.photos.length > 0 && photoShowing.showing?.current && (
        <PhotoPanel
          photos={[
            ...photos.photos,
            ...[photoShowing.showing.current, photoShowing.showing.next]
              .filter((showing): showing is NonNullable<typeof showing> => showing !== null)
              .map(({ photo }) => ({
                src: photo.url,
                alt: photo.place ?? "",
                ...(photo.place ? { place: photo.place } : {}),
                takenAt: new Date(photo.takenAt * 1000),
                ...(photo.focus ? { focus: photo.focus } : {}),
              })),
          ].filter((photo, index, all) => all.findIndex((candidate) => candidate.src === photo.src) === index)}
          currentSrc={photoShowing.showing.current.photo.url}
          onAdvance={() => void photoShowing.advance("TAP")}
        />
      )}
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

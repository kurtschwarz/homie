import { Clock } from "@homie/ui/components/clock";
import { HomeEvents, HomeLayout } from "@homie/ui/components/layout";
import { PhotoPanel, type Photo } from "@homie/ui/components/photo-panel";
import { WeatherGlance } from "@homie/ui/components/weather-glance";
import type { WeatherCondition } from "@homie/ui/components/weather-icon";
import { formatShortDate } from "@homie/ui/lib/time";
import { useNow } from "@homie/ui/hooks/use-now";
import { useWeather } from "./useWeather";

// Placeholder photos to preview the slideshow's crossfade and per-shape
// resizing until real ones come from the API. Remove once that lands.
const PLACEHOLDER_PHOTOS: Photo[] = [
  { src: "https://picsum.photos/seed/homie-1/1600/1200", alt: "", place: "Lake Tahoe, CA", takenAt: new Date("2026-08-14T20:12:00") },
  { src: "https://picsum.photos/seed/homie-2/1200/1600", alt: "", place: "Portland, OR" },
  { src: "https://picsum.photos/seed/homie-3/1400/1400", alt: "" },
  { src: "https://picsum.photos/seed/homie-4/1920/1080", alt: "", place: "Banff National Park", takenAt: new Date("2025-07-02T09:30:00") },
];

export function App() {
  const now = useNow();
  const weather = useWeather();

  return (
    <HomeLayout>
      <PhotoPanel photos={PLACEHOLDER_PHOTOS} intervalMs={4_000} />
      <HomeEvents>
        <div className="flex flex-col gap-hm-8">
          <Clock date={now} />
          <div className="text-hm-lead text-hm-ink-2">{formatShortDate(now)}</div>
        </div>
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

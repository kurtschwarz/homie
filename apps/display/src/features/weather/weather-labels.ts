import type { WeatherCondition } from "@homie/ui/components/weather-icon";

export function weatherAriaLabel(
  conditionCode: string,
  temperature: number,
  fetchedAt: number,
  stale: boolean,
): string {
  const ageSeconds = Math.max(0, Math.floor(Date.now() / 1000) - fetchedAt);
  const age =
    ageSeconds < 60
      ? "less than a minute ago"
      : ageSeconds < 3600
        ? `${Math.floor(ageSeconds / 60)} minutes ago`
        : `${Math.floor(ageSeconds / 3600)} hours ago`;
  return `${conditionCode}, ${temperature} degrees, updated ${age}${stale ? ", stale weather data" : ""}`;
}

export function weatherIconCondition(
  condition: "SUN" | "CLOUD_SUN" | "CLOUD" | "RAIN" | "MOON",
): WeatherCondition {
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

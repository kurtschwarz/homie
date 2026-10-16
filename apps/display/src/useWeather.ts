import { useEffect, useState } from "react";

type Weather = {
  condition: "SUN" | "CLOUD_SUN" | "CLOUD" | "RAIN" | "MOON";
  conditionCode: string;
  temperature: number;
};

type WeatherState =
  | { kind: "loading" }
  | { kind: "ready"; weather: Weather }
  | { kind: "error" };

const QUERY = `
  query Weather {
    weather {
      condition
      conditionCode
      temperature
    }
  }
`;

/** Fetches current conditions through GraphQL and refreshes on a ten-minute cadence. */
export function useWeather(): WeatherState {
  const [state, setState] = useState<WeatherState>({ kind: "loading" });

  useEffect(() => {
    let cancelled = false;
    let timeout: ReturnType<typeof setTimeout> | undefined;

    async function fetchWeather() {
      try {
        const response = await fetch("/graphql", {
          method: "POST",
          headers: { "Content-Type": "application/json" },
          body: JSON.stringify({ query: QUERY }),
        });
        if (!response.ok) throw new Error(`Weather request failed: ${response.status}`);
        const body = (await response.json()) as {
          data?: { weather?: Weather };
          errors?: unknown[];
        };
        if (body.errors?.length || !body.data?.weather) throw new Error("Weather query failed");
        if (!cancelled) setState({ kind: "ready", weather: body.data.weather });
      } catch {
        if (!cancelled) setState({ kind: "error" });
      } finally {
        if (!cancelled) timeout = setTimeout(fetchWeather, 10 * 60_000);
      }
    }

    void fetchWeather();
    return () => {
      cancelled = true;
      if (timeout) clearTimeout(timeout);
    };
  }, []);

  return state;
}

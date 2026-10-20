import * as React from "react";
import { TOKEN_KEY } from "./lib/graphql";
import { useHealth, type HealthState } from "./lib/use-health";
import { useSettings } from "./lib/use-settings";

type AppData = ReturnType<typeof useSettings> & {
  health: HealthState;
  token: string;
  saveToken: (token: string) => void;
};

const AppDataContext = React.createContext<AppData | null>(null);

/**
 * Settings, health and the API token, held once for the whole shell so the
 * sidebar badge and every page read the same copy — and so health is polled
 * once rather than per consumer. Saving the token reloads: every request
 * already in flight was made with the old one.
 */
export function AppDataProvider({ children }: { children: React.ReactNode }) {
  const settings = useSettings();
  const health = useHealth();
  const [token, setToken] = React.useState(localStorage.getItem(TOKEN_KEY) ?? "");

  const saveToken = React.useCallback((value: string) => {
    const cleaned = value.trim();
    localStorage.setItem(TOKEN_KEY, cleaned);
    setToken(cleaned);
    window.location.reload();
  }, []);

  const value = React.useMemo(
    () => ({ ...settings, health, token, saveToken }),
    [settings, health, token, saveToken],
  );
  return <AppDataContext.Provider value={value}>{children}</AppDataContext.Provider>;
}

export function useAppData() {
  const value = React.useContext(AppDataContext);
  if (!value) throw new Error("App data context is missing");
  return value;
}

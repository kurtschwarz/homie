import { StrictMode } from "react";
import { createRoot } from "react-dom/client";
import "./index.css";
import { HomieThemeProvider } from "@homie/ui/components/theme-provider";
import { App } from "./App";

// Lets index.css keep the pointer visible while you are developing; the
// kiosk build never sets this, so the panel stays cursorless.
if (import.meta.env.DEV) {
  document.documentElement.dataset.dev = "";
}

createRoot(document.getElementById("root")!).render(
  <StrictMode>
    <HomieThemeProvider theme="auto" surface="kiosk">
      <App />
    </HomieThemeProvider>
  </StrictMode>,
);

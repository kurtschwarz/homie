import { StrictMode } from "react";
import { createRoot } from "react-dom/client";
import "./index.css";
import { HomieThemeProvider } from "@homie/ui/components/theme-provider";
import { App } from "./App";

createRoot(document.getElementById("root")!).render(
  <StrictMode>
    <HomieThemeProvider theme="auto" surface="kiosk">
      <App />
    </HomieThemeProvider>
  </StrictMode>,
);

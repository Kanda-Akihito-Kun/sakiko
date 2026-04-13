import React from "react";
import ReactDOM from "react-dom/client";
import App from "./App";
import "./base.css";
import "./app.css";
import { AppThemeProvider } from "./theme/themeMode";
import { FALLBACK_LANGUAGE, initializeLanguage } from "./services/i18n";
import { preloadLanguage } from "./services/preload";

const root = ReactDOM.createRoot(document.getElementById("root") as HTMLElement);

function renderApp() {
  root.render(
    <React.StrictMode>
      <AppThemeProvider>
        <App />
      </AppThemeProvider>
    </React.StrictMode>,
  );
}

async function bootstrap() {
  const initialLanguage = await preloadLanguage();
  await initializeLanguage(initialLanguage);
  renderApp();
}

bootstrap().catch(async () => {
  await initializeLanguage(FALLBACK_LANGUAGE);
  renderApp();
});

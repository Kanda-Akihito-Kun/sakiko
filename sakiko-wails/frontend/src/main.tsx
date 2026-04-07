import React from "react";
import ReactDOM from "react-dom/client";
import App from "./App";
import "./base.css";
import "./app.css";
import { AppThemeProvider } from "./theme/themeMode";

ReactDOM.createRoot(document.getElementById('root') as HTMLElement).render(
  <React.StrictMode>
    <AppThemeProvider>
      <App />
    </AppThemeProvider>
  </React.StrictMode>,
)

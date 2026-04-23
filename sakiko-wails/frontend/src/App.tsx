import { DesktopNotificationCenter } from "./components/feedback/DesktopNotificationCenter";
import { AppErrorBoundary } from "./components/shared/AppErrorBoundary";
import { useDesktopNotifications } from "./hooks/useDesktopNotifications";
import { DashboardPage } from "./pages/DashboardPage";

function App() {
  useDesktopNotifications();

  return (
    <AppErrorBoundary title="Sakiko could not finish rendering.">
      <DashboardPage />
      <DesktopNotificationCenter />
    </AppErrorBoundary>
  );
}

export default App;

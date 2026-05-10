import { DesktopNotificationCenter } from "./components/feedback/DesktopNotificationCenter";
import { AppErrorBoundary } from "./components/shared/AppErrorBoundary";
import { useDesktopNotifications } from "./hooks/useDesktopNotifications";
import { DashboardView } from "./views/dashboard";

function App() {
  useDesktopNotifications();

  return (
    <AppErrorBoundary title="Sakiko could not finish rendering.">
      <DashboardView />
      <DesktopNotificationCenter />
    </AppErrorBoundary>
  );
}

export default App;

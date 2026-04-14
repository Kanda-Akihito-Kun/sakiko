import { DesktopNotificationCenter } from "./components/feedback/DesktopNotificationCenter";
import { useDesktopNotifications } from "./hooks/useDesktopNotifications";
import { DashboardPage } from "./pages/DashboardPage";

function App() {
  useDesktopNotifications();

  return (
    <>
      <DashboardPage />
      <DesktopNotificationCenter />
    </>
  );
}

export default App;

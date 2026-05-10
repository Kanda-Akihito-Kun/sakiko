import { Box } from "@mui/material";
import { useTranslation } from "react-i18next";
import { WorkspacePage } from "../../components/layout/WorkspacePage";
import { DashboardPageOutlet } from "./components/DashboardPageOutlet";
import { DashboardSidebar } from "./components/DashboardSidebar";
import { useDashboardViewModel } from "./useDashboardViewModel";

export function DashboardView() {
  const { t } = useTranslation();
  const view = useDashboardViewModel();

  return (
    <Box className="sakiko-shell">
      <Box className="sakiko-shell__backdrop" />

      <Box className="sakiko-workspace">
        <DashboardSidebar
          items={view.navItems}
          loading={view.loading}
          onSelect={view.setPage}
          selectedPage={view.page}
          title={t("dashboard.app.title")}
          workspaceReadyLabel={t("dashboard.workspace.ready")}
          workspaceSyncingLabel={t("dashboard.workspace.syncing")}
        />

        <Box className="sakiko-content">
          <WorkspacePage
            title={view.activeNav.label}
            action={view.pageAction}
            errorBoundaryKey={view.page}
          >
            <DashboardPageOutlet
              activePageLabel={view.activeNav.label}
              onOpenConfigs={() => view.setPage("configs")}
              onOpenProfiles={() => view.setPage("profiles")}
              page={view.page}
            />
          </WorkspacePage>
        </Box>
      </Box>
    </Box>
  );
}

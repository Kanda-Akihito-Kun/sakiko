import RefreshRounded from "@mui/icons-material/RefreshRounded";
import { Button } from "@mui/material";
import { useShallow } from "zustand/react/shallow";
import { startTransition, useMemo, useState } from "react";
import { useTranslation } from "react-i18next";
import { useDashboardLifecycle } from "../../hooks/useDashboardLifecycle";
import { useDashboardStore } from "../../store/dashboardStore";
import { dashboardNavItems, type DashboardPageKey } from "./model";

const dashboardPageStorageKey = "sakiko.dashboard.page";

export function useDashboardViewModel() {
  useDashboardLifecycle();

  const [page, setPage] = useState<DashboardPageKey>(() => readStoredDashboardPage());
  const { t } = useTranslation();
  const dashboard = useDashboardStore(useShallow((state) => ({
    activeProfileId: state.activeProfileId,
    loading: state.loading,
    refreshDashboard: state.refreshDashboard,
    refreshRemoteStatus: state.refreshRemoteStatus,
    refreshRemoteWorkspace: state.refreshRemoteWorkspace,
  })));

  const navItems = useMemo(() => dashboardNavItems.map((item) => ({
    ...item,
    label: t(item.labelKey),
  })), [t]);

  const activeNav = navItems.find((item) => item.id === page) || navItems[0];

  const pageAction = (
    <Button
      variant="outlined"
      startIcon={<RefreshRounded />}
      disabled={dashboard.loading}
      onClick={() => {
        void dashboard.refreshDashboard(dashboard.activeProfileId);
        if (page === "remote") {
          void dashboard.refreshRemoteWorkspace(false);
          return;
        }
        void dashboard.refreshRemoteStatus();
      }}
    >
      {t("dashboard.workspace.refreshWorkspace")}
    </Button>
  );

  return {
    activeNav,
    loading: dashboard.loading,
    navItems,
    pageAction,
    page,
    setPage: (nextPage: DashboardPageKey) => {
      writeStoredDashboardPage(nextPage);
      startTransition(() => setPage(nextPage));
    },
  };
}

function readStoredDashboardPage(): DashboardPageKey {
  if (typeof window === "undefined") {
    return "configs";
  }
  const raw = window.sessionStorage.getItem(dashboardPageStorageKey);
  switch (raw) {
    case "overview":
      writeStoredDashboardPage("configs");
      return "configs";
    case "configs":
    case "profiles":
    case "tasks":
    case "results":
    case "remote":
    case "settings":
      return raw;
    default:
      return "configs";
  }
}

function writeStoredDashboardPage(page: DashboardPageKey) {
  if (typeof window === "undefined") {
    return;
  }
  window.sessionStorage.setItem(dashboardPageStorageKey, page);
}

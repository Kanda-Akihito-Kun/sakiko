import HomeRounded from "@mui/icons-material/HomeRounded";
import HubRounded from "@mui/icons-material/HubRounded";
import InsightsRounded from "@mui/icons-material/InsightsRounded";
import RefreshRounded from "@mui/icons-material/RefreshRounded";
import SettingsRounded from "@mui/icons-material/SettingsRounded";
import StorageRounded from "@mui/icons-material/StorageRounded";
import TuneRounded from "@mui/icons-material/TuneRounded";
import {
  Box,
  Button,
  Chip,
  CircularProgress,
  List,
  ListItemButton,
  ListItemIcon,
  ListItemText,
  Stack,
  Typography,
} from "@mui/material";
import { lazy, Suspense, startTransition, useState } from "react";
import { useTranslation } from "react-i18next";
import { WorkspacePage } from "../components/layout/WorkspacePage";
import { APP_VERSION } from "../constants/appMeta";
import { useShallow } from "zustand/react/shallow";
import { useDashboardLifecycle } from "../hooks/useDashboardLifecycle";
import { useDashboardStore } from "../store/dashboardStore";
import { useThemeMode } from "../theme/themeMode";

const OverviewSection = lazy(() => import("../sections/OverviewSection").then((module) => ({ default: module.OverviewSection })));
const ProfilesSection = lazy(() => import("../sections/ProfilesSection").then((module) => ({ default: module.ProfilesSection })));
const TasksSection = lazy(() => import("../sections/TasksSection").then((module) => ({ default: module.TasksSection })));
const ResultsSection = lazy(() => import("../sections/ResultsSection").then((module) => ({ default: module.ResultsSection })));
const SettingsSection = lazy(() => import("../sections/SettingsSection").then((module) => ({ default: module.SettingsSection })));

export type DashboardSection = "overview" | "profiles" | "tasks" | "results" | "settings";

type NavItem = {
  id: DashboardSection;
  icon: typeof HomeRounded;
};

const navItems: NavItem[] = [
  { id: "overview", icon: HomeRounded },
  { id: "profiles", icon: StorageRounded },
  { id: "tasks", icon: TuneRounded },
  { id: "results", icon: InsightsRounded },
  { id: "settings", icon: SettingsRounded },
];

export function DashboardPage() {
  useDashboardLifecycle();

  const [section, setSection] = useState<DashboardSection>("overview");
  const { mode, resolvedMode } = useThemeMode();
  const { t } = useTranslation();
  const dashboard = useDashboardStore(useShallow((state) => ({
    activeProfile: state.activeProfile,
    activeProfileId: state.activeProfileId,
    activeTask: state.activeTask,
    activeTaskId: state.activeTaskId,
    ensureResultArchive: state.ensureResultArchive,
    downloadTargets: state.downloadTargets,
    downloadTargetSearch: state.downloadTargetSearch,
    downloadTargetsLoading: state.downloadTargetsLoading,
    error: state.error,
    handleDeleteResultArchive: state.handleDeleteResultArchive,
    handleDeleteProfile: state.handleDeleteProfile,
    handleDeleteTask: state.handleDeleteTask,
    handleImport: state.handleImport,
    handleInspectTask: state.handleInspectTask,
    handleProfileSelect: state.handleProfileSelect,
    handleRefreshProfile: state.handleRefreshProfile,
    handleMoveProfileNode: state.handleMoveProfileNode,
    handleSetProfileNodeEnabled: state.handleSetProfileNodeEnabled,
    handleRunTask: state.handleRunTask,
    handleStopTask: state.handleStopTask,
    importForm: state.importForm,
    loading: state.loading,
    message: state.message,
    nodeFilter: state.nodeFilter,
    patchTaskConfig: state.patchTaskConfig,
    profiles: state.profiles,
    profilesPath: state.profilesPath,
    refreshResultArchives: state.refreshResultArchives,
    refreshDashboard: state.refreshDashboard,
    refreshDownloadTargets: state.refreshDownloadTargets,
    resultArchiveDetails: state.resultArchiveDetails,
    resultArchiveLoading: state.resultArchiveLoading,
    resultArchives: state.resultArchives,
    resultArchivesLoading: state.resultArchivesLoading,
    resultArchivesVisibleCount: state.resultArchivesVisibleCount,
    loadMoreResultArchives: state.loadMoreResultArchives,
    setDownloadTargetSearch: state.setDownloadTargetSearch,
    setNodeFilter: state.setNodeFilter,
    setTaskPreset: state.setTaskPreset,
    submitting: state.submitting,
    taskConfig: state.taskConfig,
    taskPreset: state.taskPreset,
    tasks: state.tasks,
    updateImportForm: state.updateImportForm,
  })));
  const localizedNavItems = navItems.map((item) => ({
    ...item,
    label: t(`dashboard.nav.${item.id}.label`),
  }));
  const activeNav = localizedNavItems.find((item) => item.id === section) || localizedNavItems[0];
  const pageAction = section === "settings" || section === "overview" ? undefined : (
    <Stack direction="row" spacing={1.25} alignItems="center" sx={{ minWidth: 0, flex: "0 0 auto" }}>
      <Chip
        label={dashboard.activeProfile?.name || t("dashboard.workspace.noActiveProfile")}
        icon={<HubRounded />}
        variant="outlined"
        sx={{
          maxWidth: 280,
          minWidth: 0,
          "& .MuiChip-label": {
            overflow: "hidden",
            textOverflow: "ellipsis",
            whiteSpace: "nowrap",
          },
        }}
      />
      <Button
        variant="outlined"
        startIcon={<RefreshRounded />}
        disabled={dashboard.loading}
        onClick={() => void dashboard.refreshDashboard(dashboard.activeProfileId)}
      >
        {t("dashboard.workspace.refreshWorkspace")}
      </Button>
    </Stack>
  );

  return (
    <Box className="sakiko-shell">
      <Box className="sakiko-shell__backdrop" />

      <Box className="sakiko-workspace">
        <Box component="aside" className="sakiko-sidebar">
          <Stack className="sakiko-sidebar__header" spacing={1.25}>
            <Stack direction="row" spacing={1.25} alignItems="center">
              <Box className="sakiko-sidebar__logo">
                <Box
                  component="img"
                  className="sakiko-sidebar__logo-img"
                  src="/sakiko.png"
                  alt=""
                />
              </Box>
              <Box>
                <Typography variant="h5" fontWeight={800}>{t("dashboard.app.title")}</Typography>
              </Box>
            </Stack>

            <Chip
              label={dashboard.loading ? t("dashboard.workspace.syncing") : t("dashboard.workspace.ready")}
              color={dashboard.loading ? "warning" : "success"}
              size="small"
              sx={{ alignSelf: "flex-start" }}
            />
          </Stack>

          <List disablePadding className="sakiko-sidebar__nav">
            {localizedNavItems.map((item) => {
              const Icon = item.icon;
              return (
                <ListItemButton
                  key={item.id}
                  selected={item.id === section}
                  onClick={() => startTransition(() => setSection(item.id))}
                  sx={{ px: 1.25, py: 1.25 }}
                >
                  <ListItemIcon sx={{ minWidth: 38 }}>
                    <Icon fontSize="small" />
                  </ListItemIcon>
                  <ListItemText
                    primary={item.label}
                    primaryTypographyProps={{ fontWeight: 700, noWrap: true }}
                  />
                </ListItemButton>
              );
            })}
          </List>

          <Box className="sakiko-sidebar__meta">
            <Typography variant="caption" color="text.secondary" className="sakiko-sidebar__version">
              {APP_VERSION}
            </Typography>
          </Box>

        </Box>

        <Box className="sakiko-content">
          <WorkspacePage
            title={activeNav.label}
            action={pageAction}
            errorBoundaryKey={section}
          >
            <Suspense fallback={<SectionLoadingFallback label={activeNav.label} />}>
              {section === "overview" && (
                <OverviewSection
                  downloadTargets={dashboard.downloadTargets}
                  downloadTargetSearch={dashboard.downloadTargetSearch}
                  downloadTargetsLoading={dashboard.downloadTargetsLoading}
                  taskConfig={dashboard.taskConfig}
                  onPatchTaskConfig={dashboard.patchTaskConfig}
                  onDownloadTargetSearchChange={dashboard.setDownloadTargetSearch}
                  onRefreshDownloadTargets={dashboard.refreshDownloadTargets}
                />
              )}

              {section === "profiles" && (
                <ProfilesSection
                  activeProfile={dashboard.activeProfile}
                  activeProfileId={dashboard.activeProfileId}
                  importForm={dashboard.importForm}
                  loading={dashboard.loading}
                  profiles={dashboard.profiles}
                  submitting={dashboard.submitting}
                  onDeleteProfile={dashboard.handleDeleteProfile}
                  onImport={dashboard.handleImport}
                  onImportFormChange={dashboard.updateImportForm}
                  onNodeEnabledChange={dashboard.handleSetProfileNodeEnabled}
                  onNodeMove={dashboard.handleMoveProfileNode}
                  onRefreshProfile={dashboard.handleRefreshProfile}
                  onReload={dashboard.refreshDashboard}
                  onSelectProfile={dashboard.handleProfileSelect}
                />
              )}

              {section === "tasks" && (
                <TasksSection
                  activeProfileId={dashboard.activeProfileId}
                  activeTask={dashboard.activeTask}
                  activeTaskId={dashboard.activeTaskId}
                  submitting={dashboard.submitting}
                  taskConfig={dashboard.taskConfig}
                  taskPreset={dashboard.taskPreset}
                  tasks={dashboard.tasks}
                  onInspectTask={dashboard.handleInspectTask}
                  onDeleteTask={dashboard.handleDeleteTask}
                  onOpenConfigs={() => startTransition(() => setSection("overview"))}
                  onRunTask={dashboard.handleRunTask}
                  onStopTask={dashboard.handleStopTask}
                  onTaskPresetChange={dashboard.setTaskPreset}
                />
              )}

              {section === "results" && (
                <ResultsSection
                  archiveDetails={dashboard.resultArchiveDetails}
                  archiveLoading={dashboard.resultArchiveLoading}
                  archives={dashboard.resultArchives}
                  downloadTargets={dashboard.downloadTargets}
                  loading={dashboard.resultArchivesLoading}
                  visibleCount={dashboard.resultArchivesVisibleCount}
                  onDeleteArchive={dashboard.handleDeleteResultArchive}
                  onEnsureArchive={dashboard.ensureResultArchive}
                  onLoadMore={dashboard.loadMoreResultArchives}
                  onRefresh={dashboard.refreshResultArchives}
                />
              )}

              {section === "settings" && (
                <SettingsSection
                  mode={mode}
                  profilesPath={dashboard.profilesPath}
                  resolvedMode={resolvedMode}
                />
              )}
            </Suspense>
          </WorkspacePage>
        </Box>
      </Box>
    </Box>
  );
}

type SectionLoadingFallbackProps = {
  label: string;
};

function SectionLoadingFallback({ label }: SectionLoadingFallbackProps) {
  const { t } = useTranslation();

  return (
    <Stack spacing={1.5} alignItems="center" justifyContent="center" sx={{ minHeight: 240 }}>
      <CircularProgress size={28} />
      <Typography variant="body2" color="text.secondary">
        {t("shared.states.loadingSection", { label })}
      </Typography>
    </Stack>
  );
}

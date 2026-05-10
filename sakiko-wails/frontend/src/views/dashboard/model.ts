import InsightsRounded from "@mui/icons-material/InsightsRounded";
import LanRounded from "@mui/icons-material/LanRounded";
import PlayCircleOutlineRounded from "@mui/icons-material/PlayCircleOutlineRounded";
import SettingsRounded from "@mui/icons-material/SettingsRounded";
import StorageRounded from "@mui/icons-material/StorageRounded";
import TuneRounded from "@mui/icons-material/TuneRounded";

export type DashboardPageKey = "configs" | "profiles" | "tasks" | "results" | "remote" | "settings";

export type DashboardNavItem = {
  id: DashboardPageKey;
  icon: typeof TuneRounded;
  labelKey: string;
};

export const dashboardNavItems: DashboardNavItem[] = [
  { id: "configs", icon: TuneRounded, labelKey: "dashboard.nav.configs.label" },
  { id: "profiles", icon: StorageRounded, labelKey: "dashboard.nav.profiles.label" },
  { id: "tasks", icon: PlayCircleOutlineRounded, labelKey: "dashboard.nav.tasks.label" },
  { id: "results", icon: InsightsRounded, labelKey: "dashboard.nav.results.label" },
  { id: "remote", icon: LanRounded, labelKey: "dashboard.nav.remote.label" },
  { id: "settings", icon: SettingsRounded, labelKey: "dashboard.nav.settings.label" },
];

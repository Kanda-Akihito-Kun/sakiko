import { lazy, Suspense, type ComponentType } from "react";
import type { DashboardPageKey } from "../model";
import { SectionLoadingFallback } from "./SectionLoadingFallback";

const ConfigsDashboardSection = lazyWithReload(() => import("../sections/configs"), "configs");
const ProfilesDashboardSection = lazyWithReload(() => import("../sections/profiles"), "profiles");
const TasksDashboardSection = lazyWithReload(() => import("../sections/tasks"), "tasks");
const ResultsDashboardSection = lazyWithReload(() => import("../sections/results"), "results");
const RemoteDashboardSection = lazyWithReload(() => import("../sections/remote"), "remote");
const SettingsDashboardSection = lazyWithReload(() => import("../sections/settings"), "settings");

type DashboardPageOutletProps = {
  activePageLabel: string;
  onOpenConfigs: () => void;
  onOpenProfiles: () => void;
  page: DashboardPageKey;
};

export function DashboardPageOutlet({
  activePageLabel,
  onOpenConfigs,
  onOpenProfiles,
  page,
}: DashboardPageOutletProps) {
  return (
    <Suspense fallback={<SectionLoadingFallback label={activePageLabel} />}>
      {page === "configs" && <ConfigsDashboardSection />}
      {page === "profiles" && <ProfilesDashboardSection />}
      {page === "tasks" && <TasksDashboardSection onOpenConfigs={onOpenConfigs} />}
      {page === "results" && <ResultsDashboardSection />}
      {page === "remote" && <RemoteDashboardSection onOpenProfiles={onOpenProfiles} onOpenConfigs={onOpenConfigs} />}
      {page === "settings" && <SettingsDashboardSection />}
    </Suspense>
  );
}

function lazyWithReload<TModule extends { default: ComponentType<any> }>(
  importer: () => Promise<TModule>,
  key: string,
) {
  return lazy(async () => {
    try {
      return await importer();
    } catch (error) {
      if (shouldReloadLazyImport(error, key)) {
        window.sessionStorage.setItem(lazyImportReloadKey(key), "1");
        window.location.reload();
      }
      throw error;
    }
  });
}

function shouldReloadLazyImport(error: unknown, key: string) {
  if (typeof window === "undefined") {
    return false;
  }
  if (window.sessionStorage.getItem(lazyImportReloadKey(key)) === "1") {
    window.sessionStorage.removeItem(lazyImportReloadKey(key));
    return false;
  }
  const message = error instanceof Error ? error.message : String(error || "");
  return message.includes("Failed to fetch dynamically imported module");
}

function lazyImportReloadKey(key: string) {
  return `sakiko.lazy-reload.${key}`;
}

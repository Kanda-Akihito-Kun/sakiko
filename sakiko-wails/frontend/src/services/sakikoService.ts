import { Call, type CancellablePromise } from "@wailsio/runtime";
import { DownloadTarget, Profile } from "../../bindings/sakiko.local/sakiko-core/interfaces";
import { SakikoService as GeneratedSakikoService } from "../../bindings/sakiko-wails";
import type { AppSettings, AppSettingsPatch } from "../types/appSettings";
import type { ReleaseCheckResult } from "../types/release";

export const SakikoService = {
  ...GeneratedSakikoService,
  CancelTask(taskID: string): CancellablePromise<void> {
    return Call.ByName("main.SakikoService.CancelTask", taskID);
  },
  DeleteTask(taskID: string): CancellablePromise<void> {
    return Call.ByName("main.SakikoService.DeleteTask", taskID);
  },
  DeleteResultArchive(taskID: string): CancellablePromise<void> {
    return Call.ByName("main.SakikoService.DeleteResultArchive", taskID);
  },
  SetProfileNodeEnabled(profileID: string, nodeIndex: number, enabled: boolean): CancellablePromise<Profile> {
    return Call.ByName("main.SakikoService.SetProfileNodeEnabled", profileID, nodeIndex, enabled).then((result: any) => {
      return Profile.createFrom(result);
    });
  },
  MoveProfileNode(profileID: string, nodeIndex: number, targetIndex: number): CancellablePromise<Profile> {
    return Call.ByName("main.SakikoService.MoveProfileNode", profileID, nodeIndex, targetIndex).then((result: any) => {
      return Profile.createFrom(result);
    });
  },
  SearchDownloadTargets(search: string): CancellablePromise<DownloadTarget[]> {
    return Call.ByName("main.SakikoService.SearchDownloadTargets", search).then((result: any) => {
      return Array.isArray(result) ? result.map((item) => DownloadTarget.createFrom(item)) : [];
    });
  },
  GetAppSettings(): CancellablePromise<AppSettings> {
    return Call.ByName("main.SakikoService.GetAppSettings");
  },
  UpdateAppSettings(patch: AppSettingsPatch): CancellablePromise<AppSettings> {
    return Call.ByName("main.SakikoService.UpdateAppSettings", patch);
  },
  CheckForUpdates(): CancellablePromise<ReleaseCheckResult> {
    return Call.ByName("main.SakikoService.CheckForUpdates");
  },
  OpenReleasePage(url?: string): CancellablePromise<void> {
    return Call.ByName("main.SakikoService.OpenReleasePage", url || "");
  },
};

import { Call, type CancellablePromise } from "@wailsio/runtime";
import { DownloadTarget, Profile, type TaskConfig } from "../../bindings/sakiko.local/sakiko-core/interfaces/index.js";
import { SakikoService as GeneratedSakikoService } from "../../bindings/sakiko-wails/index.js";
import type { AppSettings, AppSettingsPatch } from "../types/appSettings";
import type { ReleaseCheckResult } from "../types/release";
import type { ClusterConnectedKnight, ClusterPairingCode, ClusterRemoteTask, ClusterStatus, MasterEligibility, ResultArchive, ResultArchiveListItem } from "../types/sakiko";

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
  GetRemoteStatus(): CancellablePromise<ClusterStatus> {
    return Call.ByName("main.SakikoService.GetRemoteStatus");
  },
  ProbeRemoteMasterEligibility(): CancellablePromise<MasterEligibility> {
    return Call.ByName("main.SakikoService.ProbeRemoteMasterEligibility");
  },
  EnableRemoteMaster(listenHost: string, listenPort: number): CancellablePromise<ClusterStatus> {
    return Call.ByName("main.SakikoService.EnableRemoteMaster", listenHost, listenPort);
  },
  CreateRemotePairingCode(knightName: string, ttlSeconds: number): CancellablePromise<ClusterPairingCode> {
    return Call.ByName("main.SakikoService.CreateRemotePairingCode", knightName, ttlSeconds);
  },
  EnableRemoteKnight(masterHost: string, masterPort: number, oneTimeCode: string): CancellablePromise<ClusterStatus> {
    return Call.ByName("main.SakikoService.EnableRemoteKnight", masterHost, masterPort, oneTimeCode);
  },
  DisableRemoteMode(): CancellablePromise<ClusterStatus> {
    return Call.ByName("main.SakikoService.DisableRemoteMode");
  },
  ListRemoteKnights(): CancellablePromise<ClusterConnectedKnight[]> {
    return Call.ByName("main.SakikoService.ListRemoteKnights");
  },
  KickRemoteKnight(knightId: string): CancellablePromise<ClusterStatus> {
    return Call.ByName("main.SakikoService.KickRemoteKnight", knightId);
  },
  ListRemoteTasks(): CancellablePromise<ClusterRemoteTask[]> {
    return Call.ByName("main.SakikoService.ListRemoteTasks");
  },
  SubmitRemoteProfileTask(request: {
    profileId: string;
    knightIds: string[];
    name?: string;
    preset: string;
    presets?: string[];
    config?: TaskConfig;
  }): CancellablePromise<ClusterRemoteTask[]> {
    return Call.ByName("main.SakikoService.SubmitRemoteProfileTask", request);
  },
  ListRemoteMasterResultArchives(): CancellablePromise<ResultArchiveListItem[]> {
    return Call.ByName("main.SakikoService.ListRemoteMasterResultArchives");
  },
  GetRemoteMasterResultArchive(taskID: string): CancellablePromise<ResultArchive> {
    return Call.ByName("main.SakikoService.GetRemoteMasterResultArchive", taskID);
  },
  DeleteRemoteMasterResultArchive(taskID: string): CancellablePromise<void> {
    return Call.ByName("main.SakikoService.DeleteRemoteMasterResultArchive", taskID);
  },
  ListRemoteKnightResultArchives(): CancellablePromise<ResultArchiveListItem[]> {
    return Call.ByName("main.SakikoService.ListRemoteKnightResultArchives");
  },
  GetRemoteKnightResultArchive(taskID: string): CancellablePromise<ResultArchive> {
    return Call.ByName("main.SakikoService.GetRemoteKnightResultArchive", taskID);
  },
  DeleteRemoteKnightResultArchive(taskID: string): CancellablePromise<void> {
    return Call.ByName("main.SakikoService.DeleteRemoteKnightResultArchive", taskID);
  },
};

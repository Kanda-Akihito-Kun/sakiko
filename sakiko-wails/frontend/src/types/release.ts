export type ReleaseCheckResult = {
  currentVersion: string;
  latestVersion?: string;
  releaseName?: string;
  releaseURL?: string;
  releaseNotes?: string;
  hasUpdate: boolean;
  checkedAt?: string;
};

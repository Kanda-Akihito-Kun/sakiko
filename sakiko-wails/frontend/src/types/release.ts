export type ReleaseCheckResult = {
  currentVersion: string;
  latestVersion?: string;
  releaseName?: string;
  releaseURL?: string;
  hasUpdate: boolean;
  checkedAt?: string;
};

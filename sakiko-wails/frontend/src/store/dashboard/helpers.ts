import { initialTaskConfig, resolveBackendIdentity, sanitizeBackendIdentity } from "../../constants/dashboard";
import type { Profile, ProfileSummary, TaskConfig } from "../../types/sakiko";
import type { TaskConfigPatch } from "./types";

const BACKEND_IDENTITY_STORAGE_KEY = "sakiko.task-defaults.backend-identity";

export function createInitialTaskConfig(): TaskConfig {
  return {
    ...initialTaskConfig,
    backendIdentity: resolveBackendIdentity(getStoredBackendIdentity()),
  };
}

export function normalizeTaskConfigPatch(patch: TaskConfigPatch): TaskConfigPatch {
  const nextPatch = { ...patch };

  if (typeof nextPatch.taskTimeoutMillis === "number") {
    if (!Number.isFinite(nextPatch.taskTimeoutMillis)) {
      delete nextPatch.taskTimeoutMillis;
    } else {
      nextPatch.taskTimeoutMillis = Math.max(1, Math.floor(nextPatch.taskTimeoutMillis));
    }
  }

  if (typeof nextPatch.downloadDuration === "number" && Number.isFinite(nextPatch.downloadDuration)) {
    nextPatch.downloadDuration = Math.max(0, Math.floor(nextPatch.downloadDuration));
  } else if (typeof nextPatch.downloadDuration === "number") {
    delete nextPatch.downloadDuration;
  }

  if (typeof nextPatch.downloadThreading === "number") {
    if (!Number.isFinite(nextPatch.downloadThreading)) {
      delete nextPatch.downloadThreading;
    } else {
      nextPatch.downloadThreading = Math.max(1, Math.floor(nextPatch.downloadThreading));
    }
  }

  if (typeof nextPatch.backendIdentity === "string") {
    nextPatch.backendIdentity = sanitizeBackendIdentity(nextPatch.backendIdentity);
    persistBackendIdentity(nextPatch.backendIdentity);
  }

  return nextPatch;
}

export function resolveActiveProfileId(profiles: ProfileSummary[], preferredProfileId?: string): string {
  if (preferredProfileId && profiles.some((profile) => profile.id === preferredProfileId)) {
    return preferredProfileId;
  }
  return profiles[0]?.id || "";
}

export function upsertProfileSummary(profiles: ProfileSummary[], nextProfile: Profile): ProfileSummary[] {
  const nextSummary = toProfileSummary(nextProfile);
  const nextProfiles = profiles.slice();
  const targetIndex = nextProfiles.findIndex((profile) => profile.id === nextSummary.id);
  if (targetIndex >= 0) {
    nextProfiles[targetIndex] = nextSummary;
    return nextProfiles;
  }
  return [nextSummary, ...nextProfiles];
}

export function removeProfileSummary(profiles: ProfileSummary[], profileID: string): ProfileSummary[] {
  return profiles.filter((profile) => profile.id !== profileID);
}

function getStoredBackendIdentity(): string {
  if (typeof window === "undefined") {
    return resolveBackendIdentity();
  }

  const raw = window.localStorage.getItem(BACKEND_IDENTITY_STORAGE_KEY) || "";
  return resolveBackendIdentity(raw);
}

function persistBackendIdentity(value: string) {
  if (typeof window === "undefined") {
    return;
  }

  const normalized = sanitizeBackendIdentity(value);
  if (!normalized) {
    window.localStorage.removeItem(BACKEND_IDENTITY_STORAGE_KEY);
    return;
  }

  window.localStorage.setItem(BACKEND_IDENTITY_STORAGE_KEY, normalized);
}

function toProfileSummary(profile: Profile): ProfileSummary {
  const attributes = profile.attributes as Record<string, unknown> | undefined;
  const subscription = attributes && typeof attributes === "object"
    ? (attributes.subscriptionUserinfo as Record<string, unknown> | undefined)
    : undefined;

  return {
    id: profile.id,
    name: profile.name,
    source: profile.source,
    updatedAt: profile.updatedAt,
    nodeCount: profile.nodes.length,
    remainingBytes: typeof subscription?.remaining === "number" ? subscription.remaining : undefined,
    expiresAt: typeof subscription?.expiresAt === "string" ? subscription.expiresAt : undefined,
  };
}

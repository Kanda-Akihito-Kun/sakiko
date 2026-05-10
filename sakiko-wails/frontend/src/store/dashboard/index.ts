import { create } from "zustand";
import { createDashboardCoreSlice } from "./slices/coreSlice";
import { createDashboardProfilesSlice } from "./slices/profilesSlice";
import { createDashboardRemoteSlice } from "./slices/remoteSlice";
import { createDashboardResultsSlice } from "./slices/resultsSlice";
import { createDashboardTasksSlice } from "./slices/tasksSlice";
import type { DashboardState } from "./types";

export const useDashboardStore = create<DashboardState>()((...args) => ({
  ...createDashboardCoreSlice(...args),
  ...createDashboardProfilesSlice(...args),
  ...createDashboardTasksSlice(...args),
  ...createDashboardResultsSlice(...args),
  ...createDashboardRemoteSlice(...args),
}));

export type { DashboardState, TaskConfigPatch } from "./types";

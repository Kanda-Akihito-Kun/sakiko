import type { ProfileImportRequest } from "../../../../bindings/sakiko.local/sakiko-core/interfaces/index.js";
import { initialImportForm } from "../../../constants/dashboard";
import { SakikoService } from "../../../services/sakikoService";
import { translate } from "../../../services/i18n";
import { createImportProfilePayload } from "../../dashboardPayloads";
import { normalizeError } from "../../../utils/dashboard";
import { removeProfileSummary, upsertProfileSummary } from "../helpers";
import type { DashboardProfilesSlice, DashboardSliceCreator } from "../types";

export const createDashboardProfilesSlice: DashboardSliceCreator<DashboardProfilesSlice> = (set, get) => ({
  profiles: [],
  activeProfileId: "",
  activeProfile: null,

  handleProfileSelect: async (profileId) => {
    set({
      activeProfileId: profileId,
      activeProfile: get().activeProfile?.id === profileId ? get().activeProfile : null,
      error: "",
    });

    try {
      const profile = await SakikoService.GetProfile(profileId);
      set({ activeProfile: profile });
    } catch (err) {
      set({ error: normalizeError(err) });
    }
  },

  handleImport: async () => {
    const { importForm, refreshDashboard } = get();

    set({
      submitting: true,
      error: "",
      message: translate("dashboard.messages.importingProfile", "Importing profile..."),
    });

    try {
      const request: ProfileImportRequest = createImportProfilePayload(importForm);
      const profile = await SakikoService.ImportProfile(request);

      set((state) => ({
        importForm: initialImportForm,
        activeProfileId: profile.id,
        activeProfile: profile,
        profiles: upsertProfileSummary(state.profiles, profile),
      }));

      await refreshDashboard(profile.id);
      set({
        message: translate("dashboard.messages.importedProfile", `Imported ${profile.name} (${profile.nodes.length} nodes).`, {
          name: profile.name,
          count: profile.nodes.length,
        }),
      });
    } catch (err) {
      set({ error: normalizeError(err) });
    } finally {
      set({ submitting: false });
    }
  },

  handleRefreshProfile: async () => {
    const { activeProfile, activeProfileId, refreshDashboard } = get();
    if (!activeProfileId || !activeProfile?.source?.trim()) {
      return;
    }

    set({
      submitting: true,
      error: "",
      message: translate("dashboard.messages.refreshingProfile", "Refreshing profile..."),
    });

    try {
      const profile = await SakikoService.RefreshProfile(activeProfileId);
      set((state) => ({
        activeProfile: profile,
        profiles: upsertProfileSummary(state.profiles, profile),
      }));
      await refreshDashboard(profile.id);
      set({ message: translate("dashboard.messages.refreshedProfile", `Refreshed ${profile.name}.`, { name: profile.name }) });
    } catch (err) {
      set({ error: normalizeError(err) });
    } finally {
      set({ submitting: false });
    }
  },

  handleDeleteProfile: async () => {
    const { activeProfile, activeProfileId, profiles, refreshDashboard } = get();
    if (!activeProfileId) {
      return;
    }

    set({
      submitting: true,
      error: "",
      message: translate("dashboard.messages.deletingProfile", `Deleting ${activeProfile?.name || "profile"}...`, {
        name: activeProfile?.name || translate("shared.states.profile", "profile"),
      }),
    });

    try {
      await SakikoService.DeleteProfile(activeProfileId);
      const nextProfiles = removeProfileSummary(profiles, activeProfileId);
      const nextActiveProfileId = nextProfiles[0]?.id || "";

      set({
        profiles: nextProfiles,
        activeProfileId: nextActiveProfileId,
        activeProfile: null,
        nodeFilter: "",
      });

      await refreshDashboard(nextActiveProfileId);
      set({
        message: translate("dashboard.messages.deletedProfile", `Deleted ${activeProfile?.name || "profile"}.`, {
          name: activeProfile?.name || translate("shared.states.profile", "profile"),
        }),
      });
    } catch (err) {
      set({ error: normalizeError(err) });
    } finally {
      set({ submitting: false });
    }
  },

  handleSetProfileNodeEnabled: async (nodeIndex, enabled) => {
    const { activeProfile, activeProfileId } = get();
    if (!activeProfileId || !activeProfile || nodeIndex < 0 || nodeIndex >= activeProfile.nodes.length) {
      return;
    }

    const nodeName = activeProfile.nodes[nodeIndex]?.name || translate("shared.formats.nodeNumberLower", `node ${nodeIndex + 1}`, { index: nodeIndex + 1 });
    set({
      submitting: true,
      error: "",
      message: enabled
        ? translate("dashboard.messages.includingNode", `Including ${nodeName}...`, { name: nodeName })
        : translate("dashboard.messages.skippingNode", `Skipping ${nodeName}...`, { name: nodeName }),
    });

    try {
      const profile = await SakikoService.SetProfileNodeEnabled(activeProfileId, nodeIndex, enabled);
      set((state) => ({
        activeProfile: profile,
        profiles: upsertProfileSummary(state.profiles, profile),
        message: enabled
          ? translate("dashboard.messages.includedNode", `Included ${nodeName} for future tasks.`, { name: nodeName })
          : translate("dashboard.messages.skippedNode", `Skipped ${nodeName} for future tasks.`, { name: nodeName }),
      }));
    } catch (err) {
      set({ error: normalizeError(err) });
    } finally {
      set({ submitting: false });
    }
  },

  handleMoveProfileNode: async (nodeIndex, targetIndex) => {
    const { activeProfile, activeProfileId } = get();
    if (!activeProfileId || !activeProfile || nodeIndex < 0 || targetIndex < 0 || nodeIndex >= activeProfile.nodes.length || targetIndex >= activeProfile.nodes.length) {
      return;
    }

    const nodeName = activeProfile.nodes[nodeIndex]?.name || translate("shared.formats.nodeNumberLower", `node ${nodeIndex + 1}`, { index: nodeIndex + 1 });
    set({
      submitting: true,
      error: "",
      message: translate("dashboard.messages.reorderingNode", `Reordering ${nodeName}...`, { name: nodeName }),
    });

    try {
      const profile = await SakikoService.MoveProfileNode(activeProfileId, nodeIndex, targetIndex);
      set((state) => ({
        activeProfile: profile,
        profiles: upsertProfileSummary(state.profiles, profile),
        message: translate("dashboard.messages.movedNode", `Moved ${nodeName}.`, { name: nodeName }),
      }));
    } catch (err) {
      set({ error: normalizeError(err) });
    } finally {
      set({ submitting: false });
    }
  },
});

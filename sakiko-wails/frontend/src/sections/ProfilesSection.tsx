import { Box } from "@mui/material";
import { ImportProfilePanel } from "../components/profile/ImportProfilePanel";
import { ProfileDetailPanel } from "../components/profile/ProfileDetailPanel";
import { ProfileListPanel } from "../components/profile/ProfileListPanel";
import type { ImportForm } from "../types/dashboard";
import type { Profile, ProfileSummary } from "../types/sakiko";
import { getFilteredNodes } from "../utils/dashboard";

type ProfilesSectionProps = {
  activeProfile: Profile | null;
  activeProfileId: string;
  importForm: ImportForm;
  loading: boolean;
  profiles: ProfileSummary[];
  submitting: boolean;
  onDeleteProfile: () => Promise<void>;
  onImport: () => Promise<void>;
  onImportFormChange: (field: keyof ImportForm, value: string) => void;
  onNodeEnabledChange: (nodeIndex: number, enabled: boolean) => Promise<void>;
  onNodeMove: (nodeIndex: number, targetIndex: number) => Promise<void>;
  onRefreshProfile: () => Promise<void>;
  onReload: (preferredProfileId?: string) => Promise<void>;
  onSelectProfile: (profileId: string) => Promise<void>;
};

export function ProfilesSection({
  activeProfile,
  activeProfileId,
  importForm,
  loading,
  profiles,
  submitting,
  onDeleteProfile,
  onImport,
  onImportFormChange,
  onNodeEnabledChange,
  onNodeMove,
  onRefreshProfile,
  onReload,
  onSelectProfile,
}: ProfilesSectionProps) {
  const filteredNodes = getFilteredNodes(activeProfile, "");

  return (
    <Box className="sakiko-profiles-layout">
      <Box className="sakiko-profiles-layout__top">
        <ImportProfilePanel
          activeProfileId={activeProfileId}
          importForm={importForm}
          loading={loading}
          submitting={submitting}
          onImport={onImport}
          onImportFormChange={onImportFormChange}
          onReload={onReload}
        />

        <ProfileListPanel
          profiles={profiles}
          activeProfileId={activeProfileId}
          activeProfileName={activeProfile?.name || ""}
          submitting={submitting}
          canRefreshActiveProfile={Boolean(activeProfile?.source?.trim())}
          onRefreshActiveProfile={onRefreshProfile}
          onDeleteActiveProfile={onDeleteProfile}
          onSelect={onSelectProfile}
        />
      </Box>

      <Box className="sakiko-profiles-layout__detail">
        <ProfileDetailPanel
          activeProfile={activeProfile}
          filteredNodes={filteredNodes}
          submitting={submitting}
          onNodeEnabledChange={onNodeEnabledChange}
          onNodeMove={onNodeMove}
        />
      </Box>
    </Box>
  );
}

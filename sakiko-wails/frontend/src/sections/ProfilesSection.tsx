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
  nodeFilter: string;
  profiles: ProfileSummary[];
  submitting: boolean;
  onDeleteProfile: () => Promise<void>;
  onImport: () => Promise<void>;
  onImportFormChange: (field: keyof ImportForm, value: string) => void;
  onNodeEnabledChange: (nodeIndex: number, enabled: boolean) => Promise<void>;
  onNodeFilterChange: (value: string) => void;
  onRefreshProfile: () => Promise<void>;
  onReload: (preferredProfileId?: string) => Promise<void>;
  onSelectProfile: (profileId: string) => Promise<void>;
};

export function ProfilesSection({
  activeProfile,
  activeProfileId,
  importForm,
  loading,
  nodeFilter,
  profiles,
  submitting,
  onDeleteProfile,
  onImport,
  onImportFormChange,
  onNodeEnabledChange,
  onNodeFilterChange,
  onRefreshProfile,
  onReload,
  onSelectProfile,
}: ProfilesSectionProps) {
  const filteredNodes = getFilteredNodes(activeProfile, nodeFilter);

  return (
    <Box className="sakiko-section-grid">
      <Box className="sakiko-section-grid__sidebar">
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
          onSelect={onSelectProfile}
        />
      </Box>

      <Box className="sakiko-section-grid__content">
        <ProfileDetailPanel
          activeProfile={activeProfile}
          filteredNodes={filteredNodes}
          nodeFilter={nodeFilter}
          submitting={submitting}
          onDeleteProfile={onDeleteProfile}
          onNodeEnabledChange={onNodeEnabledChange}
          onNodeFilterChange={onNodeFilterChange}
          onRefreshProfile={onRefreshProfile}
        />
      </Box>
    </Box>
  );
}

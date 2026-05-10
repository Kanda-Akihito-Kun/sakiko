import { useShallow } from "zustand/react/shallow";
import { ImportProfilePanel } from "../../../../components/profile/ImportProfilePanel";
import { ProfileDetailPanel } from "../../../../components/profile/ProfileDetailPanel";
import { ProfileListPanel } from "../../../../components/profile/ProfileListPanel";
import { useDashboardStore } from "../../../../store/dashboardStore";
import { getFilteredNodes } from "../../../../utils/dashboard";
import { SectionLayout, SectionStack } from "../../components/SectionLayout";

export function ProfilesDashboardSection() {
  const view = useDashboardStore(useShallow((state) => ({
    activeProfile: state.activeProfile,
    activeProfileId: state.activeProfileId,
    importForm: state.importForm,
    loading: state.loading,
    profiles: state.profiles,
    submitting: state.submitting,
    handleDeleteProfile: state.handleDeleteProfile,
    handleImport: state.handleImport,
    updateImportForm: state.updateImportForm,
    handleSetProfileNodeEnabled: state.handleSetProfileNodeEnabled,
    handleMoveProfileNode: state.handleMoveProfileNode,
    handleRefreshProfile: state.handleRefreshProfile,
    refreshDashboard: state.refreshDashboard,
    handleProfileSelect: state.handleProfileSelect,
  })));
  const filteredNodes = getFilteredNodes(view.activeProfile, "");

  return (
    <SectionLayout>
      <SectionLayout
        columns={{
          xs: "minmax(0, 1fr)",
          xl: "repeat(2, minmax(0, 1fr))",
        }}
      >
        <ImportProfilePanel
          activeProfileId={view.activeProfileId}
          importForm={view.importForm}
          loading={view.loading}
          submitting={view.submitting}
          onImport={view.handleImport}
          onImportFormChange={view.updateImportForm}
          onReload={view.refreshDashboard}
        />

        <ProfileListPanel
          profiles={view.profiles}
          activeProfileId={view.activeProfileId}
          activeProfileName={view.activeProfile?.name || ""}
          submitting={view.submitting}
          canRefreshActiveProfile={Boolean(view.activeProfile?.source?.trim())}
          onRefreshActiveProfile={view.handleRefreshProfile}
          onDeleteActiveProfile={view.handleDeleteProfile}
          onSelect={view.handleProfileSelect}
        />
      </SectionLayout>

      <SectionStack>
        <ProfileDetailPanel
          activeProfile={view.activeProfile}
          filteredNodes={filteredNodes}
          submitting={view.submitting}
          onNodeEnabledChange={view.handleSetProfileNodeEnabled}
          onNodeMove={view.handleMoveProfileNode}
        />
      </SectionStack>
    </SectionLayout>
  );
}

import DnsRounded from "@mui/icons-material/DnsRounded";
import HubRounded from "@mui/icons-material/HubRounded";
import { Box, Tab, Tabs } from "@mui/material";
import { useEffect, useRef, useState } from "react";
import { useTranslation } from "react-i18next";
import { SectionLayout, SectionStack } from "../../components/SectionLayout";
import { RemoteDispatchPanel } from "./components/RemoteDispatchPanel";
import { RemoteKnightAccessPanel } from "./components/RemoteKnightAccessPanel";
import { RemoteKnightsListPanel } from "./components/RemoteKnightsListPanel";
import { RemoteMasterPanel } from "./components/RemoteMasterPanel";
import { RemoteResultArchivesPanel } from "./components/RemoteResultArchivesPanel";
import { RemoteTaskBoardPanel } from "./components/RemoteTaskBoardPanel";
import { useRemoteDashboardSectionModel } from "./useRemoteDashboardSectionModel";

type RemoteDashboardSectionProps = {
  onOpenProfiles: () => void;
  onOpenConfigs: () => void;
};

type RemoteTabKey = "master" | "knight";

const remoteTabStorageKey = "sakiko.dashboard.remote.tab";

export function RemoteDashboardSection({
  onOpenProfiles,
  onOpenConfigs,
}: RemoteDashboardSectionProps) {
  const { t } = useTranslation();
  const view = useRemoteDashboardSectionModel();
  const [tab, setTab] = useState<RemoteTabKey>(() => readStoredRemoteTab());
  const lastArchiveScopeRef = useRef<string>("");

  const masterArchivesVisible = tab === "master" && view.currentRole !== "knight";
  const knightArchivesVisible = tab === "knight" && view.currentRole === "knight";

  useEffect(() => {
    const scope = masterArchivesVisible ? "master" : knightArchivesVisible ? "knight" : "";
    if (!scope || view.remoteResultArchivesLoading || lastArchiveScopeRef.current === scope) {
      return;
    }
    lastArchiveScopeRef.current = scope;
    void view.refreshRemoteResultArchives(true);
  }, [
    knightArchivesVisible,
    masterArchivesVisible,
    view.refreshRemoteResultArchives,
    view.remoteResultArchivesLoading,
  ]);

  useEffect(() => {
    if (!masterArchivesVisible && !knightArchivesVisible) {
      lastArchiveScopeRef.current = "";
    }
  }, [knightArchivesVisible, masterArchivesVisible]);

  useEffect(() => {
    if (view.currentRole === "master" && tab !== "master") {
      writeStoredRemoteTab("master");
      setTab("master");
      return;
    }
    if (view.currentRole === "knight" && tab !== "knight") {
      writeStoredRemoteTab("knight");
      setTab("knight");
    }
  }, [tab, view.currentRole]);

  return (
    <Box sx={{ display: "grid", gap: 2.5 }}>
      <Tabs
        value={tab}
        onChange={(_event, nextTab: RemoteTabKey) => {
          writeStoredRemoteTab(nextTab);
          setTab(nextTab);
        }}
        variant="fullWidth"
        sx={{
          bgcolor: "background.paper",
          borderRadius: 3,
          p: 0.5,
          border: "1px solid",
          borderColor: "divider",
        }}
      >
        <Tab icon={<DnsRounded />} iconPosition="start" label={t("dashboard.remote.masterTab")} value="master" />
        <Tab icon={<HubRounded />} iconPosition="start" label={t("dashboard.remote.knightTab")} value="knight" />
      </Tabs>

      {tab === "master" ? (
        <SectionLayout
          columns={{
            xs: "minmax(0, 1fr)",
            xl: "1.15fr 0.95fr",
          }}
        >
          <SectionStack>
            <RemoteMasterPanel
              currentEligibility={view.currentEligibility}
              currentRole={view.currentRole}
              fields={{
                knightName: view.fields.knightName,
                listenHost: view.fields.listenHost,
                listenPort: view.fields.listenPort,
                ttlSeconds: view.fields.ttlSeconds,
              }}
              masterEnabled={view.masterEnabled}
              remoteKnights={view.remoteKnights}
              remoteLoading={view.remoteLoading}
              remotePairingCode={view.remotePairingCode}
              remoteSubmitting={view.remoteSubmitting}
              onCreatePairingCode={() => void view.handleCreateRemotePairingCode(view.fields.knightName.trim(), toPositiveInt(view.fields.ttlSeconds, 600))}
              onToggleMaster={(enabled) => {
                if (enabled) {
                  void view.handleEnableRemoteMaster(view.fields.listenHost.trim(), toPort(view.fields.listenPort, 10492));
                  return;
                }
                if (view.currentRole === "master") {
                  void view.handleDisableRemoteMode();
                }
              }}
              onProbeEligibility={() => void view.handleProbeRemoteMasterEligibility()}
              onRefresh={() => void view.refreshRemoteWorkspace(masterArchivesVisible)}
              onKnightNameChange={view.setKnightName}
              onListenHostChange={view.setListenHost}
              onListenPortChange={view.setListenPort}
              onTTLChange={view.setTTLSeconds}
            />

            <RemoteDispatchPanel
              activeProfileName={view.activeProfile?.name}
              currentRole={view.currentRole}
              remoteSubmitting={view.remoteSubmitting}
              remoteTasks={view.remoteTasks}
              selectedKnightCount={view.selectedRemoteKnightIDs.length}
              taskConfig={view.taskConfig}
              taskPreset={view.taskPreset}
              onDispatchTask={() => void view.handleSubmitRemoteTask()}
              onOpenProfiles={onOpenProfiles}
              onOpenConfigs={onOpenConfigs}
              onTaskPresetChange={view.setTaskPreset}
            />
          </SectionStack>

          <SectionStack>
            <RemoteKnightsListPanel
              currentRole={view.currentRole}
              remoteSubmitting={view.remoteSubmitting}
              onKickKnight={(knightId) => void view.handleKickRemoteKnight(knightId)}
              onToggleKnight={view.toggleRemoteKnightSelection}
              remoteKnights={view.remoteKnights}
              selectedKnightIDs={view.selectedRemoteKnightIDs}
            />
            <RemoteTaskBoardPanel
              remoteLoading={view.remoteLoading}
              remoteSubmitting={view.remoteSubmitting}
              remoteTasks={view.remoteTasks}
              onRefresh={() => void view.refreshRemoteWorkspace(false)}
            />
          </SectionStack>

          {masterArchivesVisible ? (
            <Box sx={{ gridColumn: "1 / -1" }}>
              <RemoteResultArchivesPanel
                archiveDetails={view.remoteResultArchiveDetails}
                archiveLoading={view.remoteResultArchiveLoading}
                archives={view.remoteResultArchives}
                downloadTargets={view.downloadTargets}
                loading={view.remoteResultArchivesLoading}
                role="master"
                visibleCount={view.remoteResultArchivesVisibleCount}
                onDeleteArchive={view.handleDeleteRemoteResultArchive}
                onEnsureArchive={view.ensureRemoteResultArchive}
                onLoadMore={view.loadMoreRemoteResultArchives}
                onRefresh={() => view.refreshRemoteResultArchives(true)}
              />
            </Box>
          ) : null}
        </SectionLayout>
      ) : (
        <SectionLayout
          columns={{
            xs: "minmax(0, 1fr)",
            lg: "minmax(0, 1fr)",
          }}
        >
          <SectionStack>
            <RemoteKnightAccessPanel
              currentRole={view.currentRole}
              fields={{
                masterHost: view.fields.masterHost,
                masterPort: view.fields.masterPort,
                oneTimeCode: view.fields.oneTimeCode,
                remoteJoinText: view.fields.remoteJoinText,
              }}
              knightBound={view.knightBound}
              remoteStatus={view.remoteStatus}
              remoteSubmitting={view.remoteSubmitting}
              onToggleKnight={(enabled) => {
                if (enabled) {
                  void view.handleEnableRemoteKnight(view.fields.masterHost.trim(), toPort(view.fields.masterPort, 10492), view.fields.oneTimeCode.trim());
                  return;
                }
                if (view.currentRole === "knight") {
                  void view.handleDisableRemoteMode();
                }
              }}
              onMasterHostChange={view.setMasterHost}
              onMasterPortChange={view.setMasterPort}
              onOneTimeCodeChange={view.setOneTimeCode}
              onRemoteJoinTextChange={view.setRemoteJoinText}
            />
          </SectionStack>

          {knightArchivesVisible ? (
            <Box sx={{ gridColumn: "1 / -1" }}>
              <RemoteResultArchivesPanel
                archiveDetails={view.remoteResultArchiveDetails}
                archiveLoading={view.remoteResultArchiveLoading}
                archives={view.remoteResultArchives}
                downloadTargets={view.downloadTargets}
                loading={view.remoteResultArchivesLoading}
                role="knight"
                visibleCount={view.remoteResultArchivesVisibleCount}
                onDeleteArchive={view.handleDeleteRemoteResultArchive}
                onEnsureArchive={view.ensureRemoteResultArchive}
                onLoadMore={view.loadMoreRemoteResultArchives}
                onRefresh={() => view.refreshRemoteResultArchives(true)}
              />
            </Box>
          ) : null}
        </SectionLayout>
      )}
    </Box>
  );
}

function toPort(value: string, fallback: number): number {
  const parsed = Number.parseInt(value, 10);
  if (!Number.isFinite(parsed) || parsed <= 0) {
    return fallback;
  }
  return parsed;
}

function toPositiveInt(value: string, fallback: number): number {
  const parsed = Number.parseInt(value, 10);
  if (!Number.isFinite(parsed) || parsed <= 0) {
    return fallback;
  }
  return parsed;
}

function readStoredRemoteTab(): RemoteTabKey {
  if (typeof window === "undefined") {
    return "master";
  }
  const raw = window.sessionStorage.getItem(remoteTabStorageKey);
  return raw === "knight" ? "knight" : "master";
}

function writeStoredRemoteTab(tab: RemoteTabKey) {
  if (typeof window === "undefined") {
    return;
  }
  window.sessionStorage.setItem(remoteTabStorageKey, tab);
}

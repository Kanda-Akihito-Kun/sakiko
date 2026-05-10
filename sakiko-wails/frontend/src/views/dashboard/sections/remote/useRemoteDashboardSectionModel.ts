import { useEffect, useState } from "react";
import { useShallow } from "zustand/react/shallow";
import { useDashboardStore } from "../../../../store/dashboardStore";
import { parseRemoteJoinText } from "../../../../utils/remoteInvite";

export function useRemoteDashboardSectionModel() {
  const store = useDashboardStore(useShallow((state) => ({
    activeProfile: state.activeProfile,
    downloadTargets: state.downloadTargets,
    ensureRemoteResultArchive: state.ensureRemoteResultArchive,
    handleDeleteRemoteResultArchive: state.handleDeleteRemoteResultArchive,
    remoteEligibility: state.remoteEligibility,
    remoteResultArchiveDetails: state.remoteResultArchiveDetails,
    remoteResultArchiveLoading: state.remoteResultArchiveLoading,
    remoteResultArchives: state.remoteResultArchives,
    remoteResultArchivesLoading: state.remoteResultArchivesLoading,
    remoteResultArchivesVisibleCount: state.remoteResultArchivesVisibleCount,
    remoteLoading: state.remoteLoading,
    remoteKnights: state.remoteKnights,
    remotePairingCode: state.remotePairingCode,
    remoteStatus: state.remoteStatus,
    remoteSubmitting: state.remoteSubmitting,
    remoteTasks: state.remoteTasks,
    selectedRemoteKnightIDs: state.selectedRemoteKnightIDs,
    taskConfig: state.taskConfig,
    taskPreset: state.taskPreset,
    setTaskPreset: state.setTaskPreset,
    toggleRemoteKnightSelection: state.toggleRemoteKnightSelection,
    handleCreateRemotePairingCode: state.handleCreateRemotePairingCode,
    handleKickRemoteKnight: state.handleKickRemoteKnight,
    handleDisableRemoteMode: state.handleDisableRemoteMode,
    handleEnableRemoteKnight: state.handleEnableRemoteKnight,
    handleEnableRemoteMaster: state.handleEnableRemoteMaster,
    handleProbeRemoteMasterEligibility: state.handleProbeRemoteMasterEligibility,
    handleSubmitRemoteTask: state.handleSubmitRemoteTask,
    loadMoreRemoteResultArchives: state.loadMoreRemoteResultArchives,
    refreshRemoteResultArchives: state.refreshRemoteResultArchives,
    refreshRemoteStatus: state.refreshRemoteStatus,
    refreshRemoteWorkspace: state.refreshRemoteWorkspace,
  })));

  const [listenHost, setListenHost] = useState("0.0.0.0");
  const [listenPort, setListenPort] = useState("10492");
  const [knightName, setKnightName] = useState("");
  const [ttlSeconds, setTTLSeconds] = useState("600");
  const [masterHost, setMasterHost] = useState("");
  const [masterPort, setMasterPort] = useState("10492");
  const [oneTimeCode, setOneTimeCode] = useState("");
  const [remoteJoinText, setRemoteJoinTextState] = useState("");

  const currentRole = store.remoteStatus?.role || "standalone";
  const currentEligibility = store.remoteEligibility || store.remoteStatus?.master?.eligibility || null;
  const masterEnabled = Boolean(store.remoteStatus?.master?.enabled);
  const knightBound = Boolean(store.remoteStatus?.knight?.bound);

  useEffect(() => {
    void store.refreshRemoteStatus();
  }, [store.refreshRemoteStatus]);

  useEffect(() => {
    if (currentRole === "standalone") {
      return;
    }

    const timer = window.setInterval(() => {
      void store.refreshRemoteWorkspace(false);
    }, 3000);

    return () => {
      window.clearInterval(timer);
    };
  }, [currentRole, store.refreshRemoteWorkspace]);

  const setRemoteJoinText = (value: string) => {
    setRemoteJoinTextState(value);
    const parsed = parseRemoteJoinText(value);
    if (!parsed) {
      return;
    }
    setMasterHost(parsed.host);
    setMasterPort(parsed.port);
    setOneTimeCode(parsed.code);
  };

  return {
    ...store,
    currentEligibility,
    currentRole,
    knightBound,
    masterEnabled,
    fields: {
      knightName,
      listenHost,
      listenPort,
      masterHost,
      masterPort,
      oneTimeCode,
      remoteJoinText,
      ttlSeconds,
    },
    setKnightName,
    setListenHost,
    setListenPort,
    setMasterHost,
    setMasterPort,
    setOneTimeCode,
    setRemoteJoinText,
    setTTLSeconds,
  };
}

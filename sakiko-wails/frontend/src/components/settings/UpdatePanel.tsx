import OpenInNewRounded from "@mui/icons-material/OpenInNewRounded";
import SystemUpdateAltRounded from "@mui/icons-material/SystemUpdateAltRounded";
import { Button, Stack } from "@mui/material";
import { useEffect, useState } from "react";
import { useTranslation } from "react-i18next";
import { APP_VERSION } from "../../constants/appMeta";
import { SakikoService } from "../../services/sakikoService";
import type { ReleaseCheckResult } from "../../types/release";
import { formatDateTime, normalizeError } from "../../utils/dashboard";
import { OverviewRow } from "../shared/OverviewRow";
import { SectionCard } from "../shared/SectionCard";

type UpdateState = {
  loading: boolean;
  release: ReleaseCheckResult | null;
  error: string;
};

const initialState: UpdateState = {
  loading: true,
  release: null,
  error: "",
};

export function UpdatePanel() {
  const { t } = useTranslation();
  const [state, setState] = useState<UpdateState>(initialState);

  useEffect(() => {
    let active = true;

    void (async () => {
      setState((current) => ({ ...current, loading: true, error: "" }));

      try {
        const release = await SakikoService.CheckForUpdates();
        if (!active) {
          return;
        }
        setState({
          loading: false,
          release,
          error: "",
        });
      } catch (err) {
        if (!active) {
          return;
        }
        setState({
          loading: false,
          release: null,
          error: normalizeError(err),
        });
      }
    })();

    return () => {
      active = false;
    };
  }, []);

  const handleCheckForUpdates = async () => {
    setState((current) => ({ ...current, loading: true, error: "" }));

    try {
      const release = await SakikoService.CheckForUpdates();
      setState({
        loading: false,
        release,
        error: "",
      });
    } catch (err) {
      setState((current) => ({
        loading: false,
        release: current.release,
        error: normalizeError(err),
      }));
    }
  };

  const handleOpenReleasePage = async () => {
    try {
      await SakikoService.OpenReleasePage(state.release?.releaseURL);
    } catch (err) {
      setState((current) => ({
        ...current,
        error: normalizeError(err),
      }));
    }
  };

  const currentVersion = state.release?.currentVersion || APP_VERSION;
  const latestVersion = state.loading
    ? t("shared.states.loading")
    : state.release?.latestVersion || t("settings.updates.latestVersionUnavailable");
  const statusLabel = state.loading
    ? t("settings.updates.statusChecking")
    : state.error
      ? t("settings.updates.statusCheckFailed")
      : state.release?.hasUpdate
        ? t("settings.updates.statusUpdateAvailable", { version: state.release.latestVersion || t("settings.updates.latestVersionUnavailable") })
        : t("settings.updates.statusLatest");
  const checkedAt = state.release?.checkedAt
    ? formatDateTime(state.release.checkedAt)
    : t("settings.updates.notChecked");

  return (
    <SectionCard
      title={t("settings.updates.title")}
      icon={<SystemUpdateAltRounded color="primary" />}
    >
      <Stack spacing={1.25}>
        <Stack direction={{ xs: "column", sm: "row" }} spacing={1}>
          <Button
            variant="outlined"
            onClick={() => void handleCheckForUpdates()}
            disabled={state.loading}
          >
            {t("settings.updates.check")}
          </Button>
          <Button
            variant="outlined"
            startIcon={<OpenInNewRounded />}
            onClick={() => void handleOpenReleasePage()}
          >
            {t("settings.updates.openReleasePage")}
          </Button>
        </Stack>
        <OverviewRow
          label={t("settings.updates.currentVersion")}
          value={currentVersion}
          mono
        />
        <OverviewRow
          label={t("settings.updates.latestVersion")}
          value={latestVersion}
          mono
        />
        <OverviewRow
          label={t("settings.updates.status")}
          value={statusLabel}
        />
        <OverviewRow
          label={t("settings.updates.checkedAt")}
          value={checkedAt}
          mono
        />
        {state.error ? (
          <OverviewRow
            label={t("settings.updates.error")}
            value={state.error}
            multiline
          />
        ) : null}
      </Stack>
    </SectionCard>
  );
}

import RefreshRounded from "@mui/icons-material/RefreshRounded";
import SearchRounded from "@mui/icons-material/SearchRounded";
import TuneRounded from "@mui/icons-material/TuneRounded";
import {
  Box,
  Button,
  Chip,
  InputAdornment,
  List,
  ListItemButton,
  Stack,
  TextField,
  ToggleButton,
  ToggleButtonGroup,
  Typography,
} from "@mui/material";
import { useEffect, useRef } from "react";
import { useTranslation } from "react-i18next";
import { DownloadTargetSource } from "../../types/sakiko";
import type { DownloadTarget, TaskConfig } from "../../types/sakiko";
import { SectionCard } from "../shared/SectionCard";

type TaskConfigPatch = Partial<Pick<TaskConfig, "pingAddress" | "taskTimeoutMillis" | "downloadURL" | "downloadDuration" | "downloadThreading" | "backendIdentity">>;

type TaskDefaultsPanelProps = {
  downloadTargets: DownloadTarget[];
  downloadTargetSearch: string;
  downloadTargetsLoading: boolean;
  taskConfig: TaskConfig;
  onPatchTaskConfig: (patch: TaskConfigPatch) => void;
  onDownloadTargetSearchChange: (value: string) => void;
  onRefreshDownloadTargets: (search?: string) => void;
};

export function TaskDefaultsPanel({
  downloadTargets,
  downloadTargetSearch,
  downloadTargetsLoading,
  taskConfig,
  onPatchTaskConfig,
  onDownloadTargetSearchChange,
  onRefreshDownloadTargets,
}: TaskDefaultsPanelProps) {
  const { t } = useTranslation();
  const searchInitializedRef = useRef(false);
  const defaultTarget = downloadTargets.find((target) => target.source === DownloadTargetSource.DownloadTargetSourceCloudflare) || null;
  const speedtestTargets = downloadTargets.filter((target) => target.source === DownloadTargetSource.DownloadTargetSourceSpeedtest);
  const normalizedTargetSearch = downloadTargetSearch.trim();

  useEffect(() => {
    if (!searchInitializedRef.current) {
      searchInitializedRef.current = true;
      return;
    }

    const timer = window.setTimeout(() => {
      void onRefreshDownloadTargets(downloadTargetSearch);
    }, 400);

    return () => {
      window.clearTimeout(timer);
    };
  }, [downloadTargetSearch, onRefreshDownloadTargets]);

  return (
    <SectionCard
      title={t("settings.taskDefaults.title")}
      subtitle={t("settings.taskDefaults.subtitle")}
      icon={<TuneRounded color="primary" />}
    >
      <Stack spacing={2}>
        <Typography variant="body2" color="text.secondary">
          {t("settings.taskDefaults.description")}
        </Typography>

        <Box
          sx={{
            display: "grid",
            gridTemplateColumns: { xs: "1fr", md: "repeat(2, minmax(0, 1fr))" },
            gap: 1.5,
          }}
        >
          <TextField
            fullWidth
            label={t("settings.taskDefaults.pingUrl")}
            value={taskConfig.pingAddress}
            onChange={(event) => onPatchTaskConfig({ pingAddress: event.target.value })}
            helperText={t("settings.taskDefaults.pingUrlHelper")}
          />
          <TextField
            fullWidth
            label={t("settings.taskDefaults.timeout")}
            type="number"
            value={taskConfig.taskTimeoutMillis}
            onChange={(event) => onPatchTaskConfig({ taskTimeoutMillis: Number(event.target.value) })}
          />
          <TextField
            fullWidth
            label={t("settings.taskDefaults.downloadUrl")}
            value={taskConfig.downloadURL}
            onChange={(event) => onPatchTaskConfig({ downloadURL: event.target.value })}
            helperText={t("settings.taskDefaults.downloadUrlHelper")}
          />
          <TextField
            fullWidth
            label={t("settings.taskDefaults.duration")}
            type="number"
            value={taskConfig.downloadDuration}
            onChange={(event) => onPatchTaskConfig({ downloadDuration: Number(event.target.value) })}
            inputProps={{ min: 5, max: 20 }}
            helperText={t("settings.taskDefaults.durationHelper")}
          />
          <TextField
            fullWidth
            label={t("settings.taskDefaults.backendIdentity")}
            value={taskConfig.backendIdentity || ""}
            onChange={(event) => onPatchTaskConfig({ backendIdentity: event.target.value })}
            helperText={t("settings.taskDefaults.backendIdentityHelper", {
              count: Array.from(taskConfig.backendIdentity || "").length,
            })}
            inputProps={{ maxLength: 50 }}
          />
        </Box>

        <Stack spacing={1.25}>
          <Stack direction="row" justifyContent="space-between" alignItems="center" spacing={1}>
            <Box sx={{ minWidth: 0 }}>
              <Typography variant="subtitle2" color="text.secondary">
                {t("settings.taskDefaults.downloadTarget.title")}
              </Typography>
              <Typography variant="caption" color="text.secondary">
                {t("settings.taskDefaults.downloadTarget.description")}
              </Typography>
            </Box>
            <Button
              size="small"
              variant="outlined"
              startIcon={<RefreshRounded />}
              disabled={downloadTargetsLoading}
              onClick={() => void onRefreshDownloadTargets(downloadTargetSearch)}
            >
              {t("shared.actions.refresh")}
            </Button>
          </Stack>

          {defaultTarget ? (
            <Button
              variant={taskConfig.downloadURL === defaultTarget.downloadURL ? "contained" : "outlined"}
              onClick={() => onPatchTaskConfig({ downloadURL: defaultTarget.downloadURL })}
              sx={{ justifyContent: "flex-start" }}
            >
              {t("settings.taskDefaults.downloadTarget.useDefault", { name: defaultTarget.name })}
            </Button>
          ) : null}

          <TextField
            fullWidth
            size="small"
            label={t("settings.taskDefaults.downloadTarget.searchLabel")}
            value={downloadTargetSearch}
            onChange={(event) => onDownloadTargetSearchChange(event.target.value)}
            placeholder={t("settings.taskDefaults.downloadTarget.searchPlaceholder")}
            helperText={t("settings.taskDefaults.downloadTarget.searchHelper")}
            InputProps={{
              startAdornment: (
                <InputAdornment position="start">
                  <SearchRounded sx={{ color: "text.secondary" }} />
                </InputAdornment>
              ),
            }}
          />

          <Stack direction="row" spacing={1} useFlexGap sx={{ flexWrap: "wrap" }}>
            <Chip
              size="small"
              variant="outlined"
              label={
                normalizedTargetSearch
                  ? t("settings.taskDefaults.downloadTarget.results", { count: speedtestTargets.length, search: normalizedTargetSearch })
                  : t("settings.taskDefaults.downloadTarget.targets", { count: speedtestTargets.length })
              }
            />
          </Stack>

          <List
            disablePadding
            sx={{
              maxHeight: 320,
              overflowY: "auto",
              borderRadius: 2,
              border: (theme) => `1px solid ${theme.palette.divider}`,
            }}
          >
            {speedtestTargets.length === 0 ? (
              <Box px={1.5} py={2}>
                <Typography variant="body2" color="text.secondary">
                  {downloadTargetsLoading
                    ? t("settings.taskDefaults.downloadTarget.loading")
                    : normalizedTargetSearch
                      ? t("settings.taskDefaults.downloadTarget.notFound", { search: normalizedTargetSearch })
                      : t("settings.taskDefaults.downloadTarget.notLoaded")}
                </Typography>
              </Box>
            ) : (
              speedtestTargets.map((target) => (
                <ListItemButton
                  key={target.id}
                  selected={taskConfig.downloadURL === target.downloadURL}
                  onClick={() => onPatchTaskConfig({ downloadURL: target.downloadURL })}
                  sx={{ alignItems: "flex-start", flexDirection: "column", gap: 0.5 }}
                >
                  <Stack width="100%" direction="row" justifyContent="space-between" spacing={1}>
                    <Typography variant="body2" fontWeight={600} noWrap sx={{ minWidth: 0 }}>
                      {target.city || target.name}
                      {target.country ? `, ${target.country}` : ""}
                    </Typography>
                    {taskConfig.downloadURL === target.downloadURL ? (
                      <Chip size="small" color="primary" label={t("settings.taskDefaults.downloadTarget.selected")} />
                    ) : null}
                  </Stack>
                  <Typography variant="caption" color="text.secondary" noWrap sx={{ width: "100%" }}>
                    {target.sponsor || t("settings.taskDefaults.downloadTarget.unknownSponsor")}
                  </Typography>
                  <Typography variant="caption" className="sakiko-mono" color="text.secondary" noWrap sx={{ width: "100%" }}>
                    {target.host || target.endpoint}
                  </Typography>
                </ListItemButton>
              ))
            )}
          </List>
        </Stack>

        <Stack spacing={1}>
          <Typography variant="subtitle2" color="text.secondary">
            {t("settings.taskDefaults.speedMode.title")}
          </Typography>
          <ToggleButtonGroup
            exclusive
            fullWidth
            value={taskConfig.downloadThreading > 1 ? "multi" : "single"}
            onChange={(_event, value: "single" | "multi" | null) => {
              if (value === "single") {
                onPatchTaskConfig({ downloadThreading: 1 });
              }
              if (value === "multi") {
                onPatchTaskConfig({ downloadThreading: 8 });
              }
            }}
          >
            <ToggleButton value="single">{t("settings.taskDefaults.speedMode.single")}</ToggleButton>
            <ToggleButton value="multi">{t("settings.taskDefaults.speedMode.multi")}</ToggleButton>
          </ToggleButtonGroup>
          <Typography variant="caption" color="text.secondary">
            {taskConfig.downloadThreading > 1
              ? t("settings.taskDefaults.speedMode.multiDescription", { count: taskConfig.downloadThreading })
              : t("settings.taskDefaults.speedMode.singleDescription")}
          </Typography>
        </Stack>
      </Stack>
    </SectionCard>
  );
}

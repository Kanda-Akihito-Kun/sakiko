import CloudUploadRounded from "@mui/icons-material/CloudUploadRounded";
import RefreshRounded from "@mui/icons-material/RefreshRounded";
import { Button, Stack, TextField } from "@mui/material";
import { useTranslation } from "react-i18next";
import type { ImportForm } from "../../types/dashboard";
import { SectionCard } from "../shared/SectionCard";

type ImportProfilePanelProps = {
  activeProfileId: string;
  importForm: ImportForm;
  loading: boolean;
  submitting: boolean;
  onImport: () => void;
  onImportFormChange: (field: keyof ImportForm, value: string) => void;
  onReload: (preferredProfileId?: string) => Promise<void>;
};

export function ImportProfilePanel({
  activeProfileId,
  importForm,
  loading,
  submitting,
  onImport,
  onImportFormChange,
  onReload,
}: ImportProfilePanelProps) {
  const { t } = useTranslation();

  return (
    <SectionCard
      title={t("dashboard.profiles.import.title")}
      subtitle={t("dashboard.profiles.import.subtitle")}
      icon={<CloudUploadRounded color="primary" />}
    >
      <Stack spacing={2}>
        <TextField
          label={t("dashboard.profiles.import.name")}
          fullWidth
          value={importForm.name}
          onChange={(event) => onImportFormChange("name", event.target.value)}
          placeholder={t("dashboard.profiles.import.namePlaceholder")}
        />

        <TextField
          label={t("dashboard.profiles.import.source")}
          fullWidth
          value={importForm.source}
          onChange={(event) => onImportFormChange("source", event.target.value)}
          placeholder={t("dashboard.profiles.import.sourcePlaceholder")}
        />

        <Stack direction={{ xs: "column", sm: "row" }} spacing={1.5}>
          <Button
            variant="contained"
            startIcon={<CloudUploadRounded />}
            disabled={submitting}
            onClick={onImport}
          >
            {t("shared.actions.import")}
          </Button>
          <Button
            variant="outlined"
            startIcon={<RefreshRounded />}
            disabled={loading}
            onClick={() => void onReload(activeProfileId)}
          >
            {t("shared.actions.reload")}
          </Button>
        </Stack>
      </Stack>
    </SectionCard>
  );
}

import CloudUploadRounded from "@mui/icons-material/CloudUploadRounded";
import RefreshRounded from "@mui/icons-material/RefreshRounded";
import { Button, Stack, TextField } from "@mui/material";
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
  return (
    <SectionCard
      title="Import Profile"
      subtitle="Subscription URL import"
      icon={<CloudUploadRounded color="primary" />}
    >
      <Stack spacing={2}>
        <TextField
          label="Custome Profile Name"
          fullWidth
          value={importForm.name}
          onChange={(event) => onImportFormChange("name", event.target.value)}
          placeholder="Use profile's name in default"
        />

        <TextField
          label="Source URL"
          fullWidth
          value={importForm.source}
          onChange={(event) => onImportFormChange("source", event.target.value)}
          placeholder="https://example.com/sub.yaml"
        />

        <Stack direction={{ xs: "column", sm: "row" }} spacing={1.5}>
          <Button
            variant="contained"
            startIcon={<CloudUploadRounded />}
            disabled={submitting}
            onClick={onImport}
          >
            Import
          </Button>
          <Button
            variant="outlined"
            startIcon={<RefreshRounded />}
            disabled={loading}
            onClick={() => void onReload(activeProfileId)}
          >
            Reload
          </Button>
        </Stack>
      </Stack>
    </SectionCard>
  );
}

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
      subtitle="Source URL or inline Clash YAML"
      icon={<CloudUploadRounded color="primary" />}
    >
      <Stack spacing={2}>
        <TextField
          label="Name"
          fullWidth
          value={importForm.name}
          onChange={(event) => onImportFormChange("name", event.target.value)}
          placeholder="Tokyo Nodes / My Clash Sub"
        />

        <TextField
          label="Source URL"
          fullWidth
          value={importForm.source}
          onChange={(event) => onImportFormChange("source", event.target.value)}
          placeholder="https://example.com/sub.yaml"
        />

        <TextField
          label="Inline Content"
          fullWidth
          multiline
          minRows={8}
          value={importForm.content}
          onChange={(event) => onImportFormChange("content", event.target.value)}
          placeholder={"proxies:\n  - name: hk-1\n    type: vmess\n    server: 1.1.1.1\n    port: 443"}
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

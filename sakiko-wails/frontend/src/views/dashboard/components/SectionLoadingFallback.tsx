import { CircularProgress, Stack, Typography } from "@mui/material";
import { useTranslation } from "react-i18next";

type SectionLoadingFallbackProps = {
  label: string;
};

export function SectionLoadingFallback({ label }: SectionLoadingFallbackProps) {
  const { t } = useTranslation();

  return (
    <Stack spacing={1.5} alignItems="center" justifyContent="center" sx={{ minHeight: 240 }}>
      <CircularProgress size={28} />
      <Typography variant="body2" color="text.secondary">
        {t("shared.states.loadingSection", { label })}
      </Typography>
    </Stack>
  );
}

import QueryStatsRounded from "@mui/icons-material/QueryStatsRounded";
import { Box, Card, Chip, Divider, Stack, Typography } from "@mui/material";
import { alpha } from "@mui/material/styles";
import { useTranslation } from "react-i18next";
import type { EntryResult } from "../../types/sakiko";
import {
  formatDuration,
  formatMatrixLabel,
  formatMatrixPayload,
  formatProxyTypeLabel,
  shouldUseEmojiFont,
} from "../../utils/dashboard";

type ResultEntryCardProps = {
  result: EntryResult;
};

export function ResultEntryCard({ result }: ResultEntryCardProps) {
  const { t } = useTranslation();

  return (
    <Card
      variant="outlined"
      sx={{
        p: 2.25,
        height: "100%",
        minWidth: 0,
      }}
    >
      <Stack spacing={1.5} sx={{ minWidth: 0, height: "100%" }}>
        <Stack
          direction="row"
          justifyContent="space-between"
          spacing={1}
          sx={{ minWidth: 0, alignItems: "flex-start" }}
        >
          <Box sx={{ minWidth: 0, flex: "1 1 auto" }}>
            <Typography
              variant="h6"
              noWrap
              title={result.proxyInfo.name || t("shared.states.unnamedNode")}
              className={shouldUseEmojiFont("nodeName", result.proxyInfo.name) ? "sakiko-emoji" : undefined}
            >
              {result.proxyInfo.name || t("shared.states.unnamedNode")}
            </Typography>
            <Typography
              variant="body2"
              color="text.secondary"
              noWrap
              title={result.proxyInfo.address || t("shared.states.noAddress")}
            >
              {result.proxyInfo.address || t("shared.states.noAddress")}
            </Typography>
          </Box>
          <Chip
            label={formatProxyTypeLabel(result.proxyInfo.type)}
            size="small"
            color="primary"
            variant="outlined"
            sx={{ flexShrink: 0, maxWidth: "100%" }}
          />
        </Stack>

        <Stack direction="row" spacing={1} sx={{ flexWrap: "nowrap", overflowX: "auto", pb: 0.25 }}>
          <Chip
            icon={<QueryStatsRounded />}
            label={formatDuration(result.invokeDuration)}
            size="small"
          />
          <Chip
            label={t("shared.formats.matricesCount", { count: result.matrices.length })}
            size="small"
            variant="outlined"
          />
        </Stack>

        {result.error ? (
          <Typography variant="body2" color="error.main" sx={{ overflowWrap: "anywhere", wordBreak: "break-word" }}>
            {result.error}
          </Typography>
        ) : null}

        <Divider sx={{ borderColor: "divider" }} />
        <Stack spacing={1}>
          {result.matrices.map((matrix, matrixIndex) => {
            const payloadLabel = formatMatrixPayload(matrix.payload, matrix.type);

            return (
              <Box
                key={`${matrix.type}-${matrixIndex}`}
                sx={(theme) => ({
                  display: "flex",
                  justifyContent: "space-between",
                  alignItems: "center",
                  gap: 1.5,
                  p: 1.25,
                  borderRadius: 2,
                  bgcolor: alpha(theme.palette.primary.main, 0.08),
                  minWidth: 0,
                })}
              >
                <Typography variant="body2" color="text.secondary" sx={{ minWidth: 0, flex: "0 1 auto" }}>
                  {formatMatrixLabel(matrix.type)}
                </Typography>
                <Typography
                  variant="body2"
                  className="sakiko-mono"
                  noWrap
                  title={payloadLabel}
                  sx={{ minWidth: 0, flex: "1 1 auto", textAlign: "right" }}
                >
                  {payloadLabel}
                </Typography>
              </Box>
            );
          })}
        </Stack>
      </Stack>
    </Card>
  );
}

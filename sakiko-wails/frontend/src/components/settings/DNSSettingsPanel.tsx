import DnsRounded from "@mui/icons-material/DnsRounded";
import RefreshRounded from "@mui/icons-material/RefreshRounded";
import SaveRounded from "@mui/icons-material/SaveRounded";
import { Alert, Box, Button, Stack, TextField, Typography } from "@mui/material";
import { useEffect, useMemo, useState } from "react";
import { useTranslation } from "react-i18next";
import { SakikoService } from "../../services/sakikoService";
import type { DNSConfig } from "../../types/appSettings";
import { SectionCard } from "../shared/SectionCard";

function configToText(values: string[]) {
  return values.join("\n");
}

function textToConfig(value: string) {
  return value
    .split(/\r?\n|,/)
    .map((entry) => entry.trim())
    .filter((entry) => entry.length > 0);
}

export function DNSSettingsPanel() {
  const { t } = useTranslation();
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState("");
  const [message, setMessage] = useState("");
  const [bootstrapText, setBootstrapText] = useState("");
  const [resolverText, setResolverText] = useState("");

  useEffect(() => {
    let cancelled = false;

    const load = async () => {
      setLoading(true);
      setError("");
      try {
        const settings = await SakikoService.GetAppSettings();
        if (cancelled) {
          return;
        }
        setBootstrapText(configToText(settings.dns.bootstrapServers));
        setResolverText(configToText(settings.dns.resolverServers));
      } catch (loadError) {
        if (!cancelled) {
          setError(loadError instanceof Error ? loadError.message : t("settings.dns.loadFailed"));
        }
      } finally {
        if (!cancelled) {
          setLoading(false);
        }
      }
    };

    void load();
    return () => {
      cancelled = true;
    };
  }, [t]);

  const parsedConfig = useMemo<DNSConfig>(() => ({
    bootstrapServers: textToConfig(bootstrapText),
    resolverServers: textToConfig(resolverText),
  }), [bootstrapText, resolverText]);

  const handleSave = async (nextConfig?: DNSConfig) => {
    setSaving(true);
    setError("");
    setMessage("");
    try {
      const settings = await SakikoService.UpdateAppSettings({
        dns: nextConfig || parsedConfig,
      });
      setBootstrapText(configToText(settings.dns.bootstrapServers));
      setResolverText(configToText(settings.dns.resolverServers));
      setMessage(t("settings.dns.saved"));
    } catch (saveError) {
      setError(saveError instanceof Error ? saveError.message : t("settings.dns.saveFailed"));
    } finally {
      setSaving(false);
    }
  };

  return (
    <SectionCard
      title={t("settings.dns.title")}
      icon={<DnsRounded color="primary" />}
      action={(
        <Button
          size="small"
          variant="outlined"
          startIcon={<RefreshRounded />}
          disabled={loading || saving}
          onClick={() => void handleSave({ bootstrapServers: [], resolverServers: [] })}
        >
          {t("settings.dns.restoreDefaults")}
        </Button>
      )}
    >
      <Stack spacing={2}>
        <Typography variant="body2" color="text.secondary">
          {t("settings.dns.description")}
        </Typography>

        {error ? <Alert severity="error">{error}</Alert> : null}
        {message ? <Alert severity="success">{message}</Alert> : null}

        <Box
          sx={{
            display: "grid",
            gridTemplateColumns: { xs: "1fr", md: "repeat(2, minmax(0, 1fr))" },
            gap: 1.5,
          }}
        >
          <TextField
            fullWidth
            multiline
            minRows={5}
            label={t("settings.dns.bootstrapLabel")}
            value={bootstrapText}
            disabled={loading || saving}
            onChange={(event) => setBootstrapText(event.target.value)}
            helperText={t("settings.dns.bootstrapHelper", { count: parsedConfig.bootstrapServers.length })}
          />
          <TextField
            fullWidth
            multiline
            minRows={5}
            label={t("settings.dns.resolverLabel")}
            value={resolverText}
            disabled={loading || saving}
            onChange={(event) => setResolverText(event.target.value)}
            helperText={t("settings.dns.resolverHelper", { count: parsedConfig.resolverServers.length })}
          />
        </Box>

        <Stack direction="row" spacing={1.25}>
          <Button
            variant="contained"
            startIcon={<SaveRounded />}
            disabled={loading || saving}
            onClick={() => void handleSave()}
          >
            {t("shared.actions.save")}
          </Button>
        </Stack>
      </Stack>
    </SectionCard>
  );
}

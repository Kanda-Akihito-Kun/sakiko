import InfoOutlined from "@mui/icons-material/InfoOutlined";
import { alpha } from "@mui/material/styles";
import { Box, Stack, Typography } from "@mui/material";

type EmptyStateProps = {
  title: string;
  description: string;
};

export function EmptyState({ title, description }: EmptyStateProps) {
  return (
    <Stack
      spacing={1.25}
      alignItems="flex-start"
      sx={{
        px: 2.25,
        py: 2.5,
        borderRadius: 2.5,
        border: "1px solid",
        borderColor: "divider",
        backgroundColor: "background.default",
      }}
    >
      <Box
        sx={(theme) => ({
          width: 40,
          height: 40,
          borderRadius: 2,
          display: "grid",
          placeItems: "center",
          bgcolor: alpha(theme.palette.primary.main, 0.14),
          color: "primary.main",
        })}
      >
        <InfoOutlined fontSize="small" />
      </Box>
      <Typography variant="h6">{title}</Typography>
      <Typography variant="body2" color="text.secondary">
        {description}
      </Typography>
    </Stack>
  );
}

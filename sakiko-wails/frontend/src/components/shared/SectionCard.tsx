import type { ReactNode } from "react";
import type { SxProps, Theme } from "@mui/material/styles";
import { Card, CardContent, CardHeader, Typography } from "@mui/material";

type SectionCardProps = {
  title: string;
  action?: ReactNode;
  icon?: ReactNode;
  children: ReactNode;
  sx?: SxProps<Theme>;
};

export function SectionCard({ title, action, icon, children, sx }: SectionCardProps) {
  const cardSx: SxProps<Theme> = [
    {
      minWidth: 0,
      display: "flex",
      flexDirection: "column",
      overflow: "hidden",
    },
    ...(Array.isArray(sx) ? sx : sx ? [sx] : []),
  ];

  return (
    <Card
      variant="outlined"
      sx={cardSx}
    >
      <CardHeader
        avatar={icon}
        action={action}
        title={<Typography variant="subtitle1" noWrap title={title}>{title}</Typography>}
        sx={{
          display: "grid",
          gridTemplateColumns: { xs: "auto minmax(0, 1fr)", sm: "auto minmax(0, 1fr) auto" },
          gridTemplateAreas: { xs: "\"avatar content\" \"action action\"", sm: "\"avatar content action\"" },
          alignItems: "center",
          px: 2.25,
          py: 1.5,
          pb: 1.25,
          borderBottom: "1px solid",
          borderColor: "divider",
          minWidth: 0,
          columnGap: 1,
          rowGap: 1,
          "& .MuiCardHeader-content": {
            gridArea: "content",
            minWidth: 0,
            overflow: "hidden",
          },
          "& .MuiCardHeader-avatar": {
            gridArea: "avatar",
            color: "primary.main",
            flexShrink: 0,
            mr: 0,
          },
          "& .MuiCardHeader-action": {
            gridArea: "action",
            justifySelf: { xs: "stretch", sm: "end" },
            mr: 0,
            mt: 0,
            minWidth: 0,
            "& > *": {
              width: { xs: "100%", sm: "auto" },
            },
          },
        }}
      />
      <CardContent sx={{ px: 2.25, py: 2, minWidth: 0, flex: 1 }}>{children}</CardContent>
    </Card>
  );
}

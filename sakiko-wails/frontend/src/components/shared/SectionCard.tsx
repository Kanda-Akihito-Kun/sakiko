import type { ReactNode } from "react";
import { Card, CardContent, CardHeader, Typography } from "@mui/material";

type SectionCardProps = {
  title: string;
  subtitle: string;
  action?: ReactNode;
  icon?: ReactNode;
  children: ReactNode;
  subtitleWrap?: boolean;
};

export function SectionCard({ title, subtitle, action, icon, children, subtitleWrap = false }: SectionCardProps) {
  return (
    <Card
      variant="outlined"
      sx={{
        minWidth: 0,
        display: "flex",
        flexDirection: "column",
        overflow: "hidden",
      }}
    >
      <CardHeader
        avatar={icon}
        action={action}
        title={<Typography variant="subtitle1" noWrap title={title}>{title}</Typography>}
        subheader={
          <Typography
            variant="body2"
            color="text.secondary"
            noWrap={!subtitleWrap}
            title={!subtitleWrap ? subtitle : undefined}
            sx={
              subtitleWrap
                ? {
                    whiteSpace: "pre-wrap",
                    overflowWrap: "anywhere",
                    wordBreak: "break-word",
                  }
                : undefined
            }
          >
            {subtitle}
          </Typography>
        }
        sx={{
          px: 2.25,
          py: 1.5,
          pb: 1.25,
          borderBottom: "1px solid",
          borderColor: "divider",
          minWidth: 0,
          gap: 1,
          flexWrap: "wrap",
          "& .MuiCardHeader-content": {
            minWidth: 0,
            overflow: "hidden",
            flex: "1 1 260px",
          },
          "& .MuiCardHeader-avatar": {
            color: "primary.main",
            flexShrink: 0,
          },
          "& .MuiCardHeader-action": {
            flex: { xs: "1 1 100%", sm: "0 0 auto" },
            ml: { xs: 0, sm: 1.5 },
            mr: 0,
            mt: 0,
            alignSelf: "center",
            "& > *": {
              width: { xs: "100%", sm: "auto" },
            },
          },
        }}
      />
      <CardContent sx={{ px: 2.25, py: 2, minWidth: 0 }}>{children}</CardContent>
    </Card>
  );
}

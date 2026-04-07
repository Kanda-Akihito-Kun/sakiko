import type { ReactNode } from "react";
import { Card, CardContent, CardHeader, Typography } from "@mui/material";

type SectionCardProps = {
  title: string;
  subtitle: string;
  action?: ReactNode;
  icon?: ReactNode;
  children: ReactNode;
};

export function SectionCard({ title, subtitle, action, icon, children }: SectionCardProps) {
  return (
    <Card variant="outlined">
      <CardHeader
        avatar={icon}
        action={action}
        title={<Typography variant="subtitle1" noWrap>{title}</Typography>}
        subheader={<Typography variant="body2" color="text.secondary" noWrap>{subtitle}</Typography>}
        sx={{
          px: 2.25,
          py: 1.5,
          pb: 1.25,
          borderBottom: "1px solid",
          borderColor: "divider",
          "& .MuiCardHeader-content": {
            minWidth: 0,
          },
          "& .MuiCardHeader-avatar": {
            color: "primary.main",
          },
        }}
      />
      <CardContent sx={{ px: 2.25, py: 2 }}>{children}</CardContent>
    </Card>
  );
}

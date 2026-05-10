import { Box, List, ListItemButton, ListItemIcon, ListItemText, Stack, Typography } from "@mui/material";
import type { ElementType } from "react";
import { APP_VERSION } from "../../../constants/appMeta";
import type { DashboardPageKey } from "../model";

type DashboardSidebarProps = {
  items: Array<{
    id: DashboardPageKey;
    icon: ElementType;
    label: string;
  }>;
  loading: boolean;
  onSelect: (page: DashboardPageKey) => void;
  selectedPage: DashboardPageKey;
  title: string;
  workspaceReadyLabel: string;
  workspaceSyncingLabel: string;
};

export function DashboardSidebar({
  items,
  loading,
  onSelect,
  selectedPage,
  title,
  workspaceReadyLabel,
  workspaceSyncingLabel,
}: DashboardSidebarProps) {
  return (
    <Box component="aside" className="sakiko-sidebar">
      <Stack className="sakiko-sidebar__header" spacing={1.25}>
        <Stack direction="row" spacing={1.25} alignItems="center">
          <Box className="sakiko-sidebar__logo">
            <Box
              component="img"
              className="sakiko-sidebar__logo-img"
              src="/sakiko.png"
              alt=""
            />
          </Box>
          <Box>
            <Typography variant="h6" fontWeight={800} className="sakiko-sidebar__title">{title}</Typography>
          </Box>
        </Stack>

        <Box className={`sakiko-sidebar__status${loading ? " sakiko-sidebar__status--loading" : ""}`}>
          <Box className="sakiko-sidebar__status-dot" />
          <Typography variant="caption" color="text.secondary" fontWeight={700} noWrap>
            {loading ? workspaceSyncingLabel : workspaceReadyLabel}
          </Typography>
        </Box>
      </Stack>

      <List disablePadding className="sakiko-sidebar__nav">
        {items.map((item) => {
          const Icon = item.icon;
          return (
            <ListItemButton
              key={item.id}
              selected={item.id === selectedPage}
              onClick={() => onSelect(item.id)}
              sx={{ px: 1.25, py: 1.25 }}
            >
              <ListItemIcon sx={{ minWidth: 38 }}>
                <Icon fontSize="small" />
              </ListItemIcon>
              <ListItemText
                primary={item.label}
                primaryTypographyProps={{ fontWeight: 700, noWrap: true }}
              />
            </ListItemButton>
          );
        })}
      </List>

      <Box className="sakiko-sidebar__meta">
        <Typography variant="caption" color="text.secondary" className="sakiko-sidebar__version">
          {APP_VERSION}
        </Typography>
      </Box>
    </Box>
  );
}

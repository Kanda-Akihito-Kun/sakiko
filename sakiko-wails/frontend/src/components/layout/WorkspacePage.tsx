import { Box, Stack, Typography } from "@mui/material";
import type { ReactNode } from "react";
import { AppErrorBoundary } from "../shared/AppErrorBoundary";

type WorkspacePageProps = {
  title: string;
  action?: ReactNode;
  children: ReactNode;
  errorBoundaryKey?: string;
  maxWidth?: number | string;
};

export function WorkspacePage({
  title,
  action,
  children,
  errorBoundaryKey,
  maxWidth = 1480,
}: WorkspacePageProps) {
  return (
    <Box className="sakiko-page">
      <Stack
        direction="row"
        spacing={1.5}
        justifyContent="space-between"
        alignItems="center"
        useFlexGap
        flexWrap="wrap"
        className="sakiko-page__header sakiko-content__header"
      >
        <Box sx={{ minWidth: 0 }}>
          <Typography variant="h4" noWrap title={title} fontWeight={800}>
            {title}
          </Typography>
        </Box>

        {action ? (
          <Box className="sakiko-page__action">
            {action}
          </Box>
        ) : null}
      </Stack>

      <Box className="sakiko-page__scroller sakiko-content__body">
        <Box className="sakiko-page__content" sx={{ maxWidth }}>
          <AppErrorBoundary resetKey={errorBoundaryKey} title={`${title} failed to render`}>
            {children}
          </AppErrorBoundary>
        </Box>
      </Box>
    </Box>
  );
}

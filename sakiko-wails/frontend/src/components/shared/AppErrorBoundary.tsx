import ErrorOutlineRounded from "@mui/icons-material/ErrorOutlineRounded";
import RefreshRounded from "@mui/icons-material/RefreshRounded";
import { Alert, Box, Button, Stack, Typography } from "@mui/material";
import { Component, type ErrorInfo, type ReactNode } from "react";

type AppErrorBoundaryProps = {
  children: ReactNode;
  resetKey?: string;
  title?: string;
};

type AppErrorBoundaryState = {
  error: Error | null;
};

export class AppErrorBoundary extends Component<AppErrorBoundaryProps, AppErrorBoundaryState> {
  state: AppErrorBoundaryState = {
    error: null,
  };

  componentDidCatch(error: Error, info: ErrorInfo) {
    console.error("[sakiko] render boundary caught an error", error, info);
  }

  componentDidUpdate(previousProps: AppErrorBoundaryProps) {
    if (previousProps.resetKey !== this.props.resetKey && this.state.error) {
      this.setState({ error: null });
    }
  }

  static getDerivedStateFromError(error: Error): AppErrorBoundaryState {
    return { error };
  }

  render() {
    if (!this.state.error) {
      return this.props.children;
    }

    return (
      <Box
        sx={{
          display: "grid",
          placeItems: "center",
          minHeight: 280,
          px: 2,
        }}
      >
        <Alert
          severity="error"
          icon={<ErrorOutlineRounded />}
          sx={{
            width: "min(100%, 680px)",
            alignItems: "flex-start",
          }}
          action={
            <Button
              color="error"
              size="small"
              startIcon={<RefreshRounded />}
              onClick={() => this.setState({ error: null })}
            >
              Retry
            </Button>
          }
        >
          <Stack spacing={0.75}>
            <Typography fontWeight={700}>
              {this.props.title || "This view could not be rendered."}
            </Typography>
            <Typography variant="body2" sx={{ wordBreak: "break-word" }}>
              {this.state.error.message || "Unknown rendering error"}
            </Typography>
          </Stack>
        </Alert>
      </Box>
    );
  }
}

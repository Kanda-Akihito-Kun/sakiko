import type { ReactNode } from "react";
import { Box } from "@mui/material";

type SectionLayoutProps = {
  children: ReactNode;
  columns?: {
    xs: string;
    md?: string;
    lg?: string;
    xl?: string;
  };
  gap?: number;
  alignItems?: "start" | "stretch";
};

type SectionStackProps = {
  children: ReactNode;
  gap?: number;
};

export function SectionLayout({
  children,
  columns = {
    xs: "minmax(0, 1fr)",
  },
  gap = 2.25,
  alignItems = "start",
}: SectionLayoutProps) {
  return (
    <Box
      sx={{
        display: "grid",
        gap,
        gridTemplateColumns: columns,
        alignItems,
        "& > *": {
          minWidth: 0,
        },
      }}
    >
      {children}
    </Box>
  );
}

export function SectionStack({ children, gap = 2.25 }: SectionStackProps) {
  return (
    <Box
      sx={{
        display: "grid",
        gap,
        alignContent: "start",
        minWidth: 0,
        "& > *": {
          minWidth: 0,
        },
      }}
    >
      {children}
    </Box>
  );
}

export function SectionSpan({ children }: { children: ReactNode }) {
  return <Box sx={{ gridColumn: "1 / -1", minWidth: 0 }}>{children}</Box>;
}

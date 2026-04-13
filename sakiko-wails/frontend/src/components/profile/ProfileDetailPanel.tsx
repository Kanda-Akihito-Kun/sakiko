import DeleteOutlineRounded from "@mui/icons-material/DeleteOutlineRounded";
import DragIndicatorRounded from "@mui/icons-material/DragIndicatorRounded";
import HubRounded from "@mui/icons-material/HubRounded";
import RefreshRounded from "@mui/icons-material/RefreshRounded";
import SearchRounded from "@mui/icons-material/SearchRounded";
import {
  Box,
  Button,
  Card,
  Chip,
  InputAdornment,
  Stack,
  Table,
  TableBody,
  TableCell,
  TableContainer,
  TableHead,
  TableRow,
  TextField,
  Typography,
} from "@mui/material";
import { useEffect, useRef, useState } from "react";
import { DndProvider, useDrag, useDrop } from "react-dnd";
import { HTML5Backend } from "react-dnd-html5-backend";
import { useTranslation } from "react-i18next";
import type { Profile } from "../../types/sakiko";
import type { FilteredProfileNode } from "../../utils/dashboard";
import { shouldUseEmojiFont } from "../../utils/dashboard";
import { EmptyState } from "../shared/EmptyState";
import { SectionCard } from "../shared/SectionCard";

type ProfileDetailPanelProps = {
  activeProfile: Profile | null;
  filteredNodes: FilteredProfileNode[];
  nodeFilter: string;
  submitting: boolean;
  onNodeEnabledChange: (nodeIndex: number, enabled: boolean) => void;
  onNodeMove: (nodeIndex: number, targetIndex: number) => void;
  onNodeFilterChange: (value: string) => void;
  onRefreshProfile: () => void;
  onDeleteProfile: () => void;
};

const profileRowItemType = "sakiko-profile-node-row";

export function ProfileDetailPanel({
  activeProfile,
  filteredNodes,
  nodeFilter,
  submitting,
  onNodeEnabledChange,
  onNodeMove,
  onNodeFilterChange,
  onRefreshProfile,
  onDeleteProfile,
}: ProfileDetailPanelProps) {
  const { t } = useTranslation();
  const [orderedNodes, setOrderedNodes] = useState(filteredNodes);
  const canRefreshProfile = Boolean(activeProfile?.source?.trim());
  const enabledNodeCount = activeProfile?.nodes.filter((node) => node.enabled).length ?? 0;
  const canDragNodes = !submitting && orderedNodes.length > 1;

  useEffect(() => {
    setOrderedNodes(filteredNodes);
  }, [filteredNodes]);

  const movePreviewRow = (dragIndex: number, hoverIndex: number) => {
    setOrderedNodes((current) => reorderFilteredNodes(current, dragIndex, hoverIndex));
  };

  const commitRowMove = async (draggedActualIndex: number) => {
    const currentIndex = orderedNodes.findIndex((item) => item.index === draggedActualIndex);
    if (currentIndex < 0 || !hasFilteredNodeOrderChanged(filteredNodes, orderedNodes)) {
      setOrderedNodes(filteredNodes);
      return;
    }

    const targetIndex = resolveNodeTargetIndex(orderedNodes, currentIndex);
    setOrderedNodes(filteredNodes);
    if (targetIndex === draggedActualIndex) {
      return;
    }
    await onNodeMove(draggedActualIndex, targetIndex);
  };

  return (
    <SectionCard
      title={t("dashboard.profiles.detail.title")}
      subtitle={activeProfile?.source || t("dashboard.profiles.detail.pickProfile")}
      icon={<HubRounded color="primary" />}
      subtitleWrap
    >
      {activeProfile ? (
        <Stack spacing={2}>
          <Box
            sx={{
              display: "grid",
              gridTemplateColumns: "repeat(auto-fit, minmax(210px, 1fr))",
              gap: 1.5,
            }}
          >
            <MetricCard label={t("dashboard.profiles.detail.selected")} value={activeProfile.name} />
            <MetricCard label={t("dashboard.profiles.detail.nodes")} value={`${activeProfile.nodes.length}`} />
            <MetricCard label={t("dashboard.profiles.detail.enabled")} value={`${enabledNodeCount}`} />
            <MetricCard label={t("dashboard.profiles.detail.updated")} value={activeProfile.updatedAt || t("dashboard.profiles.detail.updatedUnknown")} mono />
          </Box>

          <Stack direction="row" spacing={1.5} alignItems="flex-start" useFlexGap sx={{ flexWrap: "wrap" }}>
            <TextField
              value={nodeFilter}
              onChange={(event) => onNodeFilterChange(event.target.value)}
              placeholder={t("dashboard.profiles.detail.filterPlaceholder")}
              sx={{ flex: "1 1 320px", minWidth: 240 }}
              InputProps={{
                startAdornment: (
                  <InputAdornment position="start">
                    <SearchRounded sx={{ color: "text.secondary" }} />
                  </InputAdornment>
                ),
              }}
            />
            <Button
              variant="outlined"
              startIcon={<RefreshRounded />}
              disabled={submitting || !canRefreshProfile}
              onClick={onRefreshProfile}
              sx={{ minWidth: 210, flex: "0 0 auto" }}
            >
              {t("dashboard.profiles.detail.refreshSubscription")}
            </Button>
            <Button
              variant="outlined"
              color="error"
              startIcon={<DeleteOutlineRounded />}
              disabled={submitting}
              onClick={() => {
                if (!activeProfile) {
                  return;
                }
                if (!window.confirm(t("dashboard.profiles.detail.deleteConfirm", { name: activeProfile.name || t("shared.states.unnamedProfile") }))) {
                  return;
                }
                onDeleteProfile();
              }}
              sx={{ minWidth: 132, flex: "0 0 auto" }}
            >
              {t("shared.actions.delete")}
            </Button>
          </Stack>

          <Stack direction="row" spacing={1} useFlexGap sx={{ flexWrap: "wrap", pb: 0.25 }}>
            <Chip label={t("dashboard.profiles.detail.matchingNodes", { count: filteredNodes.length })} color="primary" />
            <Chip label={t("dashboard.profiles.detail.enabledForTasks", { count: enabledNodeCount })} color="success" variant="outlined" />
            <Chip label={activeProfile.source || t("dashboard.profiles.detail.recoveredLocalProfile")} variant="outlined" />
            <Chip
              label={canDragNodes ? t("dashboard.profiles.detail.dragRows") : t("dashboard.profiles.detail.dragRowsDisabled")}
              variant="outlined"
            />
          </Stack>

          <DndProvider backend={HTML5Backend}>
            <TableContainer className="sakiko-table">
              <Table size="small" stickyHeader>
                <TableHead>
                  <TableRow>
                    <TableCell align="center">{t("dashboard.profiles.detail.columns.order")}</TableCell>
                    <TableCell>{t("dashboard.profiles.detail.columns.node")}</TableCell>
                    <TableCell>{t("dashboard.profiles.detail.columns.server")}</TableCell>
                    <TableCell align="right">{t("dashboard.profiles.detail.columns.port")}</TableCell>
                    <TableCell align="right">{t("dashboard.profiles.detail.columns.taskScope")}</TableCell>
                  </TableRow>
                </TableHead>
                <TableBody>
                  {orderedNodes.map((item, visibleIndex) => (
                    <SortableProfileRow
                      key={`${item.node.name}-${item.index}`}
                      item={item}
                      visibleIndex={visibleIndex}
                      canDrag={canDragNodes}
                      controlsDisabled={submitting}
                      onCommitMove={commitRowMove}
                      onNodeEnabledChange={onNodeEnabledChange}
                      onPreviewMove={movePreviewRow}
                    />
                  ))}
                </TableBody>
              </Table>
            </TableContainer>
          </DndProvider>
        </Stack>
      ) : (
        <EmptyState
          title={t("dashboard.profiles.detail.noProfileTitle")}
          description={t("dashboard.profiles.detail.noProfileDescription")}
        />
      )}
    </SectionCard>
  );
}

type MetricCardProps = {
  label: string;
  value: string;
  mono?: boolean;
};

function MetricCard({ label, value, mono = false }: MetricCardProps) {
  return (
    <Card
      variant="outlined"
      sx={{
        p: 1.75,
        bgcolor: "background.default",
        borderColor: "divider",
        minWidth: 0,
      }}
    >
      <Typography variant="subtitle2" color="text.secondary" gutterBottom>
        {label}
      </Typography>
      <Typography
        className={mono ? "sakiko-mono" : undefined}
        fontWeight={600}
        sx={{
          overflowWrap: "anywhere",
          wordBreak: "break-word",
          whiteSpace: "pre-wrap",
        }}
      >
        {value}
      </Typography>
    </Card>
  );
}

type SortableProfileRowProps = {
  item: FilteredProfileNode;
  visibleIndex: number;
  canDrag: boolean;
  controlsDisabled: boolean;
  onCommitMove: (draggedActualIndex: number) => Promise<void>;
  onNodeEnabledChange: (nodeIndex: number, enabled: boolean) => void;
  onPreviewMove: (dragIndex: number, hoverIndex: number) => void;
};

type DragProfileRowItem = {
  actualIndex: number;
  visibleIndex: number;
};

function SortableProfileRow({
  item,
  visibleIndex,
  canDrag,
  controlsDisabled,
  onCommitMove,
  onNodeEnabledChange,
  onPreviewMove,
}: SortableProfileRowProps) {
  const { t } = useTranslation();
  const ref = useRef<HTMLTableRowElement | null>(null);
  const handleRef = useRef<HTMLSpanElement | null>(null);
  const [{ isDragging }, drag] = useDrag(() => ({
    type: profileRowItemType,
    item: { actualIndex: item.index, visibleIndex },
    canDrag,
    collect: (monitor) => ({
      isDragging: monitor.isDragging(),
    }),
  }), [canDrag, item.index, visibleIndex]);

  const [{ handlerId, isOver }, drop] = useDrop<DragProfileRowItem, void, { handlerId: unknown; isOver: boolean }>(() => ({
    accept: profileRowItemType,
    collect: (monitor) => ({
      handlerId: monitor.getHandlerId(),
      isOver: monitor.isOver({ shallow: true }),
    }),
    hover(dragItem, monitor) {
      if (!ref.current || dragItem.visibleIndex === visibleIndex) {
        return;
      }

      const hoverBoundingRect = ref.current.getBoundingClientRect();
      const hoverMiddleY = (hoverBoundingRect.bottom - hoverBoundingRect.top) / 2;
      const clientOffset = monitor.getClientOffset();
      if (!clientOffset) {
        return;
      }

      const hoverClientY = clientOffset.y - hoverBoundingRect.top;
      if (dragItem.visibleIndex < visibleIndex && hoverClientY < hoverMiddleY) {
        return;
      }
      if (dragItem.visibleIndex > visibleIndex && hoverClientY > hoverMiddleY) {
        return;
      }

      onPreviewMove(dragItem.visibleIndex, visibleIndex);
      dragItem.visibleIndex = visibleIndex;
    },
    drop(dragItem, monitor) {
      if (monitor.didDrop()) {
        return;
      }
      void onCommitMove(dragItem.actualIndex);
    },
  }), [onCommitMove, onPreviewMove, visibleIndex]);

  drop(ref);
  drag(handleRef);

  return (
    <TableRow
      ref={ref}
      hover
      data-handler-id={String(handlerId || "")}
      sx={{
        opacity: isDragging ? 0.35 : 1,
        backgroundColor: isOver ? "action.hover" : undefined,
      }}
    >
      <TableCell align="center">
        <Stack direction="row" spacing={0.75} justifyContent="center" alignItems="center">
          <Box
            component="span"
            ref={handleRef}
            sx={{
              display: "inline-flex",
              alignItems: "center",
              color: canDrag ? "action.active" : "action.disabled",
              cursor: canDrag ? "grab" : "default",
            }}
          >
            <DragIndicatorRounded fontSize="small" />
          </Box>
          <Typography variant="body2" className="sakiko-mono" sx={{ minWidth: 24 }}>
            {visibleIndex + 1}
          </Typography>
        </Stack>
      </TableCell>
      <TableCell>
        <Stack spacing={0.25}>
          <Typography fontWeight={600} noWrap className={shouldUseEmojiFont("nodeName", item.node.name) ? "sakiko-emoji" : undefined}>
            {item.node.name}
          </Typography>
          <Stack direction="row" spacing={0.75} sx={{ flexWrap: "wrap", pt: 0.25 }}>
            <Chip
              size="small"
              label={item.node.protocol || t("shared.states.unknown")}
              variant="outlined"
              sx={{ textTransform: "uppercase" }}
            />
            <Chip
              size="small"
              label={
                item.node.udp === true
                  ? t("dashboard.profiles.detail.scopeUdp")
                  : item.node.udp === false
                    ? t("dashboard.profiles.detail.scopeNoUdp")
                    : t("dashboard.profiles.detail.scopeUnknownUdp")
              }
              color={item.node.udp === true ? "success" : "default"}
              variant={item.node.udp === true ? "filled" : "outlined"}
            />
          </Stack>
        </Stack>
      </TableCell>
      <TableCell className="sakiko-mono">{item.node.server || t("shared.states.unknownServer")}</TableCell>
      <TableCell align="right" className="sakiko-mono">
        {item.node.port || "-"}
      </TableCell>
      <TableCell align="right">
        <Button
          size="small"
          variant={item.node.enabled ? "contained" : "outlined"}
          color={item.node.enabled ? "success" : "inherit"}
          disabled={controlsDisabled}
          onClick={() => onNodeEnabledChange(item.index, !item.node.enabled)}
        >
          {item.node.enabled ? t("dashboard.profiles.detail.scopeIncluded") : t("dashboard.profiles.detail.scopeSkipped")}
        </Button>
      </TableCell>
    </TableRow>
  );
}

function reorderFilteredNodes(items: FilteredProfileNode[], dragIndex: number, hoverIndex: number): FilteredProfileNode[] {
  if (dragIndex === hoverIndex || dragIndex < 0 || hoverIndex < 0 || dragIndex >= items.length || hoverIndex >= items.length) {
    return items;
  }

  const next = items.slice();
  const [dragged] = next.splice(dragIndex, 1);
  next.splice(hoverIndex, 0, dragged);
  return next;
}

function hasFilteredNodeOrderChanged(left: FilteredProfileNode[], right: FilteredProfileNode[]): boolean {
  if (left.length !== right.length) {
    return true;
  }
  return left.some((item, index) => item.index !== right[index]?.index);
}

function resolveNodeTargetIndex(rows: FilteredProfileNode[], visibleIndex: number): number {
  const current = rows[visibleIndex];
  if (!current) {
    return -1;
  }

  const previous = rows[visibleIndex - 1];
  if (previous) {
    return previous.index < current.index
      ? previous.index + 1
      : previous.index;
  }

  const next = rows[visibleIndex + 1];
  if (next) {
    return next.index < current.index
      ? next.index
      : Math.max(0, next.index - 1);
  }

  return current.index;
}

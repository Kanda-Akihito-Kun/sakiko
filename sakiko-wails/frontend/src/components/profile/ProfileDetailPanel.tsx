import DragIndicatorRounded from "@mui/icons-material/DragIndicatorRounded";
import HubRounded from "@mui/icons-material/HubRounded";
import SortByAlphaRounded from "@mui/icons-material/SortByAlphaRounded";
import {
  Box,
  Button,
  Card,
  Chip,
  Stack,
  Table,
  TableBody,
  TableCell,
  TableContainer,
  TableHead,
  TableRow,
  Typography,
} from "@mui/material";
import { useEffect, useRef, useState } from "react";
import { DndProvider, useDrag, useDrop } from "react-dnd";
import { HTML5Backend } from "react-dnd-html5-backend";
import { useTranslation } from "react-i18next";
import type { Profile } from "../../types/sakiko";
import type { FilteredProfileNode } from "../../utils/dashboard";
import { extractProfileSubscriptionInfo, formatDataSize, formatDateTime, shouldUseEmojiFont } from "../../utils/dashboard";
import { EmptyState } from "../shared/EmptyState";
import { SectionCard } from "../shared/SectionCard";

type ProfileDetailPanelProps = {
  activeProfile: Profile | null;
  filteredNodes: FilteredProfileNode[];
  submitting: boolean;
  onNodeEnabledChange: (nodeIndex: number, enabled: boolean) => void;
  onNodeMove: (nodeIndex: number, targetIndex: number) => void;
};

const profileRowItemType = "sakiko-profile-node-row";

export function ProfileDetailPanel({
  activeProfile,
  filteredNodes,
  submitting,
  onNodeEnabledChange,
  onNodeMove,
}: ProfileDetailPanelProps) {
  const { t } = useTranslation();
  const [viewOrder, setViewOrder] = useState<"default" | "name">("default");
  const [orderedNodes, setOrderedNodes] = useState(filteredNodes);
  const enabledNodeCount = activeProfile?.nodes.filter((node) => node.enabled).length ?? 0;
  const canDragNodes = !submitting && viewOrder === "default" && orderedNodes.length > 1;
  const subscription = extractProfileSubscriptionInfo(activeProfile?.attributes);

  useEffect(() => {
    setOrderedNodes(viewOrder === "name" ? sortFilteredNodesByName(filteredNodes) : filteredNodes);
  }, [filteredNodes, viewOrder]);

  useEffect(() => {
    setViewOrder("default");
  }, [activeProfile?.id]);

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
      icon={<HubRounded color="primary" />}
    >
      {activeProfile ? (
        <Stack spacing={2}>
          <Box
            sx={{
              display: "grid",
              gridTemplateColumns: "repeat(4, minmax(0, 1fr))",
              gap: 1.5,
            }}
          >
            <MetricCard
              label={t("dashboard.profiles.detail.includedNodes", { defaultValue: "Included Nodes / All Nodes" })}
              value={`${enabledNodeCount} / ${activeProfile.nodes.length}`}
              mono
            />
            <MetricCard
              label={t("dashboard.profiles.detail.remainingTraffic", { defaultValue: "Remaining Traffic" })}
              value={typeof subscription.remainingBytes === "number" ? formatDataSize(subscription.remainingBytes) : "-"}
              mono
            />
            <MetricCard
              label={t("dashboard.profiles.detail.updatedAt", { defaultValue: "Updated At" })}
              value={activeProfile.updatedAt ? formatDateTime(activeProfile.updatedAt) : t("dashboard.profiles.detail.updatedUnknown")}
              mono
            />
            <MetricCard
              label={t("dashboard.profiles.detail.expireAt", { defaultValue: "Expires At" })}
              value={subscription.expiresAt ? formatDateTime(subscription.expiresAt) : "-"}
              mono
            />
          </Box>

          <Stack direction="row" spacing={1} useFlexGap sx={{ flexWrap: "wrap", pb: 0.25 }}>
            <Chip label={t("dashboard.profiles.detail.enabledForTasks", { count: enabledNodeCount })} color="success" variant="outlined" />
            <Chip label={activeProfile.source || t("dashboard.profiles.detail.recoveredLocalProfile")} variant="outlined" />
            <Button
              size="small"
              variant="outlined"
              startIcon={<SortByAlphaRounded />}
              onClick={() => setViewOrder((current) => current === "default" ? "name" : "default")}
            >
              {viewOrder === "default"
                ? t("dashboard.profiles.detail.sortByName", { defaultValue: "Sort By Name" })
                : t("dashboard.profiles.detail.restoreDefaultOrder", { defaultValue: "Restore Default Order" })}
            </Button>
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

function sortFilteredNodesByName(items: FilteredProfileNode[]): FilteredProfileNode[] {
  return [...items].sort((left, right) => left.node.name.localeCompare(right.node.name, "zh-Hans-CN"));
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

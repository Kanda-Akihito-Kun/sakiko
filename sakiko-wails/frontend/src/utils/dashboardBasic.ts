export type FilterableProfileNode = {
  name: string;
  protocol?: string | null;
  server?: string | null;
  port?: string | null;
  udp?: boolean | null;
};

export type FilterableProfile<TNode extends FilterableProfileNode = FilterableProfileNode> = {
  nodes: TNode[];
} | null;

export type FilteredProfileNode<TNode extends FilterableProfileNode = FilterableProfileNode> = {
  index: number;
  node: TNode;
};

export function getFilteredNodes<TNode extends FilterableProfileNode>(
  activeProfile: FilterableProfile<TNode>,
  nodeFilter: string,
): FilteredProfileNode<TNode>[] {
  const keyword = nodeFilter.trim().toLowerCase();
  if (!activeProfile) {
    return [];
  }

  if (!keyword) {
    return activeProfile.nodes.map((node, index) => ({ node, index }));
  }

  return activeProfile.nodes.flatMap((node, index) => (
    [
      node.name,
      node.protocol || "",
      node.server || "",
      node.port || "",
      node.udp === true ? "udp" : node.udp === false ? "no udp" : "",
    ].some((value) => value.toLowerCase().includes(keyword))
      ? [{ node, index }]
      : []
  ));
}

export function formatDateTime(value?: string): string {
  if (!value) {
    return "N/A";
  }

  const date = new Date(value);
  if (Number.isNaN(date.getTime())) {
    return value;
  }

  return new Intl.DateTimeFormat(undefined, {
    year: "numeric",
    month: "short",
    day: "2-digit",
    hour: "2-digit",
    minute: "2-digit",
    second: "2-digit",
    hour12: false,
  }).format(date);
}

export function formatDataSize(bytes?: number): string {
  const value = Number.isFinite(bytes) ? Math.max(0, Number(bytes)) : 0;
  if (value <= 0) {
    return "0 B";
  }

  const units = ["B", "KB", "MB", "GB", "TB"];
  let size = value;
  let index = 0;
  while (size >= 1000 && index < units.length - 1) {
    size /= 1000;
    index += 1;
  }

  const digits = size >= 100 ? 0 : size >= 10 ? 1 : 2;
  return `${size.toFixed(digits)} ${units[index]}`;
}

export type ProfileSubscriptionInfo = {
  remainingBytes?: number;
  expiresAt?: string;
};

export function extractProfileSubscriptionInfo(attributes: unknown): ProfileSubscriptionInfo {
  const info = nestedRecord(attributes, "subscriptionUserinfo");
  if (!info) {
    return {};
  }

  const remaining = numericField(info.remaining);
  const total = numericField(info.total);
  const upload = numericField(info.upload);
  const download = numericField(info.download);
  const expire = numericField(info.expire);
  const expiresAt = stringField(info.expiresAt);

  const resolvedRemaining = remaining ?? (typeof total === "number"
    ? Math.max(0, total - (upload || 0) - (download || 0))
    : undefined);

  return {
    remainingBytes: resolvedRemaining,
    expiresAt: expiresAt || (typeof expire === "number" && expire > 0
      ? new Date(expire * 1000).toISOString()
      : undefined),
  };
}

function nestedRecord(root: unknown, key: string): Record<string, unknown> | null {
  if (!root || typeof root !== "object") {
    return null;
  }
  const value = (root as Record<string, unknown>)[key];
  return value && typeof value === "object" ? value as Record<string, unknown> : null;
}

function numericField(value: unknown): number | undefined {
  if (typeof value === "number" && Number.isFinite(value)) {
    return value;
  }
  if (typeof value === "string") {
    const parsed = Number.parseFloat(value);
    return Number.isFinite(parsed) ? parsed : undefined;
  }
  return undefined;
}

function stringField(value: unknown): string | undefined {
  return typeof value === "string" && value.trim() ? value : undefined;
}

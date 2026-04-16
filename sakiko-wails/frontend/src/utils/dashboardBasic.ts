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
    month: "2-digit",
    day: "2-digit",
    hour: "2-digit",
    minute: "2-digit",
    second: "2-digit",
  }).format(date);
}

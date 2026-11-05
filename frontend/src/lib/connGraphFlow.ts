import { MarkerType, type Edge, type Node } from "@xyflow/svelte";
import { dirCssVar, namesByPidFrom, type GraphSelection, type UiConn } from "./connGraphEngine";
import type { LocalComponent } from "./connGraphClusters";

export type ProcessNodeData = {
  kind: "proc";
  pid: number;
  name: string;
  hasLocal: boolean;
  dimmed: boolean;
};

function resolvedStrokeLocal(): string {
  const fallback = "#7c3aed";
  if (typeof document === "undefined") return fallback;
  const v = getComputedStyle(document.documentElement).getPropertyValue("--local").trim();
  return v || fallback;
}

function edgeStyleCss(dim: boolean, hi: boolean): string {
  const stroke = dirCssVar("local");
  const opacity = dim ? 0.2 : hi ? 1 : 0.75;
  return `stroke: ${stroke}; stroke-width: 3.2px; opacity: ${opacity};`;
}

function connHighlighted(c: UiConn, selection: GraphSelection): boolean {
  if (selection.selectedEdgeId === c.id) return true;
  if (selection.selectedPid != null && (c.pid === selection.selectedPid || c.peerPid === selection.selectedPid)) {
    return true;
  }
  return false;
}

function procRelated(pid: number, edges: UiConn[], selection: GraphSelection): boolean {
  if (selection.selectedPid === pid) return true;
  if (selection.selectedEdgeId) {
    const hit = edges.find((c) => c.id === selection.selectedEdgeId);
    return !!hit && (hit.pid === pid || hit.peerPid === pid);
  }
  if (selection.selectedPid != null) {
    return edges.some(
      (c) =>
        (c.pid === selection.selectedPid || c.peerPid === selection.selectedPid) &&
        (c.pid === pid || c.peerPid === pid),
    );
  }
  return false;
}

function layoutPosition(
  index: number,
  total: number,
  cx: number,
  cy: number,
): { x: number; y: number } {
  if (total <= 8) {
    const radius = Math.min(280, 72 + total * 22);
    const angle = (index / Math.max(total, 1)) * Math.PI * 2 - Math.PI / 2;
    return {
      x: cx + Math.cos(angle) * radius,
      y: cy + Math.sin(angle) * radius,
    };
  }
  const cols = Math.max(3, Math.ceil(Math.sqrt(total)));
  const rows = Math.ceil(total / cols);
  const col = index % cols;
  const row = Math.floor(index / cols);
  const spacingX = 118;
  const spacingY = 88;
  const w = (cols - 1) * spacingX;
  const h = (rows - 1) * spacingY;
  return {
    x: cx - w / 2 + col * spacingX,
    y: cy - h / 2 + row * spacingY,
  };
}

/**
 * Build graph for one local-interconnect component only (purple process↔process edges).
 * External connections never appear in the graph.
 */
export function buildFlowGraph(
  allConns: UiConn[],
  component: LocalComponent | null,
  selection: GraphSelection,
  prevNodes: Node[] = [],
): { nodes: Node[]; edges: Edge[] } {
  if (!component || component.pids.length === 0) {
    return { nodes: [], edges: [] };
  }

  const names = namesByPidFrom(allConns);
  const visible = component.localEdges;
  const hasSelection = selection.selectedPid != null || selection.selectedEdgeId != null;
  const pidSet = new Set(component.pids);

  const localCount = new Map<number, number>();
  for (const c of visible) {
    if (!c.peerPid) continue;
    localCount.set(c.pid, (localCount.get(c.pid) ?? 0) + 1);
    localCount.set(c.peerPid, (localCount.get(c.peerPid) ?? 0) + 1);
  }

  const oldPos = Object.fromEntries(prevNodes.map((n) => [n.id, n.position]));
  const cx = 400;
  const cy = 280;

  const nodes: Node[] = [];
  const procList = [...component.pids].sort((a, b) => a - b);
  procList.forEach((pid, i) => {
    if (!pidSet.has(pid)) return;
    const id = `proc:${pid}`;
    const name = names.get(pid) ?? "?";
    const related = procRelated(pid, visible, selection);
    nodes.push({
      id,
      type: "process",
      position: oldPos[id] ?? layoutPosition(i, procList.length, cx, cy),
      data: {
        kind: "proc",
        pid,
        name,
        hasLocal: (localCount.get(pid) ?? 0) > 0,
        dimmed: hasSelection && !related,
      } satisfies ProcessNodeData,
      selected: selection.selectedPid === pid,
    });
  });

  const markerColor = resolvedStrokeLocal();
  const edges: Edge[] = visible
    .filter((c) => c.peerPid != null && pidSet.has(c.pid) && pidSet.has(c.peerPid))
    .map((c) => {
      const hi = connHighlighted(c, selection);
      const dim = hasSelection && !hi;
      return {
        id: c.id,
        source: `proc:${c.pid}`,
        target: `proc:${c.peerPid}`,
        type: "smoothstep",
        animated: false,
        style: edgeStyleCss(dim, hi),
        markerEnd: { type: MarkerType.ArrowClosed, color: markerColor },
        data: { dir: "local", connId: c.id },
      } satisfies Edge;
    });

  return { nodes, edges };
}

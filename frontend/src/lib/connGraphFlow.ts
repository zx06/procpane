import Dagre from "@dagrejs/dagre";
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

/** Match ConnGraphProcessNode card (dagre uses center anchor; xyflow uses top-left). */
const PROC_NODE_W = 112;
const PROC_NODE_H = 54;

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

/** Layered LR layout; acyclicer helps when local sockets form cycles. */
function layoutWithDagre(nodes: Node[], edges: Edge[]): Node[] {
  if (nodes.length === 0) return nodes;
  const g = new Dagre.graphlib.Graph();
  g.setDefaultEdgeLabel(() => ({}));
  g.setGraph({
    rankdir: "LR",
    nodesep: 56,
    ranksep: 88,
    marginx: 32,
    marginy: 32,
    acyclicer: "greedy",
  });

  for (const node of nodes) {
    g.setNode(node.id, { width: PROC_NODE_W, height: PROC_NODE_H });
  }
  for (const edge of edges) {
    g.setEdge(edge.source, edge.target);
  }
  Dagre.layout(g);

  return nodes.map((node) => {
    const pos = g.node(node.id);
    if (!pos) return node;
    return {
      ...node,
      position: {
        x: pos.x - PROC_NODE_W / 2,
        y: pos.y - PROC_NODE_H / 2,
      },
    };
  });
}

export type BuildFlowGraphOptions = {
  /** When false, reuse positions from prevNodes for matching ids. */
  relayout?: boolean;
};

/**
 * Build graph for one local-interconnect component only (purple process↔process edges).
 * External connections never appear in the graph.
 */
export function buildFlowGraph(
  allConns: UiConn[],
  component: LocalComponent | null,
  selection: GraphSelection,
  prevNodes: Node[] = [],
  options: BuildFlowGraphOptions = {},
): { nodes: Node[]; edges: Edge[] } {
  if (!component || component.pids.length === 0) {
    return { nodes: [], edges: [] };
  }

  const relayout = options.relayout !== false;
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

  let nodes: Node[] = [];
  const procList = [...component.pids].sort((a, b) => a - b);
  procList.forEach((pid) => {
    if (!pidSet.has(pid)) return;
    const id = `proc:${pid}`;
    const name = names.get(pid) ?? "?";
    const related = procRelated(pid, visible, selection);
    nodes.push({
      id,
      type: "process",
      position: { x: 0, y: 0 },
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
        sourceHandle: "source-right",
        targetHandle: "target-left",
        type: "smoothstep",
        animated: false,
        style: edgeStyleCss(dim, hi),
        markerEnd: { type: MarkerType.ArrowClosed, color: markerColor },
        data: { dir: "local", connId: c.id },
      } satisfies Edge;
    });

  if (relayout) {
    nodes = layoutWithDagre(nodes, edges);
  } else {
    nodes = nodes.map((n) => {
      const p = oldPos[n.id];
      return p ? { ...n, position: p } : n;
    });
  }

  return { nodes, edges };
}

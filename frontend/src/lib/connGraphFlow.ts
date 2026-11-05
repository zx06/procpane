import { MarkerType, type Edge, type Node } from "@xyflow/svelte";
import {
  connVisible,
  dirCssVar,
  graphEdgeRows,
  namesByPidFrom,
  type DirUi,
  type GraphFilters,
  type GraphSelection,
  type UiConn,
} from "./connGraphEngine";

export type ProcessNodeData = {
  kind: "proc";
  pid: number;
  name: string;
  hasLocal: boolean;
  dimmed: boolean;
};

export type StubNodeData = {
  kind: "ext";
  label: string;
  dir: DirUi;
  dimmed: boolean;
};

/** Resolve theme token to a concrete color for SVG markers (markers do not always resolve var()). */
function resolvedStroke(dir: DirUi): string {
  const fallback = dir === "local" ? "#7c3aed" : dir === "in" ? "#059669" : "#64748b";
  if (typeof document === "undefined") return fallback;
  const token = dir === "local" ? "--local" : dir === "in" ? "--in-edge" : "--out-edge";
  const v = getComputedStyle(document.documentElement).getPropertyValue(token).trim();
  return v || fallback;
}

function edgeStyleCss(dir: DirUi, dim: boolean, hi: boolean): string {
  const stroke = dirCssVar(dir);
  const width = dir === "local" ? 3.2 : 1.6;
  const opacity = dim ? 0.2 : hi ? 1 : 0.75;
  return `stroke: ${stroke}; stroke-width: ${width}px; opacity: ${opacity};`;
}

function connHighlighted(c: UiConn, selection: GraphSelection): boolean {
  if (selection.selectedEdgeId === c.id) return true;
  if (selection.selectedPid != null && (c.pid === selection.selectedPid || c.peerPid === selection.selectedPid)) {
    return true;
  }
  return false;
}

function procRelated(pid: number, visible: UiConn[], selection: GraphSelection): boolean {
  if (selection.selectedPid === pid) return true;
  if (selection.selectedEdgeId) {
    const hit = visible.find((c) => c.id === selection.selectedEdgeId);
    return !!hit && (hit.pid === pid || hit.peerPid === pid);
  }
  if (selection.selectedPid != null) {
    return visible.some(
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
  if (total <= 14) {
    const radius = Math.min(520, 96 + total * 26);
    const angle = (index / Math.max(total, 1)) * Math.PI * 2 - Math.PI / 2;
    return {
      x: cx + Math.cos(angle) * radius,
      y: cy + Math.sin(angle) * radius,
    };
  }
  const cols = Math.max(4, Math.ceil(Math.sqrt(total * 1.4)));
  const rows = Math.ceil(total / cols);
  const col = index % cols;
  const row = Math.floor(index / cols);
  const spacingX = 128;
  const spacingY = 96;
  const w = (cols - 1) * spacingX;
  const h = (rows - 1) * spacingY;
  return {
    x: cx - w / 2 + col * spacingX,
    y: cy - h / 2 + row * spacingY,
  };
}

/** Graph-visible rows: deduped locals, skip unpaired local (no peer PID — no 127.0.0.1 stub hairball). */
function graphVisibleRows(all: UiConn[], filters: GraphFilters, names: Map<number, string>): UiConn[] {
  return graphEdgeRows(all)
    .filter((c) => connVisible(c, filters, names))
    .filter((c) => !(c.dir === "local" && !c.peerPid));
}

/** Build Svelte Flow nodes/edges from filtered connections; preserves prior node positions by id. */
export function buildFlowGraph(
  allConns: UiConn[],
  filters: GraphFilters,
  selection: GraphSelection,
  prevNodes: Node[] = [],
): { nodes: Node[]; edges: Edge[] } {
  const names = namesByPidFrom(allConns);
  const visible = graphVisibleRows(allConns, filters, names);
  const hasSelection = selection.selectedPid != null || selection.selectedEdgeId != null;

  const procIds = new Set<number>();
  visible.forEach((c) => {
    procIds.add(c.pid);
    if (c.peerPid) procIds.add(c.peerPid);
  });

  const localCount = new Map<number, number>();
  for (const c of allConns) {
    if (c.dir !== "local" || !c.peerPid || !connVisible(c, filters, names)) continue;
    localCount.set(c.pid, (localCount.get(c.pid) ?? 0) + 1);
    localCount.set(c.peerPid, (localCount.get(c.peerPid) ?? 0) + 1);
  }

  const oldPos = Object.fromEntries(prevNodes.map((n) => [n.id, n.position]));
  const cx = 480;
  const cy = 320;

  const nodes: Node[] = [];
  const procList = [...procIds];
  procList.forEach((pid, i) => {
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

  const edges: Edge[] = [];
  visible.forEach((c, idx) => {
    const markerColor = resolvedStroke(c.dir);
    const hi = connHighlighted(c, selection);
    const dim = hasSelection && !hi;
    const style = edgeStyleCss(c.dir, dim, hi);

    if (c.dir === "local" && c.peerPid) {
      edges.push({
        id: c.id,
        source: `proc:${c.pid}`,
        target: `proc:${c.peerPid}`,
        type: "smoothstep",
        animated: false,
        style,
        markerEnd: { type: MarkerType.ArrowClosed, color: markerColor },
        data: { dir: c.dir, connId: c.id },
      });
      return;
    }

    if (c.dir === "local") {
      return;
    }

    const stubId = `ext:${c.id}`;
    if (!nodes.some((n) => n.id === stubId)) {
      const src = nodes.find((n) => n.id === `proc:${c.pid}`);
      const bx = src?.position.x ?? cx;
      const by = src?.position.y ?? cy;
      const ang = Math.atan2(by - cy, bx - cx) + ((idx % 5) - 2) * 0.25;
      nodes.push({
        id: stubId,
        type: "stub",
        position:
          oldPos[stubId] ?? {
            x: bx + Math.cos(ang) * 130,
            y: by + Math.sin(ang) * 130,
          },
        data: {
          kind: "ext",
          label: c.remote,
          dir: c.dir,
          dimmed: hasSelection && !hi,
        } satisfies StubNodeData,
      });
    }

    const source = c.dir === "out" ? `proc:${c.pid}` : stubId;
    const target = c.dir === "out" ? stubId : `proc:${c.pid}`;
    edges.push({
      id: c.id,
      source,
      target,
      type: "smoothstep",
      animated: false,
      style,
      markerEnd: { type: MarkerType.ArrowClosed, color: markerColor },
      data: { dir: c.dir, connId: c.id },
    });
  });

  return { nodes, edges };
}

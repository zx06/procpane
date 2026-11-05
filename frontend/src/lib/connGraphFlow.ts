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

function edgeStroke(dir: DirUi): string {
  return dirCssVar(dir);
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
  radius: number,
): { x: number; y: number } {
  const angle = (index / Math.max(total, 1)) * Math.PI * 2 - Math.PI / 2;
  return {
    x: cx + Math.cos(angle) * radius + (Math.random() - 0.5) * 24,
    y: cy + Math.sin(angle) * radius + (Math.random() - 0.5) * 24,
  };
}

/** Build Svelte Flow nodes/edges from filtered connections; preserves prior node positions by id. */
export function buildFlowGraph(
  allConns: UiConn[],
  filters: GraphFilters,
  selection: GraphSelection,
  prevNodes: Node[] = [],
): { nodes: Node[]; edges: Edge[] } {
  const names = namesByPidFrom(allConns);
  const visible = graphEdgeRows(allConns).filter((c) => connVisible(c, filters, names));
  const hasSelection = selection.selectedPid != null || selection.selectedEdgeId != null;

  const procIds = new Set<number>();
  visible.forEach((c) => {
    procIds.add(c.pid);
    if (c.peerPid) procIds.add(c.peerPid);
  });

  const localCount = new Map<number, number>();
  for (const c of allConns) {
    if (c.dir !== "local" || !connVisible(c, filters, names)) continue;
    localCount.set(c.pid, (localCount.get(c.pid) ?? 0) + 1);
    if (c.peerPid) localCount.set(c.peerPid, (localCount.get(c.peerPid) ?? 0) + 1);
  }

  const oldPos = Object.fromEntries(prevNodes.map((n) => [n.id, n.position]));
  const cx = 420;
  const cy = 280;
  const radius = 220;

  const nodes: Node[] = [];
  const procList = [...procIds];
  procList.forEach((pid, i) => {
    const id = `proc:${pid}`;
    const name = names.get(pid) ?? "?";
    const related = procRelated(pid, visible, selection);
    nodes.push({
      id,
      type: "process",
      position: oldPos[id] ?? layoutPosition(i, procList.length, cx, cy, radius),
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
    const color = edgeStroke(c.dir);
    const hi = connHighlighted(c, selection);
    const dim = hasSelection && !hi;
    const baseStyle = {
      stroke: color,
      strokeWidth: c.dir === "local" ? 3.2 : 1.6,
      opacity: dim ? 0.2 : hi ? 1 : 0.75,
    };

    if (c.dir === "local" && c.peerPid) {
      edges.push({
        id: c.id,
        source: `proc:${c.pid}`,
        target: `proc:${c.peerPid}`,
        type: "smoothstep",
        animated: false,
        style: baseStyle,
        markerEnd: { type: MarkerType.ArrowClosed, color },
        data: { dir: c.dir, connId: c.id },
      });
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
            x: bx + Math.cos(ang) * 120,
            y: by + Math.sin(ang) * 120,
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
      style: baseStyle,
      markerEnd: { type: MarkerType.ArrowClosed, color },
      data: { dir: c.dir, connId: c.id },
    });
  });

  return { nodes, edges };
}

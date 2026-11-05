import type { Connection } from "./api";

export type DirUi = "local" | "in" | "out";

export type GraphFilters = {
  localOnly: boolean;
  inOnly: boolean;
  outOnly: boolean;
  hideBrowser: boolean;
  search: string;
};

export type UiConn = {
  id: string;
  pid: number;
  name: string;
  local: string;
  remote: string;
  dir: DirUi;
  peerPid: number | null;
  peerName: string;
};

export function dirFromModel(d: string): DirUi {
  if (d === "local") return "local";
  if (d === "inbound") return "in";
  return "out";
}

export function isBrowserProcess(name: string): boolean {
  const n = name.toLowerCase();
  return /chrome|chromium|firefox|msedge|edge|safari|brave|vivaldi|arc|google chrome/.test(
    n,
  );
}

export function toUiConn(c: Connection): UiConn {
  return {
    id: c.id,
    pid: c.pid,
    name: c.name,
    local: `${c.localHost}:${c.localPort}`,
    remote: `${c.remoteHost}:${c.remotePort}`,
    dir: dirFromModel(c.direction),
    peerPid: c.peerPid > 0 ? c.peerPid : null,
    peerName: c.peerName || "",
  };
}

function canonicalLocalKey(c: UiConn): string {
  const a = `${c.local}`;
  const b = `${c.remote}`;
  const [x, y] = a < b ? [a, b] : [b, a];
  return `${x}|${y}|${Math.min(c.pid, c.peerPid ?? c.pid)}|${Math.max(c.pid, c.peerPid ?? c.pid)}`;
}

/** One row per undirected local edge for the graph (table still shows all rows). */
export function graphEdgeRows(all: UiConn[]): UiConn[] {
  const seen = new Set<string>();
  return all.filter((c) => {
    if (c.dir !== "local" || !c.peerPid) return true;
    const key = canonicalLocalKey(c);
    if (seen.has(key)) return false;
    seen.add(key);
    return true;
  });
}

export function connVisible(
  c: UiConn,
  filters: GraphFilters,
  namesByPid: Map<number, string>,
): boolean {
  const name = namesByPid.get(c.pid) ?? c.name;
  if (filters.hideBrowser && isBrowserProcess(name)) return false;
  if (c.peerPid) {
    const peer = namesByPid.get(c.peerPid) ?? c.peerName;
    if (filters.hideBrowser && isBrowserProcess(peer)) return false;
  }
  if (filters.localOnly && c.dir !== "local") return false;
  if (filters.inOnly && c.dir !== "in") return false;
  if (filters.outOnly && c.dir !== "out") return false;
  const q = filters.search.trim().toLowerCase();
  if (q) {
    const hay = [name, c.pid, c.local, c.remote, c.peerName, c.dir].join(" ").toLowerCase();
    if (!hay.includes(q)) return false;
  }
  return true;
}

type ProcNode = {
  id: string;
  kind: "proc";
  pid: number;
  name: string;
  x: number;
  y: number;
  vx: number;
  vy: number;
  r: number;
};

type ExtNode = {
  id: string;
  kind: "ext";
  label: string;
  x: number;
  y: number;
  vx: number;
  vy: number;
  r: number;
  dir: DirUi;
};

type GNode = ProcNode | ExtNode;

type GEdge = {
  id: string;
  source: string;
  target: string;
  dir: DirUi;
  conn: UiConn;
  s?: GNode;
  t?: GNode;
};

export type GraphSelection = {
  selectedPid: number | null;
  selectedEdgeId: string | null;
};

function cssVar(name: string): string {
  return getComputedStyle(document.documentElement).getPropertyValue(name).trim();
}

function shortName(name: string, max = 18): string {
  if (name.length <= max) return name;
  return name.slice(0, max - 1) + "…";
}

function edgeColor(dir: DirUi): string {
  const map: Record<DirUi, string> = {
    local: cssVar("--local") || "#7c3aed",
    in: cssVar("--in-edge") || "#059669",
    out: cssVar("--out-edge") || "#64748b",
  };
  return map[dir] || map.out;
}

function distToSeg(px: number, py: number, x1: number, y1: number, x2: number, y2: number) {
  const dx = x2 - x1;
  const dy = y2 - y1;
  const len2 = dx * dx + dy * dy || 1;
  let t = ((px - x1) * dx + (py - y1) * dy) / len2;
  t = Math.max(0, Math.min(1, t));
  const qx = x1 + t * dx;
  const qy = y1 + t * dy;
  return Math.hypot(px - qx, py - qy);
}

export class ConnGraphCanvas {
  private canvas: HTMLCanvasElement;
  private wrap: HTMLElement;
  private ctx: CanvasRenderingContext2D;
  private nodes: GNode[] = [];
  private edges: GEdge[] = [];
  private transform = { x: 0, y: 0, k: 1 };
  private drag:
    | { node: GNode; moved: boolean }
    | { pan: true; sx: number; sy: number; tx: number; ty: number }
    | null = null;
  private hoverNode: GNode | null = null;
  private hoverEdge: GEdge | null = null;
  private allConns: UiConn[] = [];
  private filters: GraphFilters = {
    localOnly: false,
    inOnly: false,
    outOnly: false,
    hideBrowser: true,
    search: "",
  };
  private selection: GraphSelection = { selectedPid: null, selectedEdgeId: null };
  private localCounts = new Map<number, number>();
  private onSelect: (sel: GraphSelection) => void;
  private ro: ResizeObserver;
  private bound = false;

  constructor(
    canvas: HTMLCanvasElement,
    wrap: HTMLElement,
    onSelect: (sel: GraphSelection) => void,
  ) {
    this.canvas = canvas;
    this.wrap = wrap;
    this.ctx = canvas.getContext("2d")!;
    this.onSelect = onSelect;
    this.ro = new ResizeObserver(() => this.resize());
    this.ro.observe(wrap);
  }

  bindEvents() {
    if (this.bound) return;
    this.bound = true;
    this.canvas.addEventListener("pointerdown", this.onPointerDown);
    this.canvas.addEventListener("pointermove", this.onPointerMove);
    this.canvas.addEventListener("pointerup", this.onPointerUp);
    this.canvas.addEventListener("wheel", this.onWheel, { passive: false });
  }

  destroy() {
    this.ro.disconnect();
    if (!this.bound) return;
    this.canvas.removeEventListener("pointerdown", this.onPointerDown);
    this.canvas.removeEventListener("pointermove", this.onPointerMove);
    this.canvas.removeEventListener("pointerup", this.onPointerUp);
    this.canvas.removeEventListener("wheel", this.onWheel);
  }

  setData(conns: UiConn[], filters: GraphFilters) {
    this.allConns = conns;
    this.filters = filters;
    this.rebuildGraph();
  }

  setSelection(sel: GraphSelection) {
    this.selection = sel;
    this.draw();
  }

  private namesByPid(): Map<number, string> {
    const m = new Map<number, string>();
    for (const c of this.allConns) {
      m.set(c.pid, c.name);
      if (c.peerPid) m.set(c.peerPid, c.peerName || m.get(c.peerPid) || "?");
    }
    return m;
  }

  private visibleRows(): UiConn[] {
    const names = this.namesByPid();
    return graphEdgeRows(this.allConns).filter((c) => connVisible(c, this.filters, names));
  }

  private rebuildGraph() {
    const visible = this.visibleRows();
    const names = this.namesByPid();
    const procIds = new Set<number>();
    visible.forEach((c) => {
      procIds.add(c.pid);
      if (c.peerPid) procIds.add(c.peerPid);
    });

    this.localCounts = new Map();
    for (const c of this.allConns.filter((x) => x.dir === "local" && connVisible(x, this.filters, names))) {
      this.localCounts.set(c.pid, (this.localCounts.get(c.pid) ?? 0) + 1);
      if (c.peerPid) {
        this.localCounts.set(c.peerPid, (this.localCounts.get(c.peerPid) ?? 0) + 1);
      }
    }

    const oldPos = Object.fromEntries(this.nodes.map((n) => [n.id, { x: n.x, y: n.y, vx: n.vx, vy: n.vy }]));
    const W = this.wrap.clientWidth || 600;
    const H = this.wrap.clientHeight || 400;
    const cx = W / 2;
    const cy = H / 2;

    this.nodes = [];
    this.edges = [];

    [...procIds].forEach((pid, i) => {
      const id = "proc:" + pid;
      const prev = oldPos[id];
      const name = names.get(pid) ?? "?";
      const angle = (i / Math.max(procIds.size, 1)) * Math.PI * 2 - Math.PI / 2;
      const rad = Math.min(W, H) * 0.28;
      this.nodes.push({
        id,
        kind: "proc",
        pid,
        name,
        x: prev ? prev.x : cx + Math.cos(angle) * rad + (Math.random() - 0.5) * 20,
        y: prev ? prev.y : cy + Math.sin(angle) * rad + (Math.random() - 0.5) * 20,
        vx: prev ? prev.vx : 0,
        vy: prev ? prev.vy : 0,
        r: 22,
      });
    });

    visible.forEach((c, idx) => {
      if (c.dir === "local" && c.peerPid) {
        this.edges.push({
          id: c.id,
          source: "proc:" + c.pid,
          target: "proc:" + c.peerPid,
          dir: "local",
          conn: c,
        });
        return;
      }
      const stubId = "ext:" + c.id;
      const prev = oldPos[stubId];
      const src = this.nodes.find((n) => n.id === "proc:" + c.pid);
      const baseX = src ? src.x : cx;
      const baseY = src ? src.y : cy;
      const ang = Math.atan2(baseY - cy, baseX - cx) + ((idx % 5) - 2) * 0.25;
      this.nodes.push({
        id: stubId,
        kind: "ext",
        label: c.remote,
        x: prev ? prev.x : baseX + Math.cos(ang) * 90,
        y: prev ? prev.y : baseY + Math.sin(ang) * 90,
        vx: 0,
        vy: 0,
        r: 8,
        dir: c.dir,
      });
      this.edges.push({
        id: c.id,
        source: c.dir === "out" ? "proc:" + c.pid : stubId,
        target: c.dir === "out" ? stubId : "proc:" + c.pid,
        dir: c.dir,
        conn: c,
      });
    });

    const byId = Object.fromEntries(this.nodes.map((n) => [n.id, n]));
    this.edges.forEach((e) => {
      e.s = byId[e.source];
      e.t = byId[e.target];
    });

    this.runSim(90);
    this.draw();
  }

  private runSim(ticks: number) {
    const W = this.wrap.clientWidth || 600;
    const H = this.wrap.clientHeight || 400;
    for (let k = 0; k < ticks; k++) {
      for (let i = 0; i < this.nodes.length; i++) {
        for (let j = i + 1; j < this.nodes.length; j++) {
          const a = this.nodes[i];
          const b = this.nodes[j];
          let dx = a.x - b.x;
          let dy = a.y - b.y;
          const dist2 = dx * dx + dy * dy || 1;
          const dist = Math.sqrt(dist2);
          const force = (a.kind === "ext" || b.kind === "ext" ? 1800 : 4200) / dist2;
          const fx = (dx / dist) * force;
          const fy = (dy / dist) * force;
          a.vx += fx;
          a.vy += fy;
          b.vx -= fx;
          b.vy -= fy;
        }
      }
      this.edges.forEach((e) => {
        if (!e.s || !e.t) return;
        const dx = e.t.x - e.s.x;
        const dy = e.t.y - e.s.y;
        const dist = Math.sqrt(dx * dx + dy * dy) || 1;
        const ideal = e.dir === "local" ? 110 : 75;
        const f = (dist - ideal) * 0.045;
        const fx = (dx / dist) * f;
        const fy = (dy / dist) * f;
        e.s.vx += fx;
        e.s.vy += fy;
        e.t.vx -= fx;
        e.t.vy -= fy;
      });
      this.nodes.forEach((n) => {
        n.vx += (W / 2 - n.x) * 0.008;
        n.vy += (H / 2 - n.y) * 0.008;
        n.vx *= 0.82;
        n.vy *= 0.82;
        if (this.drag && "node" in this.drag && this.drag.node === n) {
          n.vx = 0;
          n.vy = 0;
        } else {
          n.x += n.vx;
          n.y += n.vy;
        }
        n.x = Math.max(30, Math.min(W - 30, n.x));
        n.y = Math.max(30, Math.min(H - 30, n.y));
      });
    }
  }

  resize() {
    const r = this.wrap.getBoundingClientRect();
    const s = window.devicePixelRatio || 1;
    this.canvas.width = Math.floor(r.width * s);
    this.canvas.height = Math.floor(r.height * s);
    this.canvas.style.width = r.width + "px";
    this.canvas.style.height = r.height + "px";
    this.ctx.setTransform(s, 0, 0, s, 0, 0);
    this.draw();
  }

  private isHighlightedEdge(e: GEdge): boolean {
    const { selectedPid, selectedEdgeId } = this.selection;
    if (selectedEdgeId && e.id === selectedEdgeId) return true;
    if (selectedPid != null) {
      const c = e.conn;
      if (c.pid === selectedPid || c.peerPid === selectedPid) return true;
    }
    if (this.hoverEdge && this.hoverEdge.id === e.id) return true;
    return false;
  }

  private isDimmedEdge(e: GEdge): boolean {
    const { selectedPid, selectedEdgeId } = this.selection;
    if (selectedPid == null && !selectedEdgeId) return false;
    return !this.isHighlightedEdge(e);
  }

  draw() {
    const W = this.wrap.clientWidth;
    const H = this.wrap.clientHeight;
    const ctx = this.ctx;
    ctx.clearRect(0, 0, W, H);

    ctx.save();
    ctx.translate(this.transform.x, this.transform.y);
    ctx.scale(this.transform.k, this.transform.k);

    this.edges.forEach((e) => {
      if (!e.s || !e.t) return;
      const hi = this.isHighlightedEdge(e);
      const dim = this.isDimmedEdge(e);
      ctx.save();
      ctx.globalAlpha = dim ? 0.18 : hi ? 1 : 0.72;
      ctx.beginPath();
      ctx.moveTo(e.s.x, e.s.y);
      ctx.lineTo(e.t.x, e.t.y);
      ctx.strokeStyle = edgeColor(e.dir);
      ctx.lineWidth = (e.dir === "local" ? 3.2 : 1.6) * (hi ? 1.35 : 1);
      if (e.dir === "local") ctx.lineCap = "round";
      ctx.stroke();

      const dx = e.t.x - e.s.x;
      const dy = e.t.y - e.s.y;
      const ang = Math.atan2(dy, dx);
      const tipDist = e.t.r + 4;
      const ax = e.t.x - Math.cos(ang) * tipDist;
      const ay = e.t.y - Math.sin(ang) * tipDist;
      const ah = 6;
      ctx.beginPath();
      ctx.moveTo(ax, ay);
      ctx.lineTo(ax - ah * Math.cos(ang - 0.4), ay - ah * Math.sin(ang - 0.4));
      ctx.lineTo(ax - ah * Math.cos(ang + 0.4), ay - ah * Math.sin(ang + 0.4));
      ctx.closePath();
      ctx.fillStyle = edgeColor(e.dir);
      ctx.fill();
      ctx.restore();
    });

    const { selectedPid, selectedEdgeId } = this.selection;
    this.nodes.forEach((n) => {
      const selected = n.kind === "proc" && n.pid === selectedPid;
      const related =
        selectedPid != null &&
        (this.edges.some((e) => this.isHighlightedEdge(e) && (e.s === n || e.t === n)) ||
          (selectedEdgeId &&
            this.edges.some((e) => e.id === selectedEdgeId && (e.s === n || e.t === n))));
      const dim =
        selectedPid != null &&
        !selected &&
        !related &&
        !(selectedEdgeId && this.edges.some((e) => e.id === selectedEdgeId && (e.s === n || e.t === n)));

      ctx.save();
      ctx.globalAlpha = dim ? 0.25 : 1;

      if (n.kind === "proc") {
        ctx.beginPath();
        ctx.arc(n.x, n.y + 1.5, n.r, 0, Math.PI * 2);
        ctx.fillStyle = "rgba(0,0,0,0.08)";
        ctx.fill();

        ctx.beginPath();
        ctx.arc(n.x, n.y, n.r, 0, Math.PI * 2);
        ctx.fillStyle = cssVar("--pop-elev") || "#fff";
        ctx.fill();
        ctx.lineWidth = selected ? 2.5 : 1.25;
        ctx.strokeStyle = selected
          ? cssVar("--accent") || "#007aff"
          : cssVar("--hair") || "rgba(0,0,0,.12)";
        ctx.stroke();

        if ((this.localCounts.get(n.pid) ?? 0) > 0) {
          ctx.beginPath();
          ctx.arc(n.x, n.y, n.r - 3.5, 0, Math.PI * 2);
          ctx.strokeStyle = cssVar("--local") || "#7c3aed";
          ctx.globalAlpha = dim ? 0.2 : 0.35;
          ctx.lineWidth = 2;
          ctx.stroke();
          ctx.globalAlpha = dim ? 0.25 : 1;
        }

        ctx.fillStyle = cssVar("--text") || "#1c2434";
        ctx.font = "650 11px " + (cssVar("--sans") || "sans-serif");
        ctx.textAlign = "center";
        ctx.textBaseline = "middle";
        ctx.fillText(shortName(n.name, 14), n.x, n.y - 5);
        ctx.fillStyle = cssVar("--muted") || "#6b778c";
        ctx.font = "500 9px " + (cssVar("--mono") || "monospace");
        ctx.fillText(String(n.pid), n.x, n.y + 8);
      } else {
        ctx.beginPath();
        ctx.arc(n.x, n.y, n.r, 0, Math.PI * 2);
        ctx.fillStyle =
          n.dir === "in" ? cssVar("--ok-soft") || "rgba(5,150,105,.2)" : "rgba(100,116,139,.16)";
        ctx.fill();
        ctx.lineWidth = 1;
        ctx.strokeStyle =
          n.dir === "in" ? cssVar("--in-edge") || "#059669" : cssVar("--stub") || "#94a3b8";
        ctx.stroke();
        ctx.fillStyle = cssVar("--muted") || "#6b778c";
        ctx.font = "500 9px " + (cssVar("--mono") || "monospace");
        ctx.textAlign = "center";
        ctx.textBaseline = "top";
        ctx.fillText(shortName(n.label, 18), n.x, n.y + n.r + 3);
      }
      ctx.restore();
    });

    ctx.restore();
  }

  private screenToWorld(sx: number, sy: number) {
    return {
      x: (sx - this.transform.x) / this.transform.k,
      y: (sy - this.transform.y) / this.transform.k,
    };
  }

  private hitTest(sx: number, sy: number) {
    const p = this.screenToWorld(sx, sy);
    for (let i = this.nodes.length - 1; i >= 0; i--) {
      const n = this.nodes[i];
      const dx = p.x - n.x;
      const dy = p.y - n.y;
      if (dx * dx + dy * dy <= (n.r + 2) * (n.r + 2)) return { type: "node" as const, node: n };
    }
    let best: GEdge | null = null;
    let bestD = 6 / this.transform.k;
    this.edges.forEach((e) => {
      if (!e.s || !e.t) return;
      const d = distToSeg(p.x, p.y, e.s.x, e.s.y, e.t.x, e.t.y);
      if (d < bestD) {
        bestD = d;
        best = e;
      }
    });
    if (best) return { type: "edge" as const, edge: best };
    return null;
  }

  private onPointerDown = (ev: PointerEvent) => {
    const rect = this.canvas.getBoundingClientRect();
    const sx = ev.clientX - rect.left;
    const sy = ev.clientY - rect.top;
    const hit = this.hitTest(sx, sy);
    if (hit?.type === "node") {
      this.drag = { node: hit.node, moved: false };
      this.canvas.setPointerCapture(ev.pointerId);
      this.canvas.classList.add("dragging");
      if (hit.node.kind === "proc") {
        const sel = { selectedPid: hit.node.pid, selectedEdgeId: null as string | null };
        this.selection = sel;
        this.onSelect(sel);
        this.draw();
      }
    } else if (hit?.type === "edge") {
      const sel = { selectedPid: hit.edge.conn.pid, selectedEdgeId: hit.edge.id };
      this.selection = sel;
      this.onSelect(sel);
      this.draw();
    } else {
      const sel = { selectedPid: null, selectedEdgeId: null };
      this.selection = sel;
      this.onSelect(sel);
      this.drag = { pan: true, sx, sy, tx: this.transform.x, ty: this.transform.y };
      this.canvas.setPointerCapture(ev.pointerId);
      this.canvas.classList.add("dragging");
      this.draw();
    }
  };

  private onPointerMove = (ev: PointerEvent) => {
    const rect = this.canvas.getBoundingClientRect();
    const sx = ev.clientX - rect.left;
    const sy = ev.clientY - rect.top;

    if (this.drag) {
      if ("pan" in this.drag && this.drag.pan) {
        this.transform.x = this.drag.tx + (sx - this.drag.sx);
        this.transform.y = this.drag.ty + (sy - this.drag.sy);
        this.draw();
        return;
      }
      if ("node" in this.drag) {
        const w = this.screenToWorld(sx, sy);
        this.drag.node.x = w.x;
        this.drag.node.y = w.y;
        this.drag.node.vx = 0;
        this.drag.node.vy = 0;
        this.drag.moved = true;
        this.draw();
      }
      return;
    }

    const hit = this.hitTest(sx, sy);
    this.hoverNode = hit?.type === "node" ? hit.node : null;
    this.hoverEdge = hit?.type === "edge" ? hit.edge : null;
    this.canvas.style.cursor = hit ? "pointer" : "grab";
    this.draw();
  };

  private onPointerUp = (ev: PointerEvent) => {
    if (this.drag && "node" in this.drag && !this.drag.moved && this.drag.node.kind === "proc") {
      const sel = { selectedPid: this.drag.node.pid, selectedEdgeId: null as string | null };
      this.selection = sel;
      this.onSelect(sel);
    }
    this.drag = null;
    this.canvas.classList.remove("dragging");
    try {
      this.canvas.releasePointerCapture(ev.pointerId);
    } catch {
      /* ignore */
    }
    this.draw();
  };

  private onWheel = (ev: WheelEvent) => {
    ev.preventDefault();
    const rect = this.canvas.getBoundingClientRect();
    const sx = ev.clientX - rect.left;
    const sy = ev.clientY - rect.top;
    const delta = -ev.deltaY * 0.0012;
    const k0 = this.transform.k;
    const k1 = Math.max(0.35, Math.min(2.5, k0 * (1 + delta)));
    const wx = (sx - this.transform.x) / k0;
    const wy = (sy - this.transform.y) / k0;
    this.transform.k = k1;
    this.transform.x = sx - wx * k1;
    this.transform.y = sy - wy * k1;
    this.draw();
  };
}

export function dirLabel(dir: DirUi): string {
  return dir === "local" ? "本地互联" : dir === "in" ? "入站" : "出站";
}

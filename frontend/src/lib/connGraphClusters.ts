import { graphEdgeRows, type UiConn } from "./connGraphEngine";

export type LocalComponent = {
  /** Stable id (min pid in component). */
  id: string;
  pids: number[];
  /** Deduped local interconnect rows with peerPid, both ends in this component. */
  localEdges: UiConn[];
};

const STORAGE_KEY = "procpane.connGraph.componentId";

class UnionFind {
  parent = new Map<number, number>();

  find(a: number): number {
    let p = this.parent.get(a) ?? a;
    while (p !== (this.parent.get(p) ?? p)) {
      p = this.parent.get(p) ?? p;
    }
    let cur = a;
    while (cur !== p) {
      const next = this.parent.get(cur) ?? cur;
      this.parent.set(cur, p);
      cur = next;
    }
    return p;
  }

  union(a: number, b: number) {
    const ra = this.find(a);
    const rb = this.find(b);
    if (ra !== rb) this.parent.set(rb, ra);
  }

  ensure(pid: number) {
    if (!this.parent.has(pid)) this.parent.set(pid, pid);
  }
}

function shortProcName(name: string, max = 14): string {
  if (name.length <= max) return name;
  return name.slice(0, max - 1) + "…";
}

/** Local purple edges only (paired). */
export function localInterconnectEdges(all: UiConn[]): UiConn[] {
  return graphEdgeRows(all).filter((c) => c.dir === "local" && c.peerPid);
}

/** Connected components over process nodes linked by local↔local pairs. Singletons for processes with no paired local edge. */
export function buildLocalComponents(all: UiConn[]): LocalComponent[] {
  const localEdges = localInterconnectEdges(all);
  const uf = new UnionFind();

  for (const c of localEdges) {
    uf.ensure(c.pid);
    uf.ensure(c.peerPid!);
    uf.union(c.pid, c.peerPid!);
  }

  const allPids = new Set<number>();
  for (const c of all) {
    allPids.add(c.pid);
    if (c.peerPid) allPids.add(c.peerPid);
  }

  for (const pid of allPids) uf.ensure(pid);

  const buckets = new Map<number, number[]>();
  for (const pid of allPids) {
    const root = uf.find(pid);
    const list = buckets.get(root) ?? [];
    list.push(pid);
    buckets.set(root, list);
  }

  const components: LocalComponent[] = [];
  for (const pids of buckets.values()) {
    pids.sort((a, b) => a - b);
    const id = String(pids[0]);
    const pidSet = new Set(pids);
    const edges = localEdges.filter(
      (c) => pidSet.has(c.pid) && c.peerPid != null && pidSet.has(c.peerPid),
    );
    components.push({ id, pids, localEdges: edges });
  }

  components.sort((a, b) => {
    if (b.pids.length !== a.pids.length) return b.pids.length - a.pids.length;
    return Number(a.id) - Number(b.id);
  });
  return components;
}

export function componentForPid(components: LocalComponent[], pid: number): LocalComponent | undefined {
  return components.find((c) => c.pids.includes(pid));
}

export function pickDefaultComponentId(
  components: LocalComponent[],
  remembered: string | null,
): string | null {
  if (components.length === 0) return null;
  if (remembered && components.some((c) => c.id === remembered)) return remembered;
  return components[0].id;
}

export function loadRememberedComponentId(): string | null {
  try {
    return sessionStorage.getItem(STORAGE_KEY);
  } catch {
    return null;
  }
}

export function rememberComponentId(id: string | null) {
  try {
    if (id == null) sessionStorage.removeItem(STORAGE_KEY);
    else sessionStorage.setItem(STORAGE_KEY, id);
  } catch {
    /* ignore */
  }
}

/** Label: `Cursor↔gopls (3)` or `Cursor 等 (5)`. */
export function componentLabel(comp: LocalComponent, names: Map<number, string>): string {
  const sorted = [...comp.pids].sort((a, b) => {
    const na = names.get(a) ?? "?";
    const nb = names.get(b) ?? "?";
    return na.localeCompare(nb) || a - b;
  });
  const labels = sorted.map((pid) => shortProcName(names.get(pid) ?? "?"));
  const n = comp.pids.length;
  if (n === 1) return `${labels[0]} (${n})`;
  if (n === 2) return `${labels[0]}↔${labels[1]} (${n})`;
  return `${labels[0]}↔${labels[1]} 等 (${n})`;
}

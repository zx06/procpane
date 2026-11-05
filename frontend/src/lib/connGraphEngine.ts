import type { Connection } from "./api";

export type DirUi = "local" | "in" | "out";

export type GraphFilters = {
  localOnly: boolean;
  inOnly: boolean;
  outOnly: boolean;
  search: string;
};

export type GraphSelection = {
  selectedPid: number | null;
  selectedEdgeId: string | null;
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
  /** Local loopback row without a matched peer process (graph skips; table may show). */
  unpairedLocal?: boolean;
};

export function dirFromModel(d: string): DirUi {
  if (d === "local") return "local";
  if (d === "inbound") return "in";
  return "out";
}

export function toUiConn(c: Connection): UiConn {
  const dir = dirFromModel(c.direction);
  const peerPid = c.peerPid > 0 ? c.peerPid : null;
  return {
    id: c.id,
    pid: c.pid,
    name: c.name,
    local: `${c.localHost}:${c.localPort}`,
    remote: `${c.remoteHost}:${c.remotePort}`,
    dir,
    peerPid,
    peerName: c.peerName || "",
    unpairedLocal: dir === "local" && peerPid == null,
  };
}

function canonicalLocalKey(c: UiConn): string {
  const a = `${c.local}`;
  const b = `${c.remote}`;
  const [x, y] = a < b ? [a, b] : [b, a];
  return `${x}|${y}|${Math.min(c.pid, c.peerPid ?? c.pid)}|${Math.max(c.pid, c.peerPid ?? c.pid)}`;
}

/** Filtered rows for table/sidebar (dedupes mirrored local interconnect rows). */
export function displayTableRows(
  all: UiConn[],
  filters: GraphFilters,
  namesByPid: Map<number, string>,
): UiConn[] {
  return graphEdgeRows(all.filter((c) => connVisible(c, filters, namesByPid)));
}

/** One row per undirected local edge for the graph (table uses the same dedupe). */
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

export function namesByPidFrom(conns: UiConn[]): Map<number, string> {
  const m = new Map<number, string>();
  for (const c of conns) {
    m.set(c.pid, c.name);
    if (c.peerPid) m.set(c.peerPid, c.peerName || m.get(c.peerPid) || "?");
  }
  return m;
}

export function connVisible(
  c: UiConn,
  filters: GraphFilters,
  namesByPid: Map<number, string>,
): boolean {
  const name = namesByPid.get(c.pid) ?? c.name;
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

export function dirLabel(dir: DirUi): string {
  return dir === "local" ? "本地互联" : dir === "in" ? "入站" : "出站";
}

export function dirCssVar(dir: DirUi): string {
  if (dir === "local") return "var(--local)";
  if (dir === "in") return "var(--in-edge)";
  return "var(--out-edge)";
}

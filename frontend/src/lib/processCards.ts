import type { ChainNode, Listener } from "./api";

export type PortRef = { port: number; proto: string };

export type ProcessCard = {
  id: string;
  pid: number;
  name: string;
  cpuPercent: number;
  memMb: number;
  cmdline: string;
  cwd: string;
  user: string;
  started: string;
  leftover?: boolean;
  chain: ChainNode[];
  ports: PortRef[];
};

function sortPorts(a: PortRef, b: PortRef) {
  return a.port - b.port || a.proto.localeCompare(b.proto);
}

/** Group listeners by PID into one card per process. */
export function listenersToProcessCards(listeners: Listener[]): ProcessCard[] {
  const byPid = new Map<number, Listener[]>();
  for (const l of listeners) {
    const list = byPid.get(l.pid);
    if (list) list.push(l);
    else byPid.set(l.pid, [l]);
  }

  const cards: ProcessCard[] = [];
  for (const [pid, items] of byPid) {
    const first = items[0];
    const ports = items
      .map((l) => ({ port: l.port, proto: l.proto }))
      .sort(sortPorts);
    cards.push({
      id: `pid-${pid}`,
      pid,
      name: first.name,
      cpuPercent: first.cpuPercent,
      memMb: first.memMb,
      cmdline: first.cmdline,
      cwd: first.cwd,
      user: first.user,
      started: first.started,
      leftover: items.some((l) => l.leftover),
      chain: first.chain,
      ports,
    });
  }

  cards.sort((a, b) => {
    const pa = a.ports[0]?.port ?? 0;
    const pb = b.ports[0]?.port ?? 0;
    return pa - pb || a.pid - b.pid;
  });
  return cards;
}

/**
 * Search filter: include a process if any of its listeners match port, name, or pid.
 * When matched, all listeners for that PID are kept (full port list on the card).
 */
export function filterListenersBySearch(
  rows: Listener[],
  query: string,
): Listener[] {
  const q = query.trim().toLowerCase();
  if (!q) return rows;

  const matchingPids = new Set<number>();
  for (const r of rows) {
    if (
      String(r.port).includes(q) ||
      r.name.toLowerCase().includes(q) ||
      String(r.pid).includes(q)
    ) {
      matchingPids.add(r.pid);
    }
  }
  return rows.filter((r) => matchingPids.has(r.pid));
}

export function formatPortsSnippet(ports: PortRef[]): string {
  return ports.map((p) => `:${p.port}/${p.proto}`).join(" · ");
}

import { Application, Call, Events } from "@wailsio/runtime";

const svc = "github.com/zx06/procpane/internal/gui.ProcpaneService";

export type ChainNode = { pid: number; name: string; note: string };

export type Listener = {
  id: string;
  port: number;
  proto: string;
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
};

export type ListResponse = {
  generatedAt: string;
  listeners: Listener[];
};

export type KillResult = {
  pid: number;
  name: string;
  signal: string;
  success: boolean;
  error?: string;
};

export type PortRef = { port: number; proto: string };

export type ProcessInfo = {
  pid: number;
  name: string;
  cmdline: string;
  cwd: string;
  user: string;
  started: string;
  cpuPercent: number;
  memMb: number;
  memBytes?: number;
  chain: ChainNode[];
  ports?: PortRef[];
};

export async function listListeners(proto: string): Promise<Listener[]> {
  return Call.ByName(`${svc}.ListListeners`, proto || "all");
}

export async function showProcess(pid: number): Promise<ProcessInfo> {
  return Call.ByName(`${svc}.ShowProcess`, pid);
}

export async function killProcess(pid: number, signal = "TERM"): Promise<KillResult> {
  return Call.ByName(`${svc}.KillProcess`, pid, signal);
}

export async function version(): Promise<string> {
  return Call.ByName(`${svc}.Version`);
}

export function onRefresh(handler: () => void): () => void {
  return Events.On("procpane:refresh", handler);
}

export function quitApp(): Promise<void> {
  return Application.Quit();
}

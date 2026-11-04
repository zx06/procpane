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
  source?: string;
  host?: string;
  containerId?: string;
  containerName?: string;
  image?: string;
  publishHost?: string;
  publishPort?: number;
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

export type SSHHostView = {
  alias: string;
  hostname: string;
  port?: string;
  enabled: boolean;
};

export type DockerContextView = {
  name: string;
  description?: string;
  host: string;
  skipTLSVerify?: boolean;
  current: boolean;
  selected: boolean;
  error?: string;
};

export type DockerStatus = {
  endpoint: string;
  reachable: boolean;
  error?: string;
  selectedContext?: string;
  currentContext?: string;
  contexts?: DockerContextView[];
};

export type SettingsView = {
  statePath: string;
  refreshIntervalSec: number;
  docker: DockerStatus;
  sshHosts: SSHHostView[];
};

export async function listListeners(proto: string): Promise<Listener[]> {
  return Call.ByName(`${svc}.ListListeners`, proto || "all");
}

export async function listDockerListeners(): Promise<Listener[]> {
  return Call.ByName(`${svc}.ListDockerListeners`);
}

export async function dockerStatus(): Promise<DockerStatus> {
  return Call.ByName(`${svc}.DockerStatus`);
}

export async function listSSHHosts(): Promise<SSHHostView[]> {
  return Call.ByName(`${svc}.ListSSHHosts`);
}

export async function listSSHListeners(alias: string): Promise<Listener[]> {
  return Call.ByName(`${svc}.ListSSHListeners`, alias);
}

export async function getSettings(): Promise<SettingsView> {
  return Call.ByName(`${svc}.GetSettings`);
}

export async function setSSHEnabled(alias: string, enabled: boolean): Promise<void> {
  return Call.ByName(`${svc}.SetSSHEnabled`, alias, enabled);
}

export async function setDockerContext(name: string): Promise<void> {
  return Call.ByName(`${svc}.SetDockerContext`, name);
}

export async function setRefreshIntervalSec(sec: number): Promise<void> {
  return Call.ByName(`${svc}.SetRefreshIntervalSec`, sec);
}

export async function openSettings(): Promise<void> {
  return Call.ByName(`${svc}.OpenSettings`);
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

export type SettingsChangedPayload = {
  refreshIntervalSec?: number;
};

export function onSettingsChanged(
  handler: (data: SettingsChangedPayload) => void,
): () => void {
  return Events.On("procpane:settings-changed", handler);
}

export function emitSettingsChanged(data: SettingsChangedPayload): void {
  Events.Emit("procpane:settings-changed", data);
}

export function quitApp(): Promise<void> {
  return Application.Quit();
}

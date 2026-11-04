/** Backend uses "?" when a field is unknown (see internal/model). */
export const UNKNOWN_FIELD = "?";

export const UNKNOWN_PROCESS_LABEL = "未知进程";

export function displayProcessName(name: string): string {
  if (!name || name === UNKNOWN_FIELD) {
    return UNKNOWN_PROCESS_LABEL;
  }
  return name;
}

/** Secondary hint for well-known ports when process owner is missing (not a fake process name). */
export function wellKnownPortHint(port: number, proto: string): string | null {
  const p = proto.toLowerCase();
  if (p === "udp") {
    switch (port) {
      case 68:
        return "bootps / dhcp";
      case 111:
        return "sunrpc";
      case 5353:
        return "mdns";
    }
  }
  if (p === "tcp") {
    switch (port) {
      case 111:
        return "sunrpc";
    }
  }
  return null;
}

export function listenerMatchesSearch(
  q: string,
  port: number,
  name: string,
  pid: number,
  extra?: string,
): boolean {
  const hay = [
    String(port),
    name,
    displayProcessName(name),
    String(pid),
    extra ?? "",
  ]
    .join("\n")
    .toLowerCase();
  return hay.includes(q);
}

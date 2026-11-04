import { getSettings, onSettingsChanged } from "./api";

export const DEFAULT_REFRESH_INTERVAL_SEC = 30;

export async function loadRefreshIntervalMs(): Promise<number> {
  const settings = await getSettings();
  const sec = settings.refreshIntervalSec || DEFAULT_REFRESH_INTERVAL_SEC;
  return sec * 1000;
}

/** Subscribe to refresh-interval changes (settings window or tray). */
export function watchRefreshInterval(
  onChange: (intervalMs: number) => void,
): () => void {
  return onSettingsChanged((data) => {
    const sec = data.refreshIntervalSec;
    if (sec && sec > 0) {
      onChange(sec * 1000);
    }
  });
}

/** Footer label: last successful refresh as local clock time. */
export function formatLastUpdated(ts: number | null | undefined): string {
  if (ts == null || ts <= 0) return "上次更新 —";
  const d = new Date(ts);
  const h = d.getHours().toString().padStart(2, "0");
  const m = d.getMinutes().toString().padStart(2, "0");
  const s = d.getSeconds().toString().padStart(2, "0");
  return `上次更新 ${h}:${m}:${s}`;
}

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

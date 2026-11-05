/** Auto-refresh while the tray panel is visible and this tab is active. */
export function shouldPoll(active: boolean): boolean {
  return active && !document.hidden;
}

export function bindPanelVisibility(
  active: () => boolean,
  onShow: () => void,
  onHide: () => void,
): () => void {
  const onVis = () => {
    if (shouldPoll(active())) onShow();
    else onHide();
  };
  document.addEventListener("visibilitychange", onVis);
  return () => document.removeEventListener("visibilitychange", onVis);
}

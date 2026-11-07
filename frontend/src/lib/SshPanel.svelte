<script lang="ts">
  import { onMount } from "svelte";
  import {
    listSSHHosts,
    listSSHListeners,
    openSettings,
    onRefresh,
    onSettingsChanged,
    refreshTargetsPanel,
    type Listener,
    type RefreshPayload,
    type SSHHostView,
  } from "./api";
  import { displayProcessName, wellKnownPortHint } from "./displayListener";
  import {
    filterSSHCardsBySearch,
    listenersToSSHProcessCards,
    type SSHProcessCard,
  } from "./processCards";
  import {
    formatLastUpdated,
    loadRefreshIntervalMs,
    settingsAffectsSource,
    watchRefreshInterval,
  } from "./refreshPoll";
  import { bindPanelVisibility, shouldPoll } from "./panelPoll";
  import RefreshButton from "./RefreshButton.svelte";

  let { active = false }: { active?: boolean } = $props();

  type HostCache = {
    rows: Listener[];
    lastSuccessAt: number | null;
    error: string;
  };

  let hosts: SSHHostView[] = $state([]);
  let selected = $state<string | null>(null);
  let rows: Listener[] = $state([]);
  let query = $state("");
  let refreshing = $state(false);
  let error = $state("");
  let lastSuccessAt = $state<number | null>(null);

  const hostCache = new Map<string, HostCache>();
  let refreshGen = 0;

  const enabled = $derived(hosts.filter((h) => h.enabled));

  const cards = $derived(
    filterSSHCardsBySearch(listenersToSSHProcessCards(rows), query),
  );

  function freshLabel() {
    return formatLastUpdated(lastSuccessAt);
  }

  function restoreHostCache(alias: string) {
    const cached = hostCache.get(alias);
    if (cached) {
      rows = cached.rows;
      lastSuccessAt = cached.lastSuccessAt;
      error = cached.error;
    } else {
      rows = [];
      lastSuccessAt = null;
      error = "";
    }
  }

  function saveHostCache(alias: string) {
    hostCache.set(alias, {
      rows: [...rows],
      lastSuccessAt,
      error,
    });
  }

  async function loadHosts() {
    hosts = await listSSHHosts();
    const en = hosts.filter((h) => h.enabled);
    if (en.length > 0 && !selected) {
      selected = en[0].alias;
    }
    if (selected && !en.some((h) => h.alias === selected)) {
      selected = en[0]?.alias ?? null;
    }
  }

  async function doRefresh() {
    const gen = ++refreshGen;
    const host = selected;
    refreshing = true;
    try {
      await loadHosts();
      if (gen !== refreshGen || host !== selected) return;
      if (!selected) {
        rows = [];
        error = "";
        return;
      }
      const data = await listSSHListeners(selected);
      if (gen !== refreshGen || host !== selected) return;
      rows = data;
      lastSuccessAt = Date.now();
      error = "";
      saveHostCache(selected);
    } catch (e) {
      console.error(e);
      if (gen !== refreshGen || host !== selected) return;
      error = e instanceof Error ? e.message : "SSH 列表失败";
      saveHostCache(selected ?? "");
    } finally {
      if (gen === refreshGen) refreshing = false;
    }
  }

  async function pickHost(alias: string) {
    if (alias === selected) return;
    refreshGen++;
    selected = alias;
    restoreHostCache(alias);
    await doRefresh();
  }

  let pollMs = $state(30_000);
  let pollTimer: ReturnType<typeof setInterval> | undefined;
  let unsubSettings: (() => void) | undefined;
  let unsubRefresh: (() => void) | undefined;

  function stopPoll() {
    if (pollTimer !== undefined) {
      clearInterval(pollTimer);
      pollTimer = undefined;
    }
  }

  function startPoll() {
    if (pollTimer !== undefined) return;
    if (!shouldPoll(active)) return;
    pollTimer = setInterval(() => {
      if (!refreshing && selected && shouldPoll(active)) void doRefresh();
    }, pollMs);
  }

  function restartPoll() {
    stopPoll();
    if (shouldPoll(active)) startPoll();
  }

  onMount(async () => {
    pollMs = await loadRefreshIntervalMs();
    restoreHostCache(selected ?? "");
    void doRefresh();
    restartPoll();
    unsubRefresh = onRefresh((data: RefreshPayload) => {
      if (refreshTargetsPanel(data, "ssh")) void doRefresh();
    });
    unsubSettings = watchRefreshInterval((ms) => {
      pollMs = ms;
      restartPoll();
    });
    const unsubSettingsChanged = onSettingsChanged((data) => {
      if (settingsAffectsSource(data, "ssh")) {
        void loadHosts().then(() => void doRefresh());
        restartPoll();
      }
    });
    const unbindVis = bindPanelVisibility(
      () => active,
      () => restartPoll(),
      () => stopPoll(),
    );
    return () => {
      stopPoll();
      unbindVis();
      unsubSettings?.();
      unsubRefresh?.();
      unsubSettingsChanged();
    };
  });

  $effect(() => {
    if (shouldPoll(active)) restartPoll();
    else stopPoll();
  });

  function cardHint(card: SSHProcessCard) {
    if (card.pid > 0) return "";
    const p = card.ports[0];
    if (!p) return "";
    return wellKnownPortHint(p.port, p.proto) ?? "";
  }
</script>

<div class="tab-panel">
  <header class="pop-hd">
    <div class="hd-text">
      <h1>SSH</h1>
      <div class="sub">远程主机 LISTEN（只读，无法结束进程）</div>
    </div>
    <span class="count-pill">{cards.length}</span>
  </header>

  <div class="host-chips">
    {#each enabled as h (h.alias)}
      <button
        type="button"
        class="host-chip"
        class:on={selected === h.alias}
        onclick={() => pickHost(h.alias)}
      >
        {h.alias}
      </button>
    {/each}
  </div>

  <div class="tools">
    <div class="search m3-ctrl">
      <input type="search" placeholder="端口 / 进程…" bind:value={query} />
    </div>
    <RefreshButton spinning={refreshing} onclick={() => doRefresh()} />
  </div>

  <div class="cards-scroll">
    {#if enabled.length === 0}
      <div class="empty">
        <p>尚未启用任何 SSH 主机</p>
        <button type="button" class="btn primary" onclick={() => openSettings()}>
          去设置启用 SSH 主机
        </button>
      </div>
    {:else if cards.length > 0}
      {#if error}
        <div class="hint-bar warn">
          {error}
          <button type="button" class="link-btn" onclick={() => doRefresh()}>再试一次</button>
        </div>
      {/if}
      <div class="cards">
        {#each cards as card (card.id)}
          <div class="pcard m3-card pcard-readonly">
            <div class="pcard-top">
              <div class="pcard-ports">
                {#each card.ports as p (p.port + p.proto)}
                  <span class="port-chip">
                    <span class="port">:{p.port}</span>
                    <span class="badge {p.proto}">{p.proto}</span>
                  </span>
                {/each}
              </div>
              <div
                class="name"
                class:unknown={card.name === "?" && card.pid <= 0}
              >
                {displayProcessName(card.name)}
              </div>
            </div>
            <div class="pcard-mid">
              {#if card.pid > 0}
                <span>PID {card.pid}</span>
              {/if}
              {#if cardHint(card)}
                <span class="dim">{cardHint(card)}</span>
              {/if}
              <span class="dim">{card.host}</span>
            </div>
          </div>
        {/each}
      </div>
    {:else if error && rows.length === 0}
      <div class="empty">
        <p>{error}</p>
        <button type="button" class="btn primary" onclick={() => doRefresh()}>再试一次</button>
      </div>
    {:else if error}
      <div class="hint-bar warn">
        {error}
        <button type="button" class="link-btn" onclick={() => doRefresh()}>再试一次</button>
      </div>
    {:else if refreshing && rows.length === 0}
      <div class="empty">加载中…</div>
    {:else if query.trim() && rows.length > 0}
      <div class="empty">没有匹配「{query}」的结果</div>
    {:else}
      <div class="empty">没有 LISTEN 端口</div>
    {/if}
  </div>

  <footer class="foot">
    <span class="live">
      <span class="dot" class:busy={refreshing}></span>
      <span>{freshLabel()}</span>
    </span>
  </footer>
</div>

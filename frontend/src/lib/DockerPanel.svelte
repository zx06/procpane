<script lang="ts">
  import { onMount } from "svelte";
  import {
    dockerStatus,
    listDockerListeners,
    onRefresh,
    onSettingsChanged,
    refreshTargetsPanel,
    type Listener,
    type RefreshPayload,
  } from "./api";
  import { displayProcessName } from "./displayListener";
  import {
    filterDockerCardsBySearch,
    listenersToDockerCards,
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

  let rows: Listener[] = $state([]);
  let query = $state("");
  let refreshing = $state(false);
  let lastSuccessAt = $state<number | null>(null);
  let status = $state<Awaited<ReturnType<typeof dockerStatus>> | null>(null);
  let error = $state("");
  let refreshWarn = $state("");

  const cards = $derived(
    filterDockerCardsBySearch(listenersToDockerCards(rows), query),
  );

  function freshLabel() {
    return formatLastUpdated(lastSuccessAt);
  }

  async function doRefresh() {
    refreshing = true;
    refreshWarn = "";
    try {
      status = await dockerStatus();
      if (!status.reachable) {
        const msg = status.error || "无法连接 Docker Engine";
        if (rows.length === 0) {
          error = msg;
        } else {
          refreshWarn = msg;
        }
        return;
      }
      error = "";
      rows = await listDockerListeners();
      lastSuccessAt = Date.now();
    } catch (e) {
      console.error(e);
      const msg =
        e instanceof Error ? e.message : "加载 Docker 端口列表失败";
      if (rows.length === 0) {
        error = msg;
      } else {
        refreshWarn = msg;
      }
    } finally {
      refreshing = false;
    }
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
      if (!refreshing && shouldPoll(active)) void doRefresh();
    }, pollMs);
  }

  function restartPoll() {
    stopPoll();
    if (shouldPoll(active)) startPoll();
  }

  onMount(async () => {
    pollMs = await loadRefreshIntervalMs();
    void doRefresh();
    restartPoll();
    unsubRefresh = onRefresh((data: RefreshPayload) => {
      if (refreshTargetsPanel(data, "docker")) void doRefresh();
    });
    unsubSettings = watchRefreshInterval((ms) => {
      pollMs = ms;
      restartPoll();
    });
    const unsubSettingsChanged = onSettingsChanged((data) => {
      if (settingsAffectsSource(data, "docker")) void doRefresh();
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
</script>

<div class="tab-panel">
  <header class="pop-hd">
    <div class="hd-text">
      <h1>Docker</h1>
      <div class="sub">已发布端口（只读，无法结束容器进程）</div>
    </div>
    <span class="count-pill">{cards.length}</span>
  </header>

  {#if status && !status.reachable}
    <div class="hint-bar warn">Docker 未连接 · {status.endpoint}</div>
  {:else if status?.reachable}
    <div class="hint-bar ok">Engine · {status.endpoint}</div>
  {/if}

  {#if refreshWarn}
    <div class="hint-bar warn">
      {refreshWarn}
      <button type="button" class="link-btn" onclick={() => doRefresh()}>再试一次</button>
    </div>
  {/if}

  <div class="tools">
    <div class="search">
      <input type="search" placeholder="容器 / 端口 / 镜像…" bind:value={query} />
    </div>
    <RefreshButton spinning={refreshing} onclick={() => doRefresh()} />
  </div>

  <div class="cards-scroll">
    {#if cards.length > 0}
      <div class="cards">
        {#each cards as card (card.id)}
          <div class="pcard docker-card pcard-readonly">
            <div class="pcard-top">
              <div class="pcard-ports">
                {#each card.ports as p (p.port + p.proto)}
                  <span class="port-chip">
                    <span class="port">:{p.port}</span>
                    <span class="badge {p.proto}">{p.proto}</span>
                  </span>
                {/each}
              </div>
              <div class="name">
                {card.containerName || displayProcessName("?")}
              </div>
            </div>
            <div class="pcard-mid docker-meta">
              <span class="dim">{card.containerId}</span>
              {#if card.image}
                <span class="dim">{card.image}</span>
              {/if}
              {#if card.publishHost}
                <span class="dim">{card.publishHost}</span>
              {/if}
            </div>
          </div>
        {/each}
      </div>
    {:else if error}
      <div class="empty">
        <p>{error}</p>
        <p class="empty-hint">请启动 Docker 或检查 DOCKER_HOST / 本地 socket。</p>
        <button type="button" class="btn primary" onclick={() => doRefresh()}>再试一次</button>
      </div>
    {:else if refreshing}
      <div class="empty">加载中…</div>
    {:else if query.trim() && rows.length > 0}
      <div class="empty">没有匹配「{query}」的结果</div>
    {:else}
      <div class="empty">没有已发布的容器端口</div>
    {/if}
  </div>

  <footer class="foot">
    <span class="live">
      <span class="dot" class:busy={refreshing}></span>
      <span>{freshLabel()}</span>
    </span>
  </footer>
</div>

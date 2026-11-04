<script lang="ts">
  import { onMount, onDestroy } from "svelte";
  import { dockerStatus, listDockerListeners, type Listener } from "./api";
  import { displayProcessName, listenerMatchesSearch } from "./displayListener";
  import {
    formatLastUpdated,
    loadRefreshIntervalMs,
    watchRefreshInterval,
  } from "./refreshPoll";

  let rows: Listener[] = $state([]);
  let query = $state("");
  let refreshing = $state(false);
  let lastSuccessAt = $state<number | null>(null);
  let status = $state<Awaited<ReturnType<typeof dockerStatus>> | null>(null);
  let error = $state("");
  /** Non-fatal refresh error when stale rows are still shown. */
  let refreshWarn = $state("");

  const filtered = $derived(
    rows.filter((r) => {
      const q = query.trim().toLowerCase();
      if (!q) return true;
      return (
        listenerMatchesSearch(q, r.port, r.name, r.pid) ||
        (r.containerName ?? "").toLowerCase().includes(q) ||
        (r.image ?? "").toLowerCase().includes(q)
      );
    }),
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
      const msg = "加载失败（见控制台）";
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

  function restartPollTimer() {
    if (pollTimer !== undefined) {
      clearInterval(pollTimer);
      pollTimer = undefined;
    }
    pollTimer = setInterval(() => {
      if (!refreshing) void doRefresh();
    }, pollMs);
  }

  onMount(async () => {
    pollMs = await loadRefreshIntervalMs();
    void doRefresh();
    restartPollTimer();
    unsubSettings = watchRefreshInterval((ms) => {
      pollMs = ms;
      restartPollTimer();
    });
  });

  onDestroy(() => {
    if (pollTimer) clearInterval(pollTimer);
    unsubSettings?.();
  });
</script>

<div class="tab-panel">
  <header class="pop-hd">
    <div class="hd-text">
      <h1>Docker</h1>
      <div class="sub">已发布端口（Engine API）</div>
    </div>
    <span class="count-pill">{filtered.length}</span>
  </header>

  {#if status && !status.reachable}
    <div class="hint-bar warn">Docker 未连接 · {status.endpoint}</div>
  {:else if status?.reachable}
    <div class="hint-bar ok">Engine · {status.endpoint}</div>
  {/if}

  {#if refreshWarn}
    <div class="hint-bar warn">{refreshWarn}</div>
  {/if}

  <div class="tools">
    <div class="search">
      <input type="search" placeholder="容器 / 端口 / 镜像…" bind:value={query} />
    </div>
    <button
      type="button"
      class="icon-btn"
      class:spin={refreshing}
      title="刷新"
      aria-label="刷新"
      onclick={() => doRefresh()}
    >
      ↻
    </button>
  </div>

  <div class="cards-scroll">
    {#if filtered.length > 0}
      <div class="cards">
        {#each filtered as row (row.id)}
          <div class="pcard docker-card">
            <div class="pcard-top">
              <div class="pcard-ports">
                <span class="port-chip">
                  <span class="port">:{row.port}</span>
                  <span class="badge {row.proto}">{row.proto}</span>
                </span>
              </div>
              <div class="name">
                {row.containerName || displayProcessName(row.name)}
              </div>
            </div>
            <div class="pcard-mid docker-meta">
              <span class="dim">{row.containerId}</span>
              {#if row.image}
                <span class="dim">{row.image}</span>
              {/if}
              {#if row.publishHost}
                <span class="dim">{row.publishHost}</span>
              {/if}
            </div>
          </div>
        {/each}
      </div>
    {:else if error}
      <div class="empty">
        <p>{error}</p>
        <p class="empty-hint">请启动 Docker 或检查 DOCKER_HOST / 本地 socket。</p>
      </div>
    {:else if refreshing}
      <div class="empty">加载中…</div>
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

<script lang="ts">
  import { onMount, onDestroy } from "svelte";
  import {
    listSSHHosts,
    listSSHListeners,
    openSettings,
    type Listener,
    type SSHHostView,
  } from "./api";
  import {
    displayProcessName,
    listenerMatchesSearch,
    wellKnownPortHint,
  } from "./displayListener";
  import {
    formatLastUpdated,
    loadRefreshIntervalMs,
    watchRefreshInterval,
  } from "./refreshPoll";

  let hosts: SSHHostView[] = $state([]);
  let selected = $state<string | null>(null);
  let rows: Listener[] = $state([]);
  let query = $state("");
  let refreshing = $state(false);
  let error = $state("");
  let lastSuccessAt = $state<number | null>(null);

  const enabled = $derived(hosts.filter((h) => h.enabled));

  const filtered = $derived(
    rows.filter((r) => {
      const q = query.trim().toLowerCase();
      if (!q) return true;
      return (
        listenerMatchesSearch(q, r.port, r.name, r.pid, r.host ?? "") ||
        (wellKnownPortHint(r.port, r.proto) ?? "").toLowerCase().includes(q)
      );
    }),
  );

  function freshLabel() {
    return formatLastUpdated(lastSuccessAt);
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
    refreshing = true;
    try {
      await loadHosts();
      if (!selected) {
        rows = [];
        error = "";
        return;
      }
      rows = await listSSHListeners(selected);
      lastSuccessAt = Date.now();
      error = "";
    } catch (e) {
      console.error(e);
      error = e instanceof Error ? e.message : "SSH 列表失败";
    } finally {
      refreshing = false;
    }
  }

  async function pickHost(alias: string) {
    if (alias !== selected) {
      rows = [];
      error = "";
      lastSuccessAt = null;
    }
    selected = alias;
    await doRefresh();
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
      if (!refreshing && selected) void doRefresh();
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
      <h1>SSH</h1>
      <div class="sub">已启用主机的 LISTEN</div>
    </div>
    <span class="count-pill">{filtered.length}</span>
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
    <div class="search">
      <input type="search" placeholder="端口 / 进程…" bind:value={query} />
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
    {#if enabled.length === 0}
      <div class="empty">
        <p>尚未启用任何 SSH 主机</p>
        <button type="button" class="btn primary" onclick={() => openSettings()}>
          去设置启用 SSH 主机
        </button>
      </div>
    {:else if filtered.length > 0}
      {#if error}
        <div class="hint-bar warn">{error}</div>
      {/if}
      <div class="cards">
        {#each filtered as row (row.id)}
          <div class="pcard">
            <div class="pcard-top">
              <div class="pcard-ports">
                <span class="port-chip">
                  <span class="port">:{row.port}</span>
                  <span class="badge {row.proto}">{row.proto}</span>
                </span>
              </div>
              <div
                class="name"
                class:unknown={row.name === "?" && row.pid <= 0}
              >
                {displayProcessName(row.name)}
              </div>
            </div>
            <div class="pcard-mid">
              {#if row.pid > 0}
                <span>PID {row.pid}</span>
              {/if}
              {#if row.pid <= 0 && row.name === "?"}
                {@const hint = wellKnownPortHint(row.port, row.proto)}
                {#if hint}
                  <span class="dim">{hint}</span>
                {/if}
              {/if}
              <span class="dim">{row.host}</span>
            </div>
          </div>
        {/each}
      </div>
    {:else if error}
      <div class="empty">{error}</div>
    {:else if refreshing && rows.length === 0}
      <div class="empty">加载中…</div>
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

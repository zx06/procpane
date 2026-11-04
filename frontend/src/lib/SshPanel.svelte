<script lang="ts">
  import { onMount, onDestroy } from "svelte";
  import {
    listSSHHosts,
    listSSHListeners,
    openSettings,
    type Listener,
    type SSHHostView,
  } from "./api";

  let hosts: SSHHostView[] = $state([]);
  let selected = $state<string | null>(null);
  let rows: Listener[] = $state([]);
  let query = $state("");
  let refreshing = $state(false);
  let error = $state("");
  let lastRefresh = $state(Date.now());
  let labelTick = $state(0);

  const enabled = $derived(hosts.filter((h) => h.enabled));

  const filtered = $derived(
    rows.filter((r) => {
      const q = query.trim().toLowerCase();
      if (!q) return true;
      return (
        String(r.port).includes(q) ||
        r.name.toLowerCase().includes(q) ||
        (r.host ?? "").toLowerCase().includes(q)
      );
    }),
  );

  function freshLabel() {
    void labelTick;
    const sec = (Date.now() - lastRefresh) / 1000;
    return refreshing ? "刷新中…" : `已刷新 · ${sec.toFixed(1)}s 前`;
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
    error = "";
    try {
      await loadHosts();
      if (!selected) {
        rows = [];
        return;
      }
      rows = await listSSHListeners(selected);
      lastRefresh = Date.now();
    } catch (e) {
      console.error(e);
      error = e instanceof Error ? e.message : "SSH 列表失败";
      rows = [];
    } finally {
      refreshing = false;
    }
  }

  async function pickHost(alias: string) {
    selected = alias;
    await doRefresh();
  }

  const POLL_MS = 5000;
  let pollTimer: ReturnType<typeof setInterval> | undefined;
  let labelTimer: ReturnType<typeof setInterval> | undefined;

  onMount(() => {
    void doRefresh();
    pollTimer = setInterval(() => {
      if (!refreshing && selected) void doRefresh();
    }, POLL_MS);
    labelTimer = setInterval(() => labelTick++, 1000);
  });

  onDestroy(() => {
    if (pollTimer) clearInterval(pollTimer);
    if (labelTimer) clearInterval(labelTimer);
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
    {:else if error}
      <div class="empty">{error}</div>
    {:else if filtered.length === 0}
      <div class="empty">没有 LISTEN 端口</div>
    {:else}
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
              <div class="name">{row.name}</div>
            </div>
            <div class="pcard-mid">
              {#if row.pid > 0}
                <span>PID {row.pid}</span>
              {/if}
              <span class="dim">{row.host}</span>
            </div>
          </div>
        {/each}
      </div>
    {/if}
  </div>

  <footer class="foot">
    <span class="live">
      <span class="dot" class:busy={refreshing}></span>
      <span>{freshLabel()}</span>
    </span>
  </footer>
</div>

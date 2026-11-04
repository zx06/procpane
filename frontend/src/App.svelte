<script lang="ts">
  import { onMount, onDestroy } from "svelte";
  import {
    listListeners,
    killProcess,
    onRefresh,
    quitApp,
    type Listener,
  } from "./lib/api";

  let rows: Listener[] = $state([]);
  let selectedId: string | null = $state(null);
  let protoFilter = $state("all");
  let query = $state("");
  let lastRefresh = $state(Date.now());
  let tick = $state(0);
  let refreshing = $state(false);
  let killOpen = $state(false);
  let toast = $state("");
  let toastVisible = $state(false);
  let focused = $state(true);

  const selected = $derived(rows.find((r) => r.id === selectedId) ?? null);

  const filtered = $derived(
    rows.filter((r) => {
      if (protoFilter !== "all" && r.proto !== protoFilter) return false;
      const q = query.trim().toLowerCase();
      if (!q) return true;
      return (
        String(r.port).includes(q) ||
        r.name.toLowerCase().includes(q) ||
        String(r.pid).includes(q)
      );
    }),
  );

  function fmtMem(mb: number) {
    if (mb <= 0) return "?";
    return mb >= 100 ? `${mb.toFixed(0)} MB` : `${mb.toFixed(1)} MB`;
  }

  function cpuClass(c: number) {
    if (c >= 10) return "hot";
    if (c >= 5) return "warn";
    return "";
  }

  function cpuBarWidth(c: number) {
    return Math.min(100, Math.round(c * 6));
  }

  function freshLabel() {
    void tick;
    if (refreshing) return "更新中…";
    const sec = (Date.now() - lastRefresh) / 1000;
    return sec < 3 ? "刚刚更新" : `${Math.floor(sec)}s 前更新`;
  }

  async function doRefresh() {
    refreshing = true;
    try {
      rows = await listListeners(protoFilter === "all" ? "all" : protoFilter);
      lastRefresh = Date.now();
      if (selectedId && !rows.find((r) => r.id === selectedId)) {
        selectedId = null;
      }
    } catch (e) {
      console.error(e);
      showToast("刷新失败（见控制台）");
    } finally {
      refreshing = false;
    }
  }

  function showToast(msg: string) {
    toast = msg;
    toastVisible = true;
    setTimeout(() => (toastVisible = false), 2200);
  }

  function toggleCard(id: string) {
    selectedId = selectedId === id ? null : id;
  }

  async function confirmKill() {
    if (!selected) return;
    const r = selected;
    const res = await killProcess(r.pid, "TERM");
    killOpen = false;
    if (!res.success) {
      showToast(res.error || "结束进程失败");
      return;
    }
    showToast(`已结束 ${r.name} (PID ${r.pid}) · 列表已刷新`);
    await doRefresh();
    selectedId = null;
  }

  async function copyValue(e: MouseEvent, label: string, text: string) {
    e.stopPropagation();
    try {
      await navigator.clipboard.writeText(text);
      showToast(`已复制 ${label}`);
    } catch {
      showToast("复制失败");
    }
  }

  function openKill(e: MouseEvent, r: Listener) {
    e.stopPropagation();
    selectedId = r.id;
    killOpen = true;
  }

  async function handleQuit() {
    try {
      await quitApp();
    } catch (e) {
      console.error(e);
      showToast("退出失败");
    }
  }

  let tickInterval: ReturnType<typeof setInterval>;
  let refreshInterval: ReturnType<typeof setInterval>;
  let unsubRefresh: (() => void) | undefined;

  onMount(() => {
    doRefresh();
    unsubRefresh = onRefresh(() => doRefresh());
    window.addEventListener("focus", () => (focused = true));
    window.addEventListener("blur", () => (focused = false));
    tickInterval = setInterval(() => {
      tick = Date.now();
    }, 1000);
    refreshInterval = setInterval(() => {
      const interval = focused ? 2000 : 5000;
      if (Date.now() - lastRefresh >= interval && !refreshing) {
        doRefresh();
      }
    }, 200);
  });

  onDestroy(() => {
    clearInterval(tickInterval);
    clearInterval(refreshInterval);
    unsubRefresh?.();
  });
</script>

<div class="app">
  <header class="pop-hd">
    <div class="mark" aria-hidden="true">
      <svg viewBox="0 0 18 18" fill="none">
        <rect x="2.5" y="2.5" width="13" height="13" rx="3" stroke="currentColor" stroke-width="1.5" />
        <path
          d="M5.5 6.5h7M5.5 9h7M5.5 11.5h7"
          stroke="currentColor"
          stroke-width="1.4"
          stroke-linecap="round"
        />
      </svg>
    </div>
    <div class="hd-text">
      <h1>procpane</h1>
      <div class="sub">本地 LISTEN 端口</div>
    </div>
    <span class="count-pill">{filtered.length}</span>
  </header>

  <div class="tools">
    <div class="search">
      <svg viewBox="0 0 16 16" fill="none" aria-hidden="true">
        <circle cx="7" cy="7" r="4.5" stroke="currentColor" stroke-width="1.4" />
        <path
          d="M10.5 10.5 13.5 13.5"
          stroke="currentColor"
          stroke-width="1.4"
          stroke-linecap="round"
        />
      </svg>
      <input type="search" placeholder="端口 / 进程…" bind:value={query} />
    </div>
    <div class="mini-seg">
      {#each ["all", "tcp", "udp"] as p}
        <button
          type="button"
          class:on={protoFilter === p}
          onclick={() => {
            protoFilter = p;
            doRefresh();
          }}
        >
          {p === "all" ? "全部" : p.toUpperCase()}
        </button>
      {/each}
    </div>
    <button
      type="button"
      class="icon-btn"
      class:spin={refreshing}
      title="刷新"
      aria-label="刷新"
      onclick={() => doRefresh()}
    >
      <svg viewBox="0 0 16 16" fill="none" aria-hidden="true">
        <path
          d="M3.2 8a4.8 4.8 0 0 1 8.3-3.2M12.8 8a4.8 4.8 0 0 1-8.3 3.2"
          stroke="currentColor"
          stroke-width="1.4"
          stroke-linecap="round"
        />
        <path
          d="M11.2 2.8v2.6H8.6M4.8 13.2v-2.6h2.6"
          stroke="currentColor"
          stroke-width="1.4"
          stroke-linecap="round"
          stroke-linejoin="round"
        />
      </svg>
    </button>
  </div>

  <div class="cards-scroll">
    {#if filtered.length === 0}
      <div class="empty">没有匹配的 LISTEN 端口</div>
    {:else}
      <div class="cards">
        {#each filtered as r (r.id)}
          <!-- svelte-ignore a11y_click_events_have_key_events -->
          <!-- svelte-ignore a11y_no_static_element_interactions -->
          <div
            class="pcard"
            class:on={r.id === selectedId}
            onclick={() => toggleCard(r.id)}
          >
            <div class="pcard-top">
              <div class="port">:{r.port}</div>
              <div class="name">
                {r.name}
                {#if r.leftover}
                  <span class="tag">leftover</span>
                {/if}
              </div>
              <span class="badge {r.proto}">{r.proto}</span>
            </div>
            <div class="pcard-mid">
              <span>PID {r.pid}</span>
              <div class="bar" class:hot={cpuClass(r.cpuPercent) === "hot"}>
                <i style="width: {cpuBarWidth(r.cpuPercent)}%"></i>
              </div>
              <span>
                <span class={cpuClass(r.cpuPercent)}>{r.cpuPercent.toFixed(1)}%</span>
                · {fmtMem(r.memMb)}
              </span>
            </div>
            {#if r.id === selectedId}
              <div class="pcard-detail">
                <div class="kv">
                  <div class="r">
                    <div class="k">cmdline</div>
                    <!-- svelte-ignore a11y_click_events_have_key_events -->
                    <!-- svelte-ignore a11y_no_static_element_interactions -->
                    <div
                      class="v copyable"
                      title="点击复制"
                      onclick={(e) => copyValue(e, "cmdline", r.cmdline)}
                    >
                      {r.cmdline}
                    </div>
                  </div>
                  <div class="r">
                    <div class="k">cwd</div>
                    <!-- svelte-ignore a11y_click_events_have_key_events -->
                    <!-- svelte-ignore a11y_no_static_element_interactions -->
                    <div
                      class="v copyable"
                      class:q={r.cwd === "?"}
                      title="点击复制"
                      onclick={(e) => copyValue(e, "cwd", r.cwd)}
                    >
                      {r.cwd}
                    </div>
                  </div>
                  <div class="r">
                    <div class="k">user</div>
                    <!-- svelte-ignore a11y_click_events_have_key_events -->
                    <!-- svelte-ignore a11y_no_static_element_interactions -->
                    <div
                      class="v copyable"
                      title="点击复制"
                      onclick={(e) =>
                        copyValue(e, "user", `${r.user} · ${r.started}`)}
                    >
                      {r.user} · {r.started}
                    </div>
                  </div>
                </div>
                {#if r.chain.length > 0}
                  <div class="chain">
                    {#each r.chain as node, i}
                      {#if i > 0}
                        <span class="arrow">›</span>
                      {/if}
                      <span class="n">{node.name} · {node.pid}</span>
                    {/each}
                  </div>
                {/if}
                <div class="acts">
                  <button type="button" class="btn danger" onclick={(e) => openKill(e, r)}>
                    结束进程
                  </button>
                </div>
              </div>
            {/if}
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
    <button type="button" class="quit-btn" onclick={handleQuit}>退出</button>
  </footer>
</div>

{#if killOpen && selected}
  <!-- svelte-ignore a11y_click_events_have_key_events -->
  <div class="overlay" role="presentation" onclick={() => (killOpen = false)}>
    <!-- svelte-ignore a11y_no_static_element_interactions -->
    <div class="sheet" onclick={(e) => e.stopPropagation()}>
      <h3>结束进程？</h3>
      <p>向目标发送 SIGTERM，结束后列表会刷新。</p>
      <div class="snippet">
        PID {selected.pid} {selected.name} · :{selected.port}/{selected.proto}
      </div>
      <div class="rowbtns">
        <button type="button" class="btn" onclick={() => (killOpen = false)}>取消</button>
        <button type="button" class="btn danger" onclick={confirmKill}>结束进程</button>
      </div>
    </div>
  </div>
{/if}

{#if toastVisible}
  <div class="toast">{toast}</div>
{/if}

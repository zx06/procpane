<script lang="ts">
  import { onMount, onDestroy } from "svelte";
  import {
    listListeners,
    killProcess,
    onRefresh,
    quitApp,
    type Listener,
  } from "./lib/api";
  import {
    filterListenersBySearch,
    formatPortsSnippet,
    listenersToProcessCards,
    type ProcessCard,
  } from "./lib/processCards";

  let rows: Listener[] = $state([]);
  let selectedId: string | null = $state(null);
  let protoFilter = $state("all");
  let query = $state("");
  let lastRefresh = $state(Date.now());
  let refreshing = $state(false);
  let killOpen = $state(false);
  let toast = $state("");
  let toastVisible = $state(false);
  /** Bumps every ~1s while the panel is visible so the freshness label updates. */
  let labelTick = $state(0);

  const processCards = $derived(
    listenersToProcessCards(filterListenersBySearch(rows, query)),
  );

  const selected = $derived(
    processCards.find((c) => c.id === selectedId) ?? null,
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
    void labelTick;
    const sec = (Date.now() - lastRefresh) / 1000;
    return refreshing ? "刷新中…" : `已刷新 · ${sec.toFixed(1)}s 前`;
  }

  async function doRefresh() {
    refreshing = true;
    try {
      rows = await listListeners(protoFilter === "all" ? "all" : protoFilter);
      lastRefresh = Date.now();
      if (
        selectedId &&
        !listenersToProcessCards(rows).some((c) => c.id === selectedId)
      ) {
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

  function openKill(e: MouseEvent, card: ProcessCard) {
    e.stopPropagation();
    selectedId = card.id;
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

  const POLL_MS = 2000;
  const LABEL_TICK_MS = 1000;

  let pollTimer: ReturnType<typeof setInterval> | undefined;
  let labelTimer: ReturnType<typeof setInterval> | undefined;
  let unsubRefresh: (() => void) | undefined;

  function startAutoRefresh() {
    if (pollTimer !== undefined) return;
    void doRefresh();
    pollTimer = setInterval(() => {
      if (!refreshing) void doRefresh();
    }, POLL_MS);
  }

  function stopAutoRefresh() {
    if (pollTimer !== undefined) {
      clearInterval(pollTimer);
      pollTimer = undefined;
    }
  }

  function startLabelTimer() {
    if (labelTimer !== undefined) return;
    labelTimer = setInterval(() => {
      labelTick++;
    }, LABEL_TICK_MS);
  }

  function stopLabelTimer() {
    if (labelTimer !== undefined) {
      clearInterval(labelTimer);
      labelTimer = undefined;
    }
  }

  function onPanelVisibilityChange() {
    if (document.hidden) {
      stopAutoRefresh();
      stopLabelTimer();
    } else {
      startAutoRefresh();
      startLabelTimer();
    }
  }

  onMount(() => {
    unsubRefresh = onRefresh(() => void doRefresh());
    document.addEventListener("visibilitychange", onPanelVisibilityChange);
    if (!document.hidden) {
      startAutoRefresh();
      startLabelTimer();
    }
  });

  onDestroy(() => {
    stopAutoRefresh();
    stopLabelTimer();
    document.removeEventListener("visibilitychange", onPanelVisibilityChange);
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
    <span class="count-pill">{processCards.length}</span>
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
    {#if processCards.length === 0}
      <div class="empty">没有匹配的 LISTEN 端口</div>
    {:else}
      <div class="cards">
        {#each processCards as card (card.id)}
          <!-- svelte-ignore a11y_click_events_have_key_events -->
          <!-- svelte-ignore a11y_no_static_element_interactions -->
          <div
            class="pcard"
            class:on={card.id === selectedId}
            onclick={() => toggleCard(card.id)}
          >
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
                {card.name}
                {#if card.leftover}
                  <span class="tag">leftover</span>
                {/if}
              </div>
            </div>
            <div class="pcard-mid">
              <span>PID {card.pid}</span>
              <div class="bar" class:hot={cpuClass(card.cpuPercent) === "hot"}>
                <i style="width: {cpuBarWidth(card.cpuPercent)}%"></i>
              </div>
              <span>
                <span class={cpuClass(card.cpuPercent)}>{card.cpuPercent.toFixed(1)}%</span>
                · {fmtMem(card.memMb)}
              </span>
            </div>
            {#if card.id === selectedId}
              <div class="pcard-detail">
                <div class="ports-detail">
                  {#each card.ports as p (p.port + p.proto)}
                    <span class="port-chip">
                      <span class="port">:{p.port}</span>
                      <span class="badge {p.proto}">{p.proto}</span>
                    </span>
                  {/each}
                </div>
                <div class="kv">
                  <div class="r">
                    <div class="k">cmdline</div>
                    <!-- svelte-ignore a11y_click_events_have_key_events -->
                    <!-- svelte-ignore a11y_no_static_element_interactions -->
                    <div
                      class="v copyable"
                      title="点击复制"
                      onclick={(e) => copyValue(e, "cmdline", card.cmdline)}
                    >
                      {card.cmdline}
                    </div>
                  </div>
                  <div class="r">
                    <div class="k">cwd</div>
                    <!-- svelte-ignore a11y_click_events_have_key_events -->
                    <!-- svelte-ignore a11y_no_static_element_interactions -->
                    <div
                      class="v copyable"
                      class:q={card.cwd === "?"}
                      title="点击复制"
                      onclick={(e) => copyValue(e, "cwd", card.cwd)}
                    >
                      {card.cwd}
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
                        copyValue(
                          e,
                          "user",
                          `${card.user} · ${card.started}`,
                        )}
                    >
                      {card.user} · {card.started}
                    </div>
                  </div>
                </div>
                {#if card.chain.length > 0}
                  <div class="chain">
                    {#each card.chain as node, i}
                      {#if i > 0}
                        <span class="arrow">›</span>
                      {/if}
                      <span class="n">{node.name} · {node.pid}</span>
                    {/each}
                  </div>
                {/if}
                <div class="acts">
                  <button
                    type="button"
                    class="btn danger"
                    onclick={(e) => openKill(e, card)}
                  >
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
        PID {selected.pid} {selected.name} · {formatPortsSnippet(selected.ports)}
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

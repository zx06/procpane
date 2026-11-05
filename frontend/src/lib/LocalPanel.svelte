<script lang="ts">
  import { onMount } from "svelte";
  import {
    listListeners,
    killProcess,
    showProcess,
    onRefresh,
    onSettingsChanged,
    refreshTargetsPanel,
    type Listener,
    type ProcessInfo,
    type RefreshPayload,
  } from "./api";
  import {
    filterListenersBySearch,
    formatPortsSnippet,
    listenersToProcessCards,
    type ProcessCard,
  } from "./processCards";
  import { displayProcessName } from "./displayListener";
  import {
    formatLastUpdated,
    loadRefreshIntervalMs,
    settingsAffectsSource,
    watchRefreshInterval,
  } from "./refreshPoll";
  import { bindPanelVisibility, shouldPoll } from "./panelPoll";
  import RefreshButton from "./RefreshButton.svelte";

  let { active = true }: { active?: boolean } = $props();

  let rows: Listener[] = $state([]);
  let selectedId: string | null = $state(null);
  let protoFilter = $state("all");
  let query = $state("");
  let lastSuccessAt = $state<number | null>(null);
  let refreshing = $state(false);
  let killOpen = $state(false);
  let procDetailOpen = $state(false);
  let procDetailLoading = $state(false);
  let procDetail: ProcessInfo | null = $state(null);
  let procDetailNote = $state("");
  let toast = $state("");
  let toastVisible = $state(false);
  let fetchError = $state("");
  let initialLoad = $state(true);
  const processCards = $derived(
    listenersToProcessCards(filterListenersBySearch(rows, query)),
  );
  const allCards = $derived(listenersToProcessCards(rows));

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
    return formatLastUpdated(lastSuccessAt);
  }

  async function doRefresh() {
    refreshing = true;
    fetchError = "";
    try {
      rows = await listListeners(protoFilter === "all" ? "all" : protoFilter);
      lastSuccessAt = Date.now();
      if (
        selectedId &&
        !listenersToProcessCards(rows).some((c) => c.id === selectedId)
      ) {
        selectedId = null;
      }
    } catch (e) {
      console.error(e);
      fetchError =
        e instanceof Error ? e.message : "无法加载本地 LISTEN 列表";
    } finally {
      refreshing = false;
      initialLoad = false;
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

  function chainNodeNote(info: ProcessInfo, fallback: string) {
    if (fallback) return fallback;
    return info.chain.find((n) => n.pid === info.pid)?.note ?? "";
  }

  function closeProcDetail() {
    procDetailOpen = false;
    procDetail = null;
    procDetailNote = "";
  }

  async function openProcessDetail(
    e: MouseEvent,
    pid: number,
    note = "",
  ) {
    e.stopPropagation();
    procDetailOpen = true;
    procDetailLoading = true;
    procDetail = null;
    procDetailNote = note;
    try {
      procDetail = await showProcess(pid);
    } catch (err) {
      console.error(err);
      showToast("失败");
      closeProcDetail();
    } finally {
      procDetailLoading = false;
    }
  }

  function onWindowKeydown(e: KeyboardEvent) {
    if (e.key !== "Escape") return;
    if (procDetailOpen) closeProcDetail();
    else if (killOpen) killOpen = false;
  }

  let pollMs = $state(30_000);
  let pollTimer: ReturnType<typeof setInterval> | undefined;
  let unsubRefresh: (() => void) | undefined;
  let unsubSettings: (() => void) | undefined;

  function startAutoRefresh() {
    if (pollTimer !== undefined) return;
    if (!shouldPoll(active)) return;
    void doRefresh();
    pollTimer = setInterval(() => {
      if (!refreshing && shouldPoll(active)) void doRefresh();
    }, pollMs);
  }

  function restartAutoRefreshPoll() {
    stopAutoRefresh();
    if (shouldPoll(active)) {
      startAutoRefresh();
    }
  }

  function stopAutoRefresh() {
    if (pollTimer !== undefined) {
      clearInterval(pollTimer);
      pollTimer = undefined;
    }
  }

  onMount(async () => {
    pollMs = await loadRefreshIntervalMs();
    unsubRefresh = onRefresh((data: RefreshPayload) => {
      if (refreshTargetsPanel(data, "local")) void doRefresh();
    });
    unsubSettings = watchRefreshInterval((ms) => {
      pollMs = ms;
      restartAutoRefreshPoll();
    });
    const unsubSettingsChanged = onSettingsChanged((data) => {
      if (settingsAffectsSource(data, "local")) {
        restartAutoRefreshPoll();
        void doRefresh();
      }
    });
    const unbindVis = bindPanelVisibility(
      () => active,
      () => restartAutoRefreshPoll(),
      () => stopAutoRefresh(),
    );
    window.addEventListener("keydown", onWindowKeydown);
    if (shouldPoll(active)) {
      startAutoRefresh();
    }
    return () => {
      stopAutoRefresh();
      unbindVis();
      unsubRefresh?.();
      unsubSettings?.();
      unsubSettingsChanged();
      window.removeEventListener("keydown", onWindowKeydown);
    };
  });

  $effect(() => {
    if (shouldPoll(active)) restartAutoRefreshPoll();
    else stopAutoRefresh();
  });
</script>

<div class="tab-panel">
  <header class="pop-hd">
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
    <RefreshButton spinning={refreshing} onclick={() => doRefresh()} />
  </div>

  {#if fetchError && rows.length > 0}
    <div class="hint-bar warn">
      {fetchError}
      <button type="button" class="link-btn" onclick={() => doRefresh()}>再试一次</button>
    </div>
  {/if}

  <div class="cards-scroll">
    {#if processCards.length === 0}
      <div class="empty">
        {#if fetchError && rows.length === 0}
          <p>{fetchError}</p>
          <button type="button" class="btn primary" onclick={() => doRefresh()}>再试一次</button>
        {:else if refreshing || initialLoad}
          <p>加载中…</p>
        {:else if rows.length === 0}
          <p>当前没有 LISTEN 端口</p>
        {:else}
          <p>没有匹配「{query}」的结果</p>
          <p class="empty-hint">共 {allCards.length} 个进程在监听，请调整搜索或协议筛选。</p>
        {/if}
      </div>
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
                {displayProcessName(card.name)}
                {#if card.leftover}
                  <span class="tag">leftover</span>
                {/if}
              </div>
            </div>
            <div class="pcard-mid">
              <!-- svelte-ignore a11y_click_events_have_key_events -->
              <!-- svelte-ignore a11y_no_static_element_interactions -->
              <span
                class="pid-link"
                role="button"
                tabindex="0"
                title="查看进程详情"
                onclick={(e) => openProcessDetail(e, card.pid)}
              >
                PID {card.pid}
              </span>
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
                      <!-- svelte-ignore a11y_click_events_have_key_events -->
                      <!-- svelte-ignore a11y_no_static_element_interactions -->
                      <span
                        class="n clickable"
                        role="button"
                        tabindex="0"
                        title="查看进程详情"
                        onclick={(e) =>
                          openProcessDetail(e, node.pid, node.note)}
                      >
                        {node.name} · {node.pid}
                      </span>
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
  </footer>
</div>

{#if procDetailOpen}
  <!-- svelte-ignore a11y_click_events_have_key_events -->
  <div class="overlay" role="presentation" onclick={closeProcDetail}>
    <!-- svelte-ignore a11y_no_static_element_interactions -->
    <div
      class="sheet sheet-detail"
      role="dialog"
      aria-modal="true"
      aria-labelledby="proc-detail-title"
      tabindex="-1"
      onclick={(e) => e.stopPropagation()}
    >
      {#if procDetailLoading}
        <p class="sheet-loading">加载中…</p>
      {:else if procDetail}
        <h3 id="proc-detail-title">{procDetail.name}</h3>
        <div class="detail-sub">PID {procDetail.pid}</div>
        {#if chainNodeNote(procDetail, procDetailNote)}
          <div class="detail-note">{chainNodeNote(procDetail, procDetailNote)}</div>
        {/if}
        <div class="kv">
          <div class="r">
            <div class="k">CPU</div>
            <!-- svelte-ignore a11y_click_events_have_key_events -->
            <!-- svelte-ignore a11y_no_static_element_interactions -->
            <div
              class="v copyable"
              title="点击复制"
              onclick={(e) =>
                copyValue(e, "CPU", `${procDetail!.cpuPercent.toFixed(1)}%`)}
            >
              {procDetail.cpuPercent.toFixed(1)}%
            </div>
          </div>
          <div class="r">
            <div class="k">mem</div>
            <!-- svelte-ignore a11y_click_events_have_key_events -->
            <!-- svelte-ignore a11y_no_static_element_interactions -->
            <div
              class="v copyable"
              title="点击复制"
              onclick={(e) =>
                copyValue(e, "mem", fmtMem(procDetail!.memMb))}
            >
              {fmtMem(procDetail.memMb)}
            </div>
          </div>
          <div class="r">
            <div class="k">cmdline</div>
            <!-- svelte-ignore a11y_click_events_have_key_events -->
            <!-- svelte-ignore a11y_no_static_element_interactions -->
            <div
              class="v copyable"
              title="点击复制"
              onclick={(e) => copyValue(e, "cmdline", procDetail!.cmdline)}
            >
              {procDetail.cmdline}
            </div>
          </div>
          <div class="r">
            <div class="k">cwd</div>
            <!-- svelte-ignore a11y_click_events_have_key_events -->
            <!-- svelte-ignore a11y_no_static_element_interactions -->
            <div
              class="v copyable"
              class:q={procDetail.cwd === "?"}
              title="点击复制"
              onclick={(e) => copyValue(e, "cwd", procDetail!.cwd)}
            >
              {procDetail.cwd}
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
                  `${procDetail!.user} · ${procDetail!.started}`,
                )}
            >
              {procDetail.user} · {procDetail.started}
            </div>
          </div>
        </div>
        {#if procDetail.chain.length > 0}
          <div class="section-label">进程链</div>
          <div class="chain">
            {#each procDetail.chain as node, i}
              {#if i > 0}
                <span class="arrow">›</span>
              {/if}
              <!-- svelte-ignore a11y_click_events_have_key_events -->
              <!-- svelte-ignore a11y_no_static_element_interactions -->
              <span
                class="n clickable"
                role="button"
                tabindex="0"
                title="查看进程详情"
                onclick={(e) => openProcessDetail(e, node.pid, node.note)}
              >
                {node.name} · {node.pid}
              </span>
            {/each}
          </div>
        {/if}
        <div class="rowbtns">
          <button type="button" class="btn" onclick={closeProcDetail}>关闭</button>
        </div>
      {/if}
    </div>
  </div>
{/if}

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

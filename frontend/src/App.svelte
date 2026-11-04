<script lang="ts">
  import { onMount, onDestroy } from "svelte";
  import {
    listListeners,
    killProcess,
    onRefresh,
    type Listener,
  } from "./lib/api";

  let rows: Listener[] = $state([]);
  let selectedId: string | null = $state(null);
  let protoFilter = $state("all");
  let query = $state("");
  let lastRefresh = $state(Date.now());
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
    if (c >= 10) return "metric hot";
    if (c >= 5) return "metric warn";
    return "metric";
  }

  function refreshLabel() {
    const sec = (Date.now() - lastRefresh) / 1000;
    return refreshing ? "刷新中…" : `已刷新 · ${sec.toFixed(1)}s 前`;
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

  function selectRow(id: string) {
    selectedId = id;
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

  async function copyCmdline() {
    if (!selected) return;
    try {
      await navigator.clipboard.writeText(selected.cmdline);
      showToast("已复制 cmdline");
    } catch {
      showToast("复制失败");
    }
  }

  let tick: ReturnType<typeof setInterval>;
  let unsubRefresh: (() => void) | undefined;

  onMount(() => {
    doRefresh();
    unsubRefresh = onRefresh(() => doRefresh());
    window.addEventListener("focus", () => (focused = true));
    window.addEventListener("blur", () => (focused = false));
    tick = setInterval(() => {
      lastRefresh = lastRefresh; // tick label
      const interval = focused ? 2000 : 5000;
      if (Date.now() - lastRefresh >= interval && !refreshing) {
        doRefresh();
      }
    }, 200);
  });

  onDestroy(() => {
    clearInterval(tick);
    unsubRefresh?.();
  });
</script>

<div class="app">
  <div class="titlebar">
    <svg class="mark" viewBox="0 0 18 18" aria-hidden="true">
      <rect x="2" y="2" width="14" height="14" rx="3" fill="none" stroke="currentColor" stroke-width="1.2" />
      <rect x="4.5" y="5" width="9" height="1.2" rx="0.6" fill="currentColor" />
      <rect x="4.5" y="8.2" width="9" height="1.2" rx="0.6" fill="currentColor" />
      <rect x="4.5" y="11.4" width="9" height="1.2" rx="0.6" fill="currentColor" />
    </svg>
    <h1>procpane</h1>
    <span class="proto-badge">LISTEN · v1</span>
  </div>

  <div class="toolbar">
    <div class="search">
      <input
        type="search"
        placeholder="端口 / 进程…"
        bind:value={query}
      />
    </div>
    <div class="chips">
      {#each ["all", "tcp", "udp"] as p}
        <button
          type="button"
          class="chip"
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
    <div class="refresh-badge">
      <span class="pulse" class:busy={refreshing}></span>
      <span>{refreshLabel()}</span>
    </div>
    <button type="button" class="btn" onclick={() => doRefresh()}>刷新</button>
  </div>

  <div class="body">
    <div class="list-pane">
      <div class="table-wrap">
        <table>
          <thead>
            <tr>
              <th>Port</th>
              <th>Proto</th>
              <th>PID</th>
              <th>Name</th>
              <th>CPU</th>
              <th>Mem</th>
            </tr>
          </thead>
          <tbody>
            {#each filtered as r (r.id)}
              <tr
                class:selected={r.id === selectedId}
                onclick={() => selectRow(r.id)}
              >
                <td class="port">:{r.port}</td>
                <td><span class="proto {r.proto}">{r.proto}</span></td>
                <td>{r.pid}</td>
                <td>
                  {r.name}
                  {#if r.leftover}
                    <span class="leftover"> leftover</span>
                  {/if}
                </td>
                <td class={cpuClass(r.cpuPercent)}>{r.cpuPercent.toFixed(1)}%</td>
                <td class="metric">{fmtMem(r.memMb)}</td>
              </tr>
            {/each}
          </tbody>
        </table>
        {#if filtered.length === 0}
          <div class="detail-empty">没有匹配的 LISTEN 端口</div>
        {/if}
      </div>
      <div class="list-foot">
        <span>{filtered.length} 个监听</span>
        <span>tcp/udp · LISTEN</span>
      </div>
    </div>

    <div class="detail-pane">
      {#if !selected}
        <div class="detail-empty">
          选择一行查看详情<br />
          <span style="font-size:10px;opacity:.85">cmdline · cwd · 父进程链</span>
        </div>
      {:else}
        <div class="detail-hd">
          <h2>{selected.name}{selected.leftover ? " · leftover" : ""}</h2>
        </div>
        <div class="detail-sub">
          :{selected.port} / {selected.proto.toUpperCase()} · PID {selected.pid}
        </div>
        <div class="kv">
          <div class="row">
            <div class="k">cmdline</div>
            <div class="v">{selected.cmdline}</div>
          </div>
          <div class="row">
            <div class="k">cwd</div>
            <div class="v" class:q={selected.cwd === "?"}>{selected.cwd}</div>
          </div>
          <div class="row">
            <div class="k">user</div>
            <div class="v">{selected.user}</div>
          </div>
          <div class="row">
            <div class="k">started</div>
            <div class="v">{selected.started}</div>
          </div>
        </div>
        <div class="section-title">Parent chain</div>
        <div class="chain">
          {#each selected.chain as node, i}
            <div class="node">
              <div class="lvl">{i + 1}</div>
              <div>
                <div>{node.name} · {node.pid}</div>
                <div style="color:var(--muted);font-size:10px">{node.note}</div>
              </div>
            </div>
          {/each}
        </div>
        <div class="detail-actions">
          <button type="button" class="btn danger" onclick={() => (killOpen = true)}>
            结束进程
          </button>
          <button type="button" class="btn" onclick={copyCmdline}>复制 cmdline</button>
        </div>
      {/if}
    </div>
  </div>
</div>

{#if killOpen && selected}
  <!-- svelte-ignore a11y_click_events_have_key_events -->
  <div class="overlay" role="presentation" onclick={() => (killOpen = false)}>
    <!-- svelte-ignore a11y_no_static_element_interactions -->
    <div class="modal" onclick={(e) => e.stopPropagation()}>
      <div class="modal-hd">
        <div class="warn-icon">!</div>
        <h3 style="margin:0;font-size:13px">确认结束进程？</h3>
      </div>
      <div class="modal-body">
        将向目标进程发送终止信号（SIGTERM）。结束后列表会刷新。
        <div class="modal-snippet">
          PID {selected.pid} {selected.name} · :{selected.port}/{selected.proto}
        </div>
      </div>
      <div class="modal-ft">
        <button type="button" class="btn" onclick={() => (killOpen = false)}>取消</button>
        <button type="button" class="btn danger" onclick={confirmKill}>结束进程</button>
      </div>
    </div>
  </div>
{/if}

{#if toastVisible}
  <div class="toast">{toast}</div>
{/if}

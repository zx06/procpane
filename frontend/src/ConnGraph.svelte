<script lang="ts">
  import { onMount } from "svelte";
  import RefreshButton from "./lib/RefreshButton.svelte";
  import { listLocalConnections } from "./lib/api";
  import {
    ConnGraphCanvas,
    connVisible,
    dirLabel,
    toUiConn,
    type GraphFilters,
    type GraphSelection,
    type UiConn,
  } from "./lib/connGraphEngine";

  let connections = $state<UiConn[]>([]);
  let loading = $state(true);
  let refreshing = $state(false);
  let loadError = $state("");
  let updatedAt = $state("");

  let filters = $state<GraphFilters>({
    localOnly: false,
    inOnly: false,
    outOnly: false,
    hideBrowser: true,
    search: "",
  });

  let selection = $state<GraphSelection>({ selectedPid: null, selectedEdgeId: null });

  let canvasEl: HTMLCanvasElement | undefined = $state();
  let wrapEl: HTMLElement | undefined = $state();
  let engine: ConnGraphCanvas | null = null;

  function namesByPid(): Map<number, string> {
    const m = new Map<number, string>();
    for (const c of connections) {
      m.set(c.pid, c.name);
      if (c.peerPid) m.set(c.peerPid, c.peerName || "?");
    }
    return m;
  }

  function filteredTableRows(): UiConn[] {
    const names = namesByPid();
    return connections.filter((c) => connVisible(c, filters, names));
  }

  function sideRows(): UiConn[] {
    const pid = selection.selectedPid;
    if (pid == null) return [];
    return filteredTableRows().filter((c) => c.pid === pid || c.peerPid === pid);
  }

  function statsFor(pid: number) {
    const rows = filteredTableRows().filter((c) => c.pid === pid || c.peerPid === pid);
    return {
      local: rows.filter((c) => c.dir === "local").length,
      in: rows.filter((c) => c.dir === "in").length,
      out: rows.filter((c) => c.dir === "out").length,
    };
  }

  async function load() {
    refreshing = true;
    loadError = "";
    try {
      const res = await listLocalConnections();
      connections = res.connections.map(toUiConn);
      updatedAt = new Date(res.generatedAt).toLocaleTimeString("zh-CN", { hour12: false });
      engine?.setData(connections, filters);
    } catch (e) {
      console.error(e);
      loadError = "加载连接失败";
    } finally {
      loading = false;
      refreshing = false;
    }
  }

  function toggleDirection(key: "localOnly" | "inOnly" | "outOnly") {
    const next = !filters[key];
    filters = {
      ...filters,
      localOnly: key === "localOnly" ? next : false,
      inOnly: key === "inOnly" ? next : false,
      outOnly: key === "outOnly" ? next : false,
    };
    engine?.setData(connections, filters);
  }

  function toggleHideBrowser() {
    filters = { ...filters, hideBrowser: !filters.hideBrowser };
    engine?.setData(connections, filters);
  }

  function onSearchInput(ev: Event) {
    filters = { ...filters, search: (ev.currentTarget as HTMLInputElement).value };
    engine?.setData(connections, filters);
  }

  function selectPid(pid: number | null, edgeId: string | null = null) {
    selection = { selectedPid: pid, selectedEdgeId: edgeId };
    engine?.setSelection(selection);
  }

  function selectRow(c: UiConn) {
    selectPid(c.pid, c.id);
  }

  $effect(() => {
    if (!canvasEl || !wrapEl || engine) return;
    engine = new ConnGraphCanvas(canvasEl, wrapEl, (sel) => {
      selection = sel;
    });
    engine.bindEvents();
    if (connections.length) engine.setData(connections, filters);
    return () => engine?.destroy();
  });

  onMount(() => {
    void load();
  });
</script>

<div class="conn-graph-app">
  <header class="cg-toolbar">
    <label class="cg-search">
      <svg viewBox="0 0 16 16" fill="none" aria-hidden="true">
        <circle cx="7" cy="7" r="4.5" stroke="currentColor" stroke-width="1.6" />
        <path d="M10.5 10.5 L14 14" stroke="currentColor" stroke-width="1.6" stroke-linecap="round" />
      </svg>
      <input
        type="search"
        placeholder="进程名 / PID / 端口"
        value={filters.search}
        oninput={onSearchInput}
      />
    </label>
    <div class="cg-chips">
      <button
        type="button"
        class="cg-chip local-filter"
        class:on={filters.localOnly}
        onclick={() => toggleDirection("localOnly")}>只看本地互联</button
      >
      <button
        type="button"
        class="cg-chip"
        class:on={filters.inOnly}
        onclick={() => toggleDirection("inOnly")}>只看入站</button
      >
      <button
        type="button"
        class="cg-chip"
        class:on={filters.outOnly}
        onclick={() => toggleDirection("outOnly")}>只看出站</button
      >
      <button
        type="button"
        class="cg-chip"
        class:on={filters.hideBrowser}
        onclick={toggleHideBrowser}>隐藏浏览器类</button
      >
    </div>
    <RefreshButton spinning={refreshing} onclick={() => load()} />
    <div class="cg-legend">
      <span><i class="l-local"></i>本地互联</span>
      <span><i class="l-in"></i>入站</span>
      <span><i class="l-out"></i>出站</span>
    </div>
  </header>

  {#if loadError}
    <p class="cg-load-err">{loadError}</p>
  {/if}

  <div class="cg-body">
    <div class="cg-graph-wrap" bind:this={wrapEl}>
      <canvas bind:this={canvasEl}></canvas>
      <div class="cg-graph-hint">拖拽节点 · 滚轮缩放 · 点击选中</div>
    </div>

    <aside class="cg-side">
      {#if selection.selectedPid == null}
        <div class="cg-side-empty">
          <strong>未选中进程</strong>
          点击图中进程节点查看连接明细
        </div>
      {:else}
        {@const pid = selection.selectedPid}
        {@const procName = namesByPid().get(pid) ?? "?"}
        {@const st = statsFor(pid)}
        <div class="cg-side-hd">
          <h2>{procName}</h2>
          <div class="sub">PID {pid}</div>
        </div>
        <div class="cg-stat-row">
          <div class="cg-stat local">
            <div class="n">{st.local}</div>
            <div class="l">本地互联</div>
          </div>
          <div class="cg-stat in">
            <div class="n">{st.in}</div>
            <div class="l">入站</div>
          </div>
          <div class="cg-stat out">
            <div class="n">{st.out}</div>
            <div class="l">出站</div>
          </div>
        </div>
        <div class="cg-conn-list">
          {#each sideRows() as c (c.id)}
            <div
              class="cg-conn-row"
              class:on={selection.selectedEdgeId === c.id}
              role="button"
              tabindex="0"
              onclick={() => selectRow(c)}
              onkeydown={(ev) => ev.key === "Enter" && selectRow(c)}
            >
              <div class="addrs">{c.local} → {c.remote}</div>
              <div class="meta">
                <span class="cg-dir-badge {c.dir}">{dirLabel(c.dir)}</span>
                {#if c.peerPid}
                  <span class="peer">{c.peerName || "?"} · {c.peerPid}</span>
                {/if}
              </div>
            </div>
          {/each}
        </div>
      {/if}
    </aside>

    <div class="cg-table-wrap">
      <table class="cg-table">
        <thead>
          <tr>
            <th>进程</th>
            <th>本地</th>
            <th>远端</th>
            <th>方向</th>
            <th>对端进程</th>
          </tr>
        </thead>
        <tbody>
          {#if loading}
            <tr><td colspan="5">加载中…</td></tr>
          {:else if filteredTableRows().length === 0}
            <tr><td colspan="5">无匹配的 ESTABLISHED 连接</td></tr>
          {:else}
            {#each filteredTableRows() as c (c.id)}
              <tr
                class:on={c.pid === selection.selectedPid || c.id === selection.selectedEdgeId}
                onclick={() => selectRow(c)}
              >
                <td class="pname">{c.name} {c.pid}</td>
                <td class="mono">{c.local}</td>
                <td class="mono">{c.remote}</td>
                <td><span class="cg-dir-badge {c.dir}">{dirLabel(c.dir)}</span></td>
                <td class="mono"
                  >{#if c.peerPid}{c.peerName} {c.peerPid}{:else}—{/if}</td
                >
              </tr>
            {/each}
          {/if}
        </tbody>
      </table>
      <div class="cg-foot">
        <span
          >{filteredTableRows().length} 条连接{#if updatedAt}
            · 上次更新 {updatedAt}{/if}</span
        >
        <span>TCP ESTABLISHED · 本机</span>
      </div>
    </div>
  </div>
</div>

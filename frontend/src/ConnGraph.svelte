<script lang="ts">
  import { onMount } from "svelte";
  import {
    Background,
    BackgroundVariant,
    Controls,
    MiniMap,
    SvelteFlow,
    type Edge,
    type Node,
  } from "@xyflow/svelte";
  import "@xyflow/svelte/dist/style.css";
  import RefreshButton from "./lib/RefreshButton.svelte";
  import ConnGraphProcessNode from "./lib/ConnGraphProcessNode.svelte";
  import ConnGraphFitView from "./lib/ConnGraphFitView.svelte";
  import { listLocalConnections } from "./lib/api";
  import { buildFlowGraph } from "./lib/connGraphFlow";
  import {
    buildLocalComponents,
    componentForPid,
    componentLabel,
    loadRememberedComponentId,
    pickDefaultComponentId,
    rememberComponentId,
    type LocalComponent,
  } from "./lib/connGraphClusters";
  import {
    dirLabel,
    displayTableRows,
    namesByPidFrom,
    toUiConn,
    type GraphFilters,
    type GraphSelection,
    type UiConn,
  } from "./lib/connGraphEngine";

  const nodeTypes = {
    process: ConnGraphProcessNode,
  };

  let connections = $state<UiConn[]>([]);
  let localComponents = $state<LocalComponent[]>([]);
  let activeComponentId = $state<string | null>(null);
  let loading = $state(true);
  let refreshing = $state(false);
  let loadError = $state("");
  let updatedAt = $state("");

  let filters = $state<GraphFilters>({
    localOnly: false,
    inOnly: false,
    outOnly: false,
    search: "",
  });

  /** Last component id we ran dagre layout for (preserve drag within same group). */
  let layoutComponentId = $state<string | null>(null);

  let selection = $state<GraphSelection>({ selectedPid: null, selectedEdgeId: null });

  let nodes = $state.raw<Node[]>([]);
  let edges = $state.raw<Edge[]>([]);
  let fitRevision = $state(0);

  function namesByPid(): Map<number, string> {
    return namesByPidFrom(connections);
  }

  function activeComponent(): LocalComponent | null {
    if (!activeComponentId) return null;
    return localComponents.find((c) => c.id === activeComponentId) ?? null;
  }

  function activeComponentIndex(): number {
    if (!activeComponentId) return 0;
    const i = localComponents.findIndex((c) => c.id === activeComponentId);
    return i >= 0 ? i + 1 : 0;
  }

  function filteredTableRows(): UiConn[] {
    return displayTableRows(connections, filters, namesByPid());
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

  function recomputeComponents(remember = true) {
    localComponents = buildLocalComponents(connections);
    const prev = remember ? activeComponentId ?? loadRememberedComponentId() : null;
    activeComponentId = pickDefaultComponentId(localComponents, prev);
    rememberComponentId(activeComponentId);
  }

  function syncFlowGraph(opts?: { fit?: boolean; relayout?: boolean }) {
    const fit = opts?.fit ?? false;
    const relayout =
      opts?.relayout ??
      (activeComponentId !== layoutComponentId || nodes.length === 0);
    const built = buildFlowGraph(
      connections,
      activeComponent(),
      selection,
      relayout ? [] : nodes,
      { relayout },
    );
    nodes = built.nodes;
    edges = built.edges;
    if (relayout && activeComponentId) layoutComponentId = activeComponentId;
    if (fit) fitRevision += 1;
  }

  function setActiveComponent(id: string, fit = true) {
    if (id === activeComponentId) return;
    activeComponentId = id;
    rememberComponentId(id);
    selection = { selectedPid: null, selectedEdgeId: null };
    layoutComponentId = null;
    syncFlowGraph({ fit, relayout: true });
  }

  async function load() {
    refreshing = true;
    loadError = "";
    try {
      const res = await listLocalConnections();
      connections = res.connections.map(toUiConn);
      updatedAt = new Date(res.generatedAt).toLocaleTimeString("zh-CN", { hour12: false });
      recomputeComponents(true);
      layoutComponentId = null;
      syncFlowGraph({ fit: true, relayout: true });
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
    syncFlowGraph({ relayout: false });
  }

  function onSearchInput(ev: Event) {
    filters = { ...filters, search: (ev.currentTarget as HTMLInputElement).value };
    syncFlowGraph({ relayout: false });
  }

  function onGroupChange(ev: Event) {
    const id = (ev.currentTarget as HTMLSelectElement).value;
    setActiveComponent(id, true);
  }

  function selectPid(pid: number | null, edgeId: string | null = null) {
    selection = { selectedPid: pid, selectedEdgeId: edgeId };
    syncFlowGraph({ relayout: false });
  }

  function selectRow(c: UiConn) {
    const comp = componentForPid(localComponents, c.pid);
    if (comp && comp.id !== activeComponentId) {
      setActiveComponent(comp.id, true);
    }
    selectPid(c.pid, c.id);
  }

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

    {#if localComponents.length > 0}
      <div class="cg-group-bar">
        <span class="cg-group-prefix">组</span>
        <select
          class="cg-group-select"
          aria-label="本地互联组"
          value={activeComponentId ?? ""}
          onchange={onGroupChange}
        >
          {#each localComponents as comp (comp.id)}
            <option value={comp.id}>{componentLabel(comp, namesByPid())}</option>
          {/each}
        </select>
        <span class="cg-group-meta"
          >{activeComponentIndex()}/{localComponents.length}</span
        >
      </div>
    {/if}

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
    </div>
    <RefreshButton spinning={refreshing} onclick={() => load()} />
    <div class="cg-legend">
      <span><i class="l-local"></i>本地互联</span>
      <span class="cg-legend-note">图仅当前组</span>
    </div>
  </header>

  {#if loadError}
    <p class="cg-load-err">{loadError}</p>
  {/if}

  <div class="cg-body">
    <div class="cg-graph-wrap cg-flow">
      <SvelteFlow
        bind:nodes
        bind:edges
        {nodeTypes}
        nodesConnectable={false}
        elementsSelectable={true}
        minZoom={0.35}
        maxZoom={2.5}
        colorMode="system"
        onnodeclick={({ node }) => {
          if (node.type === "process" && node.data?.pid != null) {
            selectPid(node.data.pid as number, null);
          }
        }}
        onedgeclick={({ edge }) => {
          const connId = (edge.data?.connId as string) ?? edge.id;
          const row = connections.find((c) => c.id === connId);
          if (row) selectRow(row);
          else selectPid(null, connId);
        }}
        onpaneclick={() => selectPid(null, null)}
      >
        <Background variant={BackgroundVariant.Dots} gap={18} size={1} />
        <Controls showInteractive={false} />
        <MiniMap nodeStrokeWidth={2} pannable zoomable class="cg-minimap" />
        <ConnGraphFitView revision={fitRevision} />
      </SvelteFlow>
      {#if !loading && localComponents.length === 0}
        <div class="cg-graph-empty">无本地互联组可展示</div>
      {/if}
      <div class="cg-graph-hint">拖拽节点 · 滚轮缩放 · 点击选中 · 底表可跳转其他组</div>
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
                {:else if c.unpairedLocal}
                  <span class="peer unpaired">未配对本地</span>
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
              {@const inActive =
                activeComponentId != null &&
                componentForPid(localComponents, c.pid)?.id === activeComponentId}
              <tr
                class:on={c.pid === selection.selectedPid || c.id === selection.selectedEdgeId}
                class:other-group={!inActive}
                onclick={() => selectRow(c)}
              >
                <td class="pname">{c.name} {c.pid}</td>
                <td class="mono">{c.local}</td>
                <td class="mono">{c.remote}</td>
                <td><span class="cg-dir-badge {c.dir}">{dirLabel(c.dir)}</span></td>
                <td class="mono"
                  >{#if c.peerPid}{c.peerName} {c.peerPid}{:else if c.unpairedLocal}<span
                      class="unpaired">未配对本地</span
                    >{:else}—{/if}</td
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
        <span>TCP ESTABLISHED · 本机 · 图={activeComponentIndex()}/{localComponents.length} 组</span>
      </div>
    </div>
  </div>
</div>

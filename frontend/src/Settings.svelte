<script lang="ts">
  import { onMount } from "svelte";
  import {
    getSettings,
    setSSHEnabled,
    setDockerContext,
    type SettingsView,
  } from "./lib/api";

  let settings: SettingsView | null = $state(null);
  let saving = $state(false);
  let message = $state("");

  async function load() {
    settings = await getSettings();
  }

  async function toggle(alias: string, enabled: boolean) {
    saving = true;
    message = "";
    try {
      await setSSHEnabled(alias, enabled);
      await load();
      message = "已保存";
    } catch (e) {
      console.error(e);
      message = "保存失败";
    } finally {
      saving = false;
    }
  }

  async function onDockerContextChange(ev: Event) {
    const sel = ev.currentTarget as HTMLSelectElement;
    saving = true;
    message = "";
    try {
      await setDockerContext(sel.value);
      await load();
      message = "已保存";
    } catch (e) {
      console.error(e);
      message = "保存失败";
    } finally {
      saving = false;
    }
  }

  function contextLabel(c: SettingsView["docker"]["contexts"][number]) {
    return c.current ? `${c.name} *` : c.name;
  }

  onMount(() => {
    void load();
  });
</script>

<div class="settings-app">
  <header class="settings-hd">
    <h1>procpane 设置</h1>
    <p class="sub">Docker 上下文 · SSH 主机启用</p>
  </header>

  {#if settings}
    <section class="settings-block">
      <h2>Docker</h2>
      <div class="readonly">
        <div class="row">
          <span class="k">Context</span>
          <select
            class="ctx-select"
            disabled={saving}
            value={settings.docker.selectedContext ?? ""}
            onchange={onDockerContextChange}
          >
            <option value="">（自动 — DOCKER_HOST / 本地 socket）</option>
            {#each settings.docker.contexts ?? [] as c (c.name)}
              <option value={c.name}>{contextLabel(c)}</option>
            {/each}
          </select>
        </div>
        <div class="row">
          <span class="k">Endpoint</span>
          <span class="v mono">{settings.docker.endpoint}</span>
        </div>
        <div class="row">
          <span class="k">状态</span>
          <span class="v" class:ok={settings.docker.reachable} class:bad={!settings.docker.reachable}>
            {settings.docker.reachable ? "已连接" : "未连接"}
          </span>
        </div>
        {#if settings.docker.currentContext}
          <p class="hint dim">Docker CLI 当前：{settings.docker.currentContext}</p>
        {/if}
        {#if settings.docker.error}
          <div class="err">{settings.docker.error}</div>
        {/if}
      </div>
    </section>

    <section class="settings-block">
      <h2>SSH 主机</h2>
      <p class="hint">来自 ~/.ssh/config；仅密钥 / agent（不支持密码）。</p>
      {#if settings.sshHosts.length === 0}
        <p class="empty">未找到 Host 条目</p>
      {:else}
        <ul class="host-list">
          {#each settings.sshHosts as h (h.alias)}
            <li>
              <label>
                <input
                  type="checkbox"
                  checked={h.enabled}
                  disabled={saving}
                  onchange={(e) =>
                    toggle(h.alias, (e.currentTarget as HTMLInputElement).checked)}
                />
                <span class="alias">{h.alias}</span>
                <span class="dim">{h.hostname}{h.port ? `:${h.port}` : ""}</span>
              </label>
            </li>
          {/each}
        </ul>
      {/if}
      <p class="state-path">状态文件 · {settings.statePath}</p>
    </section>
  {:else}
    <p class="loading">加载中…</p>
  {/if}

  {#if message}
    <div class="toast">{message}</div>
  {/if}
</div>

<style>
  .ctx-select {
    flex: 1;
    min-width: 0;
    max-width: 100%;
    font-family: inherit;
    font-size: 12px;
    padding: 4px 8px;
    border-radius: 6px;
    border: 1px solid var(--hair);
    background: var(--pop-elev);
    color: var(--text);
  }

  .ctx-select:disabled {
    opacity: 0.55;
    cursor: not-allowed;
  }
</style>

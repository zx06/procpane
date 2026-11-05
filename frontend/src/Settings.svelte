<script lang="ts">
  import { onMount } from "svelte";
  import {
    getSettings,
    setSSHEnabled,
    setDockerContext,
    setRefreshIntervalSec,
    setAutostart,
    emitSettingsChanged,
    type SettingsView,
  } from "./lib/api";

  const REFRESH_OPTIONS = [
    { sec: 15, label: "15 秒" },
    { sec: 30, label: "30 秒" },
    { sec: 60, label: "1 分钟" },
    { sec: 120, label: "2 分钟" },
    { sec: 300, label: "5 分钟" },
  ] as const;

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
      emitSettingsChanged({ sshHosts: true });
      message = "已保存";
    } catch (e) {
      console.error(e);
      message = "保存失败";
    } finally {
      saving = false;
    }
  }

  async function onAutostartChange(ev: Event) {
    const checked = (ev.currentTarget as HTMLInputElement).checked;
    saving = true;
    message = "";
    try {
      await setAutostart(checked);
      await load();
      message = "已保存";
    } catch (e) {
      console.error(e);
      message = "保存失败";
      await load();
    } finally {
      saving = false;
    }
  }

  async function onRefreshIntervalChange(ev: Event) {
    const sel = ev.currentTarget as HTMLSelectElement;
    const sec = Number(sel.value);
    saving = true;
    message = "";
    try {
      await setRefreshIntervalSec(sec);
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
      emitSettingsChanged({ dockerContext: true });
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
    <p class="sub">开机自启 · 自动刷新 · Docker · SSH</p>
  </header>

  {#if settings}
    <section class="settings-block">
      <h2>常规</h2>
      <p class="hint">登录系统后自动启动 procpane 托盘（默认关闭）。</p>
      <div class="readonly">
        <div class="row">
          <label class="toggle-row">
            <input
              type="checkbox"
              checked={settings.autostart ?? false}
              disabled={saving}
              onchange={onAutostartChange}
            />
            <span class="k">开机自启</span>
          </label>
        </div>
      </div>
    </section>

    <section class="settings-block">
      <h2>自动刷新</h2>
      <p class="hint">本机、Docker、SSH 标签页共用此间隔；手动 ↻ 仍为立即刷新。</p>
      <div class="readonly">
        <div class="row">
          <span class="k">间隔</span>
          <select
            class="ctx-select"
            disabled={saving}
            value={String(settings.refreshIntervalSec ?? 30)}
            onchange={onRefreshIntervalChange}
          >
            {#each REFRESH_OPTIONS as opt (opt.sec)}
              <option value={String(opt.sec)}>{opt.label}</option>
            {/each}
          </select>
        </div>
      </div>
    </section>

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

  .toggle-row {
    display: flex;
    align-items: center;
    gap: 8px;
    cursor: pointer;
    flex: 1;
  }

  .toggle-row input {
    margin: 0;
  }
</style>

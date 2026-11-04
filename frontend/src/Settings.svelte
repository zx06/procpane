<script lang="ts">
  import { onMount } from "svelte";
  import { getSettings, setSSHEnabled, type SettingsView } from "./lib/api";

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

  onMount(() => {
    void load();
  });
</script>

<div class="settings-app">
  <header class="settings-hd">
    <h1>procpane 设置</h1>
    <p class="sub">Docker 只读 · SSH 主机启用</p>
  </header>

  {#if settings}
    <section class="settings-block">
      <h2>Docker</h2>
      <div class="readonly">
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

<script lang="ts">
  import LocalPanel from "./lib/LocalPanel.svelte";
  import DockerPanel from "./lib/DockerPanel.svelte";
  import SshPanel from "./lib/SshPanel.svelte";
  import { onMount } from "svelte";
  import { quitApp, setActiveSource, version, type SourceTab } from "./lib/api";

  type Tab = SourceTab;
  let tab: Tab = $state("local");
  let ver = $state("");

  async function handleQuit() {
    try {
      await quitApp();
    } catch (e) {
      console.error(e);
    }
  }

  $effect(() => {
    void setActiveSource(tab).catch((e) => console.error(e));
  });

  onMount(async () => {
    try {
      ver = await version();
    } catch {
      ver = "";
    }
  });
</script>

<div class="app">
  <nav class="source-tabs" aria-label="数据源">
    {#each [
      ["local", "本机"],
      ["docker", "Docker"],
      ["ssh", "SSH"],
    ] as const as [id, label]}
      <button
        type="button"
        class:on={tab === id}
        onclick={() => (tab = id)}
      >
        {label}
      </button>
    {/each}
    {#if ver}
      <span class="ver-badge">{ver}</span>
    {/if}
  </nav>

  <div class="tab-stack">
    <div
      class="tab-panel-wrap"
      data-tab="local"
      class:hidden={tab !== "local"}
    >
      <LocalPanel active={tab === "local"} />
    </div>
    <div
      class="tab-panel-wrap"
      data-tab="docker"
      class:hidden={tab !== "docker"}
    >
      <DockerPanel active={tab === "docker"} />
    </div>
    <div class="tab-panel-wrap" data-tab="ssh" class:hidden={tab !== "ssh"}>
      <SshPanel active={tab === "ssh"} />
    </div>
  </div>

  <div class="global-quit">
    <button type="button" class="quit-btn" onclick={handleQuit}>退出</button>
  </div>
</div>

<style>
  .tab-stack {
    flex: 1;
    min-height: 0;
    display: flex;
    flex-direction: column;
  }

  .tab-panel-wrap {
    flex: 1;
    min-height: 0;
    display: flex;
    flex-direction: column;
  }

  .tab-panel-wrap.hidden {
    display: none;
  }
</style>

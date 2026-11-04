<script lang="ts">
  import LocalPanel from "./lib/LocalPanel.svelte";
  import DockerPanel from "./lib/DockerPanel.svelte";
  import SshPanel from "./lib/SshPanel.svelte";
  import { onMount } from "svelte";
  import { onRefresh, quitApp, version } from "./lib/api";

  type Tab = "local" | "docker" | "ssh";
  let tab: Tab = $state("local");
  let ver = $state("");

  async function handleQuit() {
    try {
      await quitApp();
    } catch (e) {
      console.error(e);
    }
  }

  onMount(async () => {
    try {
      ver = await version();
    } catch {
      ver = "";
    }
    onRefresh(() => {
      /* child panels poll; tray refresh is best-effort */
      window.dispatchEvent(new CustomEvent("procpane:refresh"));
    });
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

  {#if tab === "local"}
    <LocalPanel />
  {:else if tab === "docker"}
    <DockerPanel />
  {:else}
    <SshPanel />
  {/if}

  <div class="global-quit">
    <button type="button" class="quit-btn" onclick={handleQuit}>退出</button>
  </div>
</div>

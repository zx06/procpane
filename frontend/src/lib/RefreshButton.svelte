<script lang="ts">
  let { spinning = false, onclick }: { spinning?: boolean; onclick?: () => void } =
    $props();
</script>

<button
  type="button"
  class="icon-btn"
  class:spin={spinning}
  title="刷新"
  aria-label="刷新"
  {onclick}
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

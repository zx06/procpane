<script lang="ts">
  import { tick } from "svelte";
  import { useSvelteFlow } from "@xyflow/svelte";

  let { revision = 0 }: { revision?: number } = $props();
  const { fitView } = useSvelteFlow();

  $effect(() => {
    revision;
    void tick().then(() => {
      fitView({ padding: 0.18, duration: 200 });
    });
  });
</script>

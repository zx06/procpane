<script lang="ts">
  import { Handle, Position, type Node, type NodeProps } from "@xyflow/svelte";
  import type { StubNodeData } from "./connGraphFlow";

  let { data }: NodeProps<Node<StubNodeData>> = $props();
</script>

<div class="flow-stub" class:in={data.dir === "in"} class:dim={data.dimmed}>
  <Handle class="flow-handle" type="target" position={Position.Left} />
  <Handle class="flow-handle" type="source" position={Position.Right} />
  <span class="dot" aria-hidden="true"></span>
  <span class="label">{data.label}</span>
</div>

<style>
  .flow-stub {
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 4px;
    min-width: 40px;
    opacity: 1;
  }

  .flow-stub.dim {
    opacity: 0.28;
  }

  .dot {
    width: 14px;
    height: 14px;
    border-radius: 50%;
    background: rgba(100, 116, 139, 0.16);
    border: 1px solid var(--color-stub-edge);
  }

  .flow-stub.in .dot {
    background: var(--color-ok-soft);
    border-color: var(--color-in-edge);
  }

  .label {
    font-family: var(--font-mono);
    font-size: 9px;
    color: var(--color-muted);
    max-width: 120px;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
    text-align: center;
  }

  :global(.flow-handle) {
    width: 6px !important;
    height: 6px !important;
    min-width: 6px !important;
    min-height: 6px !important;
    background: var(--color-hair2) !important;
    border: 1px solid var(--color-muted) !important;
  }
</style>

<script lang="ts">
  import { Handle, Position, type Node, type NodeProps } from "@xyflow/svelte";
  import type { ProcessNodeData } from "./connGraphFlow";

  let { data, selected }: NodeProps<Node<ProcessNodeData>> = $props();
</script>

<div
  class="pcard flow-proc-node"
  class:on={selected}
  class:dim={data.dimmed}
  class:has-local={data.hasLocal}
>
  <Handle class="flow-handle" type="target" position={Position.Left} />
  <Handle class="flow-handle" type="source" position={Position.Right} />
  <Handle class="flow-handle" type="target" position={Position.Top} id="t-top" />
  <Handle class="flow-handle" type="source" position={Position.Bottom} id="s-bottom" />
  <div class="flow-proc-inner">
    <div class="name" class:unknown={data.name === "?"}>{data.name}</div>
    <div class="pid">{data.pid}</div>
  </div>
</div>

<style>
  .flow-proc-node {
    min-width: 88px;
    padding: 8px 10px;
    cursor: grab;
    position: relative;
  }

  .flow-proc-node.dim {
    opacity: 0.28;
  }

  .flow-proc-node.has-local {
    box-shadow: inset 0 0 0 2px color-mix(in srgb, var(--local) 35%, transparent);
  }

  .flow-proc-inner {
    text-align: center;
    pointer-events: none;
  }

  .flow-proc-inner .name {
    font-size: 11px;
    font-weight: 650;
    max-width: 120px;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
    margin: 0 auto;
  }

  .flow-proc-inner .pid {
    font-family: var(--mono);
    font-size: 9px;
    color: var(--muted);
    margin-top: 2px;
  }

  :global(.flow-handle) {
    width: 6px !important;
    height: 6px !important;
    min-width: 6px !important;
    min-height: 6px !important;
    background: var(--hair) !important;
    border: 1px solid var(--muted) !important;
  }
</style>

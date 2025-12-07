<script>
    export let title = "Widget";
    export let className = "";
</script>

<div
    class="flex flex-col h-full w-full bg-[var(--background-primary)]/80 backdrop-blur-md border border-[var(--border-subtle)] rounded-sm overflow-hidden shadow-lg {className}"
>
    <!-- Widget Header -->
    <div
        class="flex items-center justify-between px-3 py-1 bg-[var(--background-secondary)]/90 border-b border-[var(--border-subtle)] shrink-0"
    >
        <span
            class="text-xs font-mono uppercase tracking-wider text-[var(--text-muted)] select-none"
            >{title}</span
        >
        <div class="flex gap-1">
            <!-- Window Controls (Decorative for now) -->
            <div class="w-2 h-2 rounded-full bg-[var(--status-dnd)]/50"></div>
            <div class="w-2 h-2 rounded-full bg-[var(--status-idle)]/50"></div>
            <div
                class="w-2 h-2 rounded-full bg-[var(--status-online)]/50"
            ></div>
        </div>
    </div>

    <!-- Widget Content -->
    <div class="flex-1 overflow-hidden relative">
        <slot />
    </div>
</div>

<style>
    /* Cyberpunk/IRC aesthetic tweaks */
    :global(.widget-content) {
        scrollbar-width: thin;
        scrollbar-color: var(--border-subtle) transparent;
    }
</style>

<script>
    import { appContext } from "../stores";

    import GuildsWidget from "./GuildsWidget.svelte";
    import ChannelsWidget from "./ChannelsWidget.svelte";
    import ChatWidget from "./ChatWidget.svelte";
    import ProfileWidget from "./ProfileWidget.svelte";
    import GameTrackerWidget from "./GameTrackerWidget.svelte";
    import MembersWidget from "./MembersWidget.svelte";
    import { fade, fly } from "svelte/transition";
    import { flip } from "svelte/animate";

    // We can subscribe to the store directly here
    $: activePeerId = $appContext.activePeerId;
    $: activeChannelId = $appContext.activeChannelId;

    // Determine what to show in ChatWidget
    $: chatTargetId =
        activeChannelId || (activePeerId === "home" ? null : activePeerId);

    // Layout State
    let isEditing = false;
    let showRightSidebar = true;
    let showProfile = true;

    // Determine right sidebar content
    $: isHome = activePeerId === "home" || !activePeerId;

    // Drag & Drop State
    let widgetOrder = [
        { id: "nodes", component: "GuildsWidget", width: "w-[72px]" },
        { id: "channels", component: "ChannelsWidget", width: "w-64" },
        { id: "chat", component: "ChatWidget", width: "flex-1" },
        { id: "sidebar", component: "RightSidebar", width: "w-72" },
    ];

    let draggingItem = null;
    let hoveringItem = null;

    function handleDragStart(e, item) {
        if (!isEditing) return;
        draggingItem = item;
        e.dataTransfer.effectAllowed = "move";
        e.dataTransfer.setData("text/plain", item.id);
        // Add a ghost class or style if needed
    }

    function handleDragOver(e, item) {
        if (!isEditing || !draggingItem || draggingItem === item) return;
        e.preventDefault(); // Necessary to allow dropping
        hoveringItem = item;
    }

    function handleDrop(e, targetItem) {
        if (!isEditing || !draggingItem) return;
        e.preventDefault();

        const fromIndex = widgetOrder.indexOf(draggingItem);
        const toIndex = widgetOrder.indexOf(targetItem);

        if (fromIndex !== toIndex) {
            // Reorder
            const newOrder = [...widgetOrder];
            newOrder.splice(fromIndex, 1);
            newOrder.splice(toIndex, 0, draggingItem);
            widgetOrder = newOrder;
        }

        draggingItem = null;
        hoveringItem = null;
    }

    function handleDragEnd() {
        draggingItem = null;
        hoveringItem = null;
    }
</script>

<div
    class="flex h-full w-full overflow-hidden bg-[var(--background-tertiary)] p-2 gap-2 font-sans text-[var(--text-normal)] transition-all duration-300 ease-in-out"
>
    {#each widgetOrder as item (item.id)}
        <div
            class="{item.width} h-full relative transition-all duration-300
            {item.id === 'sidebar' && !showRightSidebar
                ? '!w-0 overflow-hidden'
                : ''}
            {isEditing ? 'cursor-move' : ''}
            {draggingItem === item ? 'opacity-50' : ''}
            {hoveringItem === item
                ? 'border-2 border-[var(--brand-primary)]'
                : ''}"
            draggable={isEditing}
            on:dragstart={(e) => handleDragStart(e, item)}
            on:dragover={(e) => handleDragOver(e, item)}
            on:drop={(e) => handleDrop(e, item)}
            on:dragend={handleDragEnd}
            animate:flip={{ duration: 300 }}
        >
            <!-- Render Component based on ID -->
            {#if item.id === "nodes"}
                <GuildsWidget />
            {:else if item.id === "channels"}
                <div class="flex flex-col gap-2 h-full">
                    <div class="flex-1 min-h-0 relative">
                        <ChannelsWidget activeServerId={activePeerId} />
                    </div>
                    {#if showProfile}
                        <div class="h-auto shrink-0 relative">
                            <ProfileWidget />
                        </div>
                    {/if}
                </div>
            {:else if item.id === "chat"}
                <div class="h-full w-full relative">
                    {#if chatTargetId}
                        <ChatWidget activePeerId={chatTargetId} />
                    {:else}
                        <div
                            class="h-full w-full flex items-center justify-center text-[var(--text-muted)] bg-[var(--background-primary)]/50 border border-[var(--border-subtle)] rounded-sm"
                        >
                            <div class="text-center">
                                <div class="text-lg font-bold">
                                    Welcome to Bonfire
                                </div>
                                <div class="text-sm">
                                    Select a server or friend to start chatting.
                                </div>
                            </div>
                        </div>
                    {/if}
                </div>
            {:else if item.id === "sidebar"}
                <div class="w-72 h-full">
                    {#if isHome}
                        <GameTrackerWidget />
                    {:else}
                        <MembersWidget />
                    {/if}
                </div>
            {/if}

            <!-- Editing Overlay -->
            {#if isEditing}
                <div
                    class="absolute inset-0 border-2 border-dashed border-[var(--brand-primary)] bg-[var(--brand-primary)]/10 z-50 pointer-events-none flex items-center justify-center"
                >
                    <span
                        class="bg-[var(--brand-primary)] text-white text-[10px] font-bold px-1 rounded-sm uppercase"
                    >
                        {item.id}
                    </span>
                </div>
            {/if}
        </div>
    {/each}
</div>

<!-- Edit Layout Controls -->
<div class="fixed bottom-4 right-4 z-[100] flex flex-col items-end gap-2">
    {#if isEditing}
        <div
            class="bg-[var(--background-primary)] border border-[var(--border-subtle)] p-3 rounded-sm shadow-2xl mb-2 flex flex-col gap-2 w-48"
            transition:fly={{ y: 20, duration: 200 }}
        >
            <div
                class="text-xs font-bold uppercase text-[var(--text-muted)] pb-1 border-b border-[var(--border-subtle)]"
            >
                Widget Visibility
            </div>

            <label
                class="flex items-center justify-between cursor-pointer group"
            >
                <span class="text-sm text-[var(--text-normal)]"
                    >Right Sidebar</span
                >
                <input
                    type="checkbox"
                    bind:checked={showRightSidebar}
                    class="accent-[var(--brand-primary)]"
                />
            </label>

            <label
                class="flex items-center justify-between cursor-pointer group"
            >
                <span class="text-sm text-[var(--text-normal)]">Profile</span>
                <input
                    type="checkbox"
                    bind:checked={showProfile}
                    class="accent-[var(--brand-primary)]"
                />
            </label>

            <div class="text-[10px] text-[var(--text-muted)] mt-2">
                Drag & Drop coming soon™
            </div>
        </div>
    {/if}

    <button
        class="h-10 px-4 bg-[var(--brand-primary)] hover:bg-[var(--brand-hover)] text-white rounded-sm shadow-lg font-bold text-xs uppercase tracking-wider flex items-center gap-2 transition-all"
        on:click={() => (isEditing = !isEditing)}
    >
        {#if isEditing}
            <svg
                width="16"
                height="16"
                viewBox="0 0 24 24"
                fill="none"
                stroke="currentColor"
                stroke-width="3"
                stroke-linecap="round"
                stroke-linejoin="round"
                ><polyline points="20 6 9 17 4 12"></polyline></svg
            >
            Done
        {:else}
            <svg
                width="16"
                height="16"
                viewBox="0 0 24 24"
                fill="none"
                stroke="currentColor"
                stroke-width="2"
                stroke-linecap="round"
                stroke-linejoin="round"
                ><path
                    d="M11 4H4a2 2 0 0 0-2 2v14a2 2 0 0 0 2 2h14a2 2 0 0 0 2-2v-7"
                ></path><path
                    d="M18.5 2.5a2.121 2.121 0 0 1 3 3L12 15l-4 1 1-4 9.5-9.5z"
                ></path></svg
            >
            Edit Layout
        {/if}
    </button>
</div>

<style>
    /* Global overrides for the new aesthetic if needed */
    :global(body) {
        background-color: var(--background-tertiary);
    }
</style>

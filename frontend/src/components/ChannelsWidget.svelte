<script>
    import { onMount } from "svelte";
    import { appContext, auth, modal, voice } from "../stores";
    import {
        GetDirectMessages,
        GetPeerChannels,
        DeleteChannel,
    } from "../../wailsjs/go/main/App.js";
    import { EventsOn } from "../../wailsjs/runtime/runtime";
    import WidgetContainer from "./WidgetContainer.svelte";
    import AddChannelModal from "./AddChannelModal.svelte";
    import EditChannelModal from "./EditChannelModal.svelte";
    import { webRTCService } from "../utils/webrtc";

    export let activeServerId = null; // null = Home

    let directMessages = [];
    let textChannels = [];
    let voiceChannels = [];
    let serverName = "Direct Messages";
    let isOwner = false;

    $: activeServerId, fetchContent();

    async function fetchContent() {
        if (!activeServerId || activeServerId === "home") {
            // Home Mode: Fetch DMs
            serverName = "Direct Messages";
            isOwner = false;
            try {
                const dms = await GetDirectMessages();
                directMessages = dms || [];
            } catch (e) {
                console.error("Failed to fetch DMs:", e);
            }
        } else {
            // Server Mode: Fetch Channels
            try {
                const server = $appContext.peers.find(
                    (p) => p.user_id === activeServerId,
                );
                // Check if local user is the owner
                if (
                    $auth.currentUser &&
                    $auth.currentUser.id === activeServerId
                ) {
                    serverName = $auth.currentUser.username || "My Server";
                    isOwner = true;
                } else {
                    serverName = server
                        ? server.server_name || server.username
                        : "Server";
                    isOwner = false;
                }

                const channels = await GetPeerChannels(activeServerId);
                textChannels = (channels || []).filter(
                    (c) => c.type === "text" || !c.type,
                );
                voiceChannels = (channels || []).filter(
                    (c) => c.type === "voice",
                );
            } catch (e) {
                console.error("Failed to fetch channels:", e);
            }
        }
    }

    function handleDMClick(dm) {
        appContext.setActiveChannel(dm.user_id);
    }

    function handleChannelClick(channel) {
        if (channel.type === "voice") {
            webRTCService.joinChannel(channel.channel_id);
            voice.update((v) => ({
                ...v,
                channelId: channel.channel_id,
                connected: true,
            }));
        } else {
            appContext.setActiveChannel(channel.channel_id);
        }
    }

    function openCreateChannelModal() {
        modal.open(AddChannelModal, {});
    }

    let contextMenu = null; // { x, y, channel }

    function handleContextMenu(e, channel) {
        if (!isOwner) return;
        e.preventDefault();
        contextMenu = {
            x: e.clientX,
            y: e.clientY,
            channel: channel,
        };
    }

    function closeContextMenu() {
        contextMenu = null;
    }

    function handleEditChannel() {
        if (contextMenu) {
            modal.open(EditChannelModal, { channel: contextMenu.channel });
            closeContextMenu();
        }
    }

    async function handleDeleteChannel() {
        if (contextMenu) {
            const channelId = contextMenu.channel.channel_id;
            if (confirm(`Are you sure you want to delete ${channelId}?`)) {
                await DeleteChannel(channelId);
                // Optimistic update or wait for event?
                // Event will trigger refresh
            }
            closeContextMenu();
        }
    }

    onMount(() => {
        fetchContent();
        document.addEventListener("click", closeContextMenu);

        EventsOn("CHANNEL_METADATA", () => fetchContent());
        EventsOn("CHANNEL_DELETED", () => fetchContent());

        return () => {
            document.removeEventListener("click", closeContextMenu);
        };
    });
</script>

<WidgetContainer title={serverName} className="h-full">
    <div class="flex flex-col h-full">
        <div class="flex-1 overflow-y-auto p-1 space-y-1 custom-scrollbar">
            {#if !activeServerId || activeServerId === "home"}
                <!-- DM List -->
                {#each directMessages as dm}
                    {@const isActive =
                        $appContext.activeChannelId === dm.user_id}
                    <button
                        class="w-full flex items-center gap-2 px-2 py-1 rounded-sm transition-colors group
                        {isActive
                            ? 'bg-[var(--background-modifier-selected)] text-[var(--interactive-active)]'
                            : 'text-[var(--channels-default)] hover:bg-[var(--background-modifier-hover)] hover:text-[var(--interactive-hover)]'}"
                        on:click={() => handleDMClick(dm)}
                    >
                        <!-- Avatar -->
                        <div class="relative w-6 h-6 shrink-0">
                            {#if dm.avatar}
                                <img
                                    src={dm.avatar}
                                    alt={dm.username}
                                    class="w-full h-full rounded-sm object-cover"
                                />
                            {:else}
                                <div
                                    class="w-full h-full rounded-sm bg-[var(--background-accent)] flex items-center justify-center text-xs font-bold text-white"
                                >
                                    {dm.username[0].toUpperCase()}
                                </div>
                            {/if}
                            <!-- Status -->
                            <div
                                class="absolute -bottom-0.5 -right-0.5 w-2.5 h-2.5 rounded-full border-2 border-[var(--background-secondary)]
                                {dm.status === 'online'
                                    ? 'bg-[var(--status-online)]'
                                    : dm.status === 'idle'
                                      ? 'bg-[var(--status-idle)]'
                                      : dm.status === 'dnd'
                                        ? 'bg-[var(--status-dnd)]'
                                        : 'bg-[var(--status-offline)]'}"
                            ></div>
                        </div>

                        <div
                            class="flex-1 text-left truncate font-medium text-sm"
                        >
                            {dm.username}
                        </div>
                    </button>
                {/each}
            {:else}
                <!-- Server Channels -->
                <!-- Text Channels -->
                {#if textChannels.length > 0}
                    <div
                        class="px-2 pt-2 pb-1 text-[10px] font-bold text-[var(--channels-default)] uppercase tracking-wider opacity-70"
                    >
                        Text Channels
                    </div>
                    {#each textChannels as channel}
                        {@const isActive =
                            $appContext.activeChannelId === channel.channel_id}
                        <button
                            class="w-full flex items-center gap-1.5 px-2 py-0.5 rounded-sm transition-colors group
                            {isActive
                                ? 'bg-[var(--background-modifier-selected)] text-[var(--interactive-active)]'
                                : 'text-[var(--channels-default)] hover:bg-[var(--background-modifier-hover)] hover:text-[var(--interactive-hover)]'}"
                            on:click={() => handleChannelClick(channel)}
                            on:contextmenu={(e) =>
                                handleContextMenu(e, channel)}
                        >
                            <span
                                class="text-lg text-[var(--text-muted)] leading-none"
                                >#</span
                            >
                            <span class="truncate font-medium text-sm"
                                >{channel.channel_id.replace(/^#/, "")}</span
                            >
                        </button>
                    {/each}
                {/if}

                <!-- Voice Channels -->
                {#if voiceChannels.length > 0}
                    <div
                        class="px-2 pt-2 pb-1 text-[10px] font-bold text-[var(--channels-default)] uppercase tracking-wider opacity-70"
                    >
                        Voice Channels
                    </div>
                    {#each voiceChannels as channel}
                        <button
                            class="w-full flex items-center gap-1.5 px-2 py-1 rounded-sm transition-colors group text-[var(--channels-default)] hover:bg-[var(--background-modifier-hover)] hover:text-[var(--interactive-hover)]"
                            on:click={() => handleChannelClick(channel)}
                            on:contextmenu={(e) =>
                                handleContextMenu(e, channel)}
                        >
                            <svg
                                class="w-4 h-4 shrink-0"
                                viewBox="0 0 24 24"
                                fill="none"
                                stroke="currentColor"
                                stroke-width="2"
                                stroke-linecap="round"
                                stroke-linejoin="round"
                            >
                                <path
                                    d="M12 1a3 3 0 0 0-3 3v8a3 3 0 0 0 6 0V4a3 3 0 0 0-3-3z"
                                ></path>
                                <path d="M19 10v2a7 7 0 0 1-14 0v-2"></path>
                                <line x1="12" y1="19" x2="12" y2="23"></line>
                                <line x1="8" y1="23" x2="16" y2="23"></line>
                            </svg>
                            <span class="truncate font-medium text-sm"
                                >{channel.channel_id.replace(/^#/, "")}</span
                            >
                        </button>
                    {/each}
                {/if}
            {/if}
        </div>

        <!-- Footer Actions (Create Channel) -->
        {#if isOwner}
            <div
                class="p-2 border-t border-[var(--border-subtle)] bg-[var(--background-tertiary)]/30"
            >
                <button
                    class="w-full flex items-center justify-center gap-2 px-2 py-1.5 bg-[var(--background-modifier-hover)] hover:bg-[var(--background-modifier-selected)] text-[var(--text-normal)] text-xs font-bold rounded-sm transition-colors"
                    on:click={openCreateChannelModal}
                >
                    <svg
                        width="12"
                        height="12"
                        viewBox="0 0 24 24"
                        fill="none"
                        stroke="currentColor"
                        stroke-width="3"
                        stroke-linecap="round"
                        stroke-linejoin="round"
                    >
                        <line x1="12" y1="5" x2="12" y2="19"></line>
                        <line x1="5" y1="12" x2="19" y2="12"></line>
                    </svg>
                    Create Channel
                </button>
            </div>
        {/if}
    </div>

    {#if contextMenu}
        <div
            class="fixed z-50 bg-[var(--background-floating)] border border-[var(--border-subtle)] rounded-sm shadow-lg py-1 min-w-[160px]"
            style="top: {contextMenu.y}px; left: {contextMenu.x}px;"
            on:click|stopPropagation
            on:keydown={() => {}}
            role="menu"
            tabindex="-1"
        >
            <button
                class="w-full text-left px-3 py-1.5 text-xs text-[var(--text-normal)] hover:bg-[var(--brand-primary)] hover:text-white transition-colors"
                on:click={handleEditChannel}
            >
                Edit Channel
            </button>
            <div class="h-px bg-[var(--border-subtle)] my-1"></div>
            <button
                class="w-full text-left px-3 py-1.5 text-xs text-[var(--status-dnd)] hover:bg-[var(--status-dnd)] hover:text-white transition-colors"
                on:click={handleDeleteChannel}
            >
                Delete Channel
            </button>
        </div>
    {/if}
</WidgetContainer>

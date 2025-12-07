<script>
    import { onMount } from "svelte";
    import { appContext } from "../stores";
    import {
        GetDirectMessages,
        GetPeerChannels,
        JoinChannel,
    } from "../../wailsjs/go/main/App.js";

    export let activeServerId = null; // null = Home

    let directMessages = [];
    let textChannels = [];
    let voiceChannels = [];
    let serverName = "Server";

    $: activeServerId, fetchContent();

    async function fetchContent() {
        if (!activeServerId) {
            // Home Mode: Fetch DMs
            try {
                const dms = await GetDirectMessages();
                directMessages = dms || [];
            } catch (e) {
                console.error("Failed to fetch DMs:", e);
            }
        } else {
            // Server Mode: Fetch Channels
            try {
                // We also need server details.
                // We can get it from appContext.peers if available
                const server = $appContext.peers.find(
                    (p) => p.user_id === activeServerId,
                );
                serverName = server
                    ? server.server_name || server.username
                    : "Server";

                const channels = await GetPeerChannels(activeServerId);
                textChannels = (channels || []).filter(
                    (c) => c.type === "text" || !c.type,
                ); // Default to text
                voiceChannels = (channels || []).filter(
                    (c) => c.type === "voice",
                );
            } catch (e) {
                console.error("Failed to fetch channels:", e);
            }
        }
    }

    function handleDMClick(dm) {
        // Set active channel to DM peer ID
        // We need to ensure we are in Home mode (which we are)
        // And set the "channel" to the peer ID.
        // App.svelte will handle routing.
        appContext.setActiveChannel(dm.user_id);
    }

    function handleChannelClick(channel) {
        appContext.setActiveChannel(channel.channel_id);
    }

    onMount(fetchContent);
</script>

<div
    class="w-60 h-full flex flex-col bg-[var(--background-secondary)] border-r border-[var(--border-subtle)]"
>
    <!-- Header -->
    <div
        class="h-12 flex items-center px-4 border-b border-[var(--border-subtle)] shadow-sm"
    >
        <h2 class="font-bold text-[var(--header-primary)] truncate">
            {activeServerId ? serverName : "Direct Messages"}
        </h2>
    </div>

    <!-- Content -->
    <div class="flex-1 overflow-y-auto p-2 space-y-4 custom-scrollbar">
        {#if !activeServerId}
            <!-- DM List -->
            <div class="space-y-0.5">
                {#each directMessages as dm}
                    {@const isActive =
                        $appContext.activeChannelId === dm.user_id}
                    <button
                        class="w-full flex items-center gap-3 px-2 py-1 transition-colors group
                        {isActive
                            ? 'bg-[var(--background-modifier-selected)] text-[var(--interactive-active)]'
                            : 'text-[var(--channels-default)] hover:bg-[var(--background-modifier-hover)] hover:text-[var(--interactive-hover)]'}"
                        on:click={() => handleDMClick(dm)}
                    >
                        <!-- Avatar -->
                        <div class="relative w-8 h-8">
                            {#if dm.avatar}
                                <img
                                    src={dm.avatar}
                                    alt={dm.username}
                                    class="w-full h-full rounded-full object-cover"
                                />
                            {:else}
                                <div
                                    class="w-full h-full rounded-full bg-[var(--background-accent)] flex items-center justify-center text-xs font-bold text-white"
                                >
                                    {dm.username[0].toUpperCase()}
                                </div>
                            {/if}
                            <!-- Status -->
                            <div
                                class="absolute bottom-0 right-0 w-2.5 h-2.5 rounded-full border-2 border-[var(--background-secondary)]
                                {dm.status === 'online'
                                    ? 'bg-green-500'
                                    : dm.status === 'idle'
                                      ? 'bg-yellow-500'
                                      : dm.status === 'dnd'
                                        ? 'bg-red-500'
                                        : 'bg-gray-500'}"
                            ></div>
                        </div>

                        <div class="flex-1 text-left truncate font-medium">
                            {dm.username}
                        </div>
                    </button>
                {/each}
            </div>
        {:else}
            <!-- Server Channels -->
            <!-- Text Channels -->
            {#if textChannels.length > 0}
                <div class="space-y-0.5">
                    <div
                        class="flex items-center justify-between px-2 pt-2 pb-1 text-xs font-bold text-[var(--channels-default)] uppercase hover:text-[var(--interactive-hover)] cursor-pointer group"
                    >
                        <span>Text Channels</span>
                        <svg
                            class="w-3 h-3 transition-transform group-hover:text-[var(--interactive-active)]"
                            viewBox="0 0 24 24"
                            fill="none"
                            stroke="currentColor"
                            stroke-width="2"
                            stroke-linecap="round"
                            stroke-linejoin="round"
                        >
                            <polyline points="6 9 12 15 18 9"></polyline>
                        </svg>
                    </div>
                    {#each textChannels as channel}
                        {@const isActive =
                            $appContext.activeChannelId === channel.channel_id}
                        <button
                            class="w-full flex items-center gap-1.5 px-2 py-0.5 transition-colors group
                            {isActive
                                ? 'bg-[var(--background-modifier-selected)] text-[var(--interactive-active)]'
                                : 'text-[var(--channels-default)] hover:bg-[var(--background-modifier-hover)] hover:text-[var(--interactive-hover)]'}"
                            on:click={() => handleChannelClick(channel)}
                        >
                            <span class="text-lg text-[var(--text-muted)]"
                                >#</span
                            >
                            <span class="truncate font-medium"
                                >{channel.channel_id.replace(/^#/, "")}</span
                            >
                        </button>
                    {/each}
                </div>
            {/if}

            <!-- Voice Channels -->
            {#if voiceChannels.length > 0}
                <div class="space-y-0.5 pt-2">
                    <div
                        class="flex items-center justify-between px-2 pt-2 pb-1 text-xs font-bold text-[var(--channels-default)] uppercase hover:text-[var(--interactive-hover)] cursor-pointer group"
                    >
                        <span>Voice Channels</span>
                        <svg
                            class="w-3 h-3 transition-transform group-hover:text-[var(--interactive-active)]"
                            viewBox="0 0 24 24"
                            fill="none"
                            stroke="currentColor"
                            stroke-width="2"
                            stroke-linecap="round"
                            stroke-linejoin="round"
                        >
                            <polyline points="6 9 12 15 18 9"></polyline>
                        </svg>
                    </div>
                    {#each voiceChannels as channel}
                        <button
                            class="w-full flex items-center gap-1.5 px-2 py-1.5 rounded transition-colors group text-[var(--channels-default)] hover:bg-[var(--background-modifier-hover)] hover:text-[var(--interactive-hover)]"
                            on:click={() => handleChannelClick(channel)}
                        >
                            <svg
                                class="w-4 h-4"
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
                            <span class="truncate font-medium"
                                >{channel.channel_id.replace(/^#/, "")}</span
                            >
                        </button>
                    {/each}
                </div>
            {/if}
        {/if}
    </div>
</div>

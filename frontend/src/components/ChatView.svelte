<script>
    import { appContext, appearance, auth } from "../stores";
    import {
        SendMessage,
        GetMessages,
        GetChannelPeers,
        RequestProfile,
    } from "../../wailsjs/go/main/App.js";
    import { EventsOn, EventsOff } from "../../wailsjs/runtime/runtime.js";
    import { onMount, onDestroy } from "svelte";
    import ChannelSettingsModal from "./ChannelSettingsModal.svelte";
    import MessageItem from "./MessageItem.svelte";

    export let activePeerId;

    let messageInput = "";
    let chatContainer;
    let channelMetadata = null;
    let showSettingsModal = false;
    let showMemberList = false;
    let channelMembers = [];

    let messages = [];

    // Helper to check grouping
    function isGroupStart(msg, prevMsg) {
        if (!prevMsg) return true;
        if (msg.sender !== prevMsg.sender) return true;
        const timeDiff = msg.timestamp - prevMsg.timestamp;
        return timeDiff > 5 * 60 * 1000; // 5 minutes
    }

    // Load messages and members when activePeerId changes
    $: if (activePeerId) {
        channelMetadata = null; // Reset metadata on switch
        messages = []; // Clear messages immediately
        loadMessages();
        if (activePeerId.startsWith("#")) {
            loadChannelMembers();
        } else {
            channelMembers = [];
        }
    }

    async function loadChannelMembers() {
        try {
            const peers = await GetChannelPeers(activePeerId);
            channelMembers = peers;
        } catch (e) {
            console.error("Failed to load channel members:", e);
        }
    }

    async function loadMessages() {
        if (!activePeerId) return;
        try {
            const msgs = await GetMessages(activePeerId);
            // Convert store.Message to UI format
            messages = (msgs || []).map((m) => ({
                id: m.id,
                sender: m.sender_id,
                senderUsername: m.sender_username, // Use stored username
                content: m.content,
                timestamp: new Date(m.timestamp),
            }));
            scrollToBottom();
        } catch (e) {
            console.error("Failed to load messages:", e);
        }
    }

    // Subscribe to P2P messages and events
    onMount(() => {
        EventsOn("P2P_MESSAGE", (msg) => {
            console.log("Received P2P_MESSAGE:", msg);
            console.log("Current activePeerId:", activePeerId);

            // msg is IncomingMessage struct from Go
            const newMsg = {
                id: Date.now(),
                sender: msg.senderId,
                senderUsername: msg.senderUsername, // Use live username
                content: msg.content,
                timestamp: new Date(),
            };
            // Check if message belongs to current view
            const isChannelMessage = !!msg.channelId;
            const isRelevant = isChannelMessage
                ? msg.channelId === activePeerId
                : msg.senderId === activePeerId;

            if (isRelevant) {
                console.log("Adding message to chat");
                messages = [...messages, newMsg];
                scrollToBottom();

                // If username is just ID, request profile
                if (
                    !msg.senderUsername ||
                    msg.senderUsername === msg.senderId
                ) {
                    RequestProfile(msg.senderId);
                }
            } else {
                console.log("Message ignored: mismatch", {
                    isChannelMessage,
                    activePeerId,
                    msgChannel: msg.channelId,
                    msgSender: msg.senderId,
                });
            }
        });

        EventsOn("P2P_EVENT", (event) => {
            console.log("Received P2P_EVENT:", event);
            if (event.type === "CHANNEL_METADATA") {
                const meta = event.payload;
                if (meta.channel_id === activePeerId) {
                    channelMetadata = meta;
                }
            }
        });

        EventsOn("PEER_UPDATED", (peer) => {
            // Update messages from this peer
            messages = messages.map((m) => {
                if (m.sender === peer.user_id) {
                    return { ...m, senderUsername: peer.username };
                }
                return m;
            });

            // Update member list if present
            channelMembers = channelMembers.map((m) => {
                if (m.user_id === peer.user_id) {
                    return {
                        ...m,
                        username: peer.username,
                        avatar: peer.avatar_url,
                        status: peer.status,
                    };
                }
                return m;
            });
        });
    });

    onDestroy(() => {
        EventsOff("P2P_MESSAGE");
        EventsOff("P2P_EVENT");
        EventsOff("PEER_UPDATED");
    });

    function scrollToBottom() {
        if (chatContainer) {
            setTimeout(() => {
                chatContainer.scrollTop = chatContainer.scrollHeight;
            }, 50);
        }
    }

    function handleKeydown(e) {
        if (e.key === "Enter" && !e.shiftKey) {
            e.preventDefault();
            sendMessage();
        }
    }

    async function sendMessage() {
        if (!messageInput.trim() || !activePeerId) return;

        const content = messageInput;
        messageInput = ""; // Clear input immediately

        // Optimistic update
        const newMsg = {
            id: Date.now(),
            sender: $auth.currentUser.id, // Me
            senderUsername: $auth.currentUser.username, // Me
            content: content,
            timestamp: new Date(),
        };
        messages = [...messages, newMsg];
        scrollToBottom();

        try {
            const success = await SendMessage(activePeerId, content);
            if (!success) {
                console.error("Failed to send message");
                // TODO: Show error state
            }
        } catch (e) {
            console.error("Error sending message:", e);
        }
    }

    function formatTime(date) {
        return date.toLocaleTimeString([], {
            hour: "2-digit",
            minute: "2-digit",
        });
    }

    // Helper to get display name
    function getDisplayName(msg) {
        if (msg.sender === $auth.currentUser.id) return "You";
        if (msg.senderUsername) return msg.senderUsername;
        // Fallback to peer list lookup
        const peer = $appContext.peers.find((p) => p.user_id === msg.sender);
        return peer ? peer.username : "Unknown";
    }
</script>

<div
    class="flex flex-col h-full bg-[var(--background-primary)] text-[var(--text-normal)] font-mono"
>
    <!-- Header -->
    <div
        class="h-12 flex items-center px-6 border-b border-[var(--border-subtle)] bg-[var(--background-primary)]/95 backdrop-blur justify-between"
    >
        <div class="flex items-center gap-3">
            <span class="text-[var(--text-muted)]">@</span>
            <div class="flex flex-col">
                <span class="font-bold text-lg tracking-tight leading-none">
                    {$appContext.peers.find((p) => p.user_id === activePeerId)
                        ?.username ||
                        activePeerId ||
                        "Unknown"}
                </span>
                {#if channelMetadata && channelMetadata.topic}
                    <span
                        class="text-xs text-[var(--text-muted)] leading-none mt-0.5"
                    >
                        {channelMetadata.topic}
                    </span>
                {/if}
            </div>

            {#if channelMetadata && channelMetadata.owner_id === $auth.currentUser?.id}
                <button
                    class="ml-2 p-1 text-[var(--text-muted)] hover:text-[var(--text-normal)] transition-colors opacity-50 hover:opacity-100"
                    on:click={() => (showSettingsModal = true)}
                    title="Edit Channel Settings"
                >
                    <svg
                        xmlns="http://www.w3.org/2000/svg"
                        width="14"
                        height="14"
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
                </button>
            {/if}

            {#if channelMetadata && channelMetadata.is_modded}
                <div
                    class="flex items-center gap-2 px-2 py-1 bg-yellow-500/10 border border-yellow-500/20 rounded text-yellow-500 text-xs font-bold uppercase tracking-wider"
                    title={`Hosted by ${channelMetadata.software_agent}`}
                >
                    <svg
                        xmlns="http://www.w3.org/2000/svg"
                        width="12"
                        height="12"
                        viewBox="0 0 24 24"
                        fill="none"
                        stroke="currentColor"
                        stroke-width="3"
                        stroke-linecap="round"
                        stroke-linejoin="round"
                        ><path
                            d="M10.29 3.86L1.82 18a2 2 0 0 0 1.71 3h16.94a2 2 0 0 0 1.71-3L13.71 3.86a2 2 0 0 0-3.42 0z"
                        ></path><line x1="12" y1="9" x2="12" y2="13"
                        ></line><line x1="12" y1="17" x2="12.01" y2="17"
                        ></line></svg
                    >
                    Modded Server
                </div>
            {:else}
                <span
                    class="px-2 py-0.5 text-[10px] uppercase tracking-wider font-bold bg-green-500/10 text-green-500 rounded-sm"
                    >Online</span
                >
            {/if}
        </div>

        <div class="flex items-center gap-2">
            {#if activePeerId.startsWith("#")}
                <button
                    class="p-2 text-[var(--text-muted)] hover:text-[var(--text-normal)] transition-colors rounded hover:bg-[var(--background-modifier-hover)] {showMemberList
                        ? 'text-[var(--text-normal)] bg-[var(--background-modifier-selected)]'
                        : ''}"
                    on:click={() => (showMemberList = !showMemberList)}
                    title="Member List"
                >
                    <svg
                        xmlns="http://www.w3.org/2000/svg"
                        width="20"
                        height="20"
                        viewBox="0 0 24 24"
                        fill="none"
                        stroke="currentColor"
                        stroke-width="2"
                        stroke-linecap="round"
                        stroke-linejoin="round"
                        ><path d="M17 21v-2a4 4 0 0 0-4-4H5a4 4 0 0 0-4 4v2"
                        ></path><circle cx="9" cy="7" r="4"></circle><path
                            d="M23 21v-2a4 4 0 0 0-3-3.87"
                        ></path><path d="M16 3.13a4 4 0 0 1 0 7.75"></path></svg
                    >
                </button>
            {/if}
        </div>
    </div>

    {#if channelMetadata && channelMetadata.is_modded}
        <div
            class="bg-yellow-500/10 border-b border-yellow-500/20 px-6 py-2 text-xs text-yellow-200 flex items-center justify-between"
        >
            <span>
                <strong>Warning:</strong> This channel is hosted by a modified
                client ({channelMetadata.software_agent}). Features may behave
                unexpectedly.
            </span>
        </div>
    {/if}

    <div class="flex flex-1 overflow-hidden">
        <!-- Messages Area -->
        <div class="flex-1 flex flex-col min-w-0">
            <div
                bind:this={chatContainer}
                class="flex-1 overflow-y-auto p-6 space-y-0.5 scrollbar-thin scrollbar-thumb-[var(--background-secondary)] scrollbar-track-transparent"
            >
                {#each messages as msg, i}
                    <MessageItem
                        message={msg}
                        isGroupStart={isGroupStart(msg, messages[i - 1])}
                    />
                {/each}
            </div>

            <!-- Input Area -->
            <div class="p-6 pt-2 pl-[270px] pr-6">
                <div
                    class="relative bg-[var(--background-secondary)] border border-[var(--border-subtle)] rounded-lg transition-colors focus-within:border-[var(--brand-primary)] focus-within:ring-1 focus-within:ring-[var(--brand-primary)]"
                >
                    <input
                        type="text"
                        bind:value={messageInput}
                        on:keydown={handleKeydown}
                        placeholder={`Message @${
                            $appContext.peers.find(
                                (p) => p.user_id === activePeerId,
                            )?.username ||
                            activePeerId ||
                            "peer"
                        }...`}
                        class="w-full bg-transparent px-4 py-3 outline-none text-[var(--text-normal)] placeholder-[var(--text-muted)] font-mono text-sm"
                        autocomplete="off"
                    />
                    <div
                        class="absolute right-3 top-3 text-[var(--text-muted)] text-xs opacity-50"
                    >
                        RET
                    </div>
                </div>
            </div>
        </div>

        <!-- Member List Sidebar -->
        {#if showMemberList && activePeerId.startsWith("#")}
            <div
                class="w-60 bg-[var(--background-secondary)] border-l border-[var(--border-subtle)] flex flex-col"
            >
                <div class="p-4 border-b border-[var(--border-subtle)]">
                    <h3
                        class="text-xs font-bold text-[var(--text-muted)] uppercase tracking-wider"
                    >
                        Members — {channelMembers.length}
                    </h3>
                </div>
                <div
                    class="flex-1 overflow-y-auto p-2 space-y-1 scrollbar-thin"
                >
                    {#each channelMembers as member}
                        <div
                            class="flex items-center gap-2 px-2 py-1.5 rounded hover:bg-[var(--background-modifier-hover)] cursor-pointer group"
                        >
                            <div
                                class="w-8 h-8 rounded-full bg-[var(--background-tertiary)] overflow-hidden shrink-0"
                            >
                                {#if member.avatar}
                                    <img
                                        src={member.avatar}
                                        alt={member.username}
                                        class="w-full h-full object-cover"
                                    />
                                {:else}
                                    <div
                                        class="w-full h-full flex items-center justify-center text-xs font-bold text-[var(--text-muted)]"
                                    >
                                        {member.username[0].toUpperCase()}
                                    </div>
                                {/if}
                            </div>
                            <div class="flex flex-col min-w-0">
                                <div class="flex items-center gap-1">
                                    <span
                                        class="text-sm font-medium text-[var(--text-normal)] truncate"
                                    >
                                        {member.username}
                                    </span>
                                    {#if channelMetadata && member.user_id === channelMetadata.owner_id}
                                        <svg
                                            xmlns="http://www.w3.org/2000/svg"
                                            width="12"
                                            height="12"
                                            viewBox="0 0 24 24"
                                            fill="none"
                                            stroke="currentColor"
                                            stroke-width="2"
                                            stroke-linecap="round"
                                            stroke-linejoin="round"
                                            class="text-yellow-500"
                                            ><path
                                                d="M12 2l3.09 6.26L22 9.27l-5 4.87 1.18 6.88L12 17.77l-6.18 3.25L7 14.14 2 9.27l6.91-1.01L12 2z"
                                            ></path></svg
                                        >
                                    {/if}
                                </div>
                                {#if member.status}
                                    <span
                                        class="text-[10px] text-[var(--text-muted)] truncate"
                                    >
                                        {member.status_message || member.status}
                                    </span>
                                {/if}
                            </div>
                        </div>
                    {/each}
                </div>
            </div>
        {/if}
    </div>

    {#if showSettingsModal && channelMetadata}
        <ChannelSettingsModal
            channelName={activePeerId}
            currentTopic={channelMetadata.topic}
            isPrivate={channelMetadata.is_private}
            on:close={() => (showSettingsModal = false)}
        />
    {/if}
</div>

<style>
    /* Custom Scrollbar for Webkit */
    .scrollbar-thin::-webkit-scrollbar {
        width: 6px;
    }
    .scrollbar-thin::-webkit-scrollbar-track {
        background: transparent;
    }
    .scrollbar-thin::-webkit-scrollbar-thumb {
        background-color: var(--background-secondary);
        border-radius: 3px;
    }
</style>

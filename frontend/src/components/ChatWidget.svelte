<script>
    import { appContext, appearance, auth } from "../stores";
    import {
        SendMessage,
        GetMessages,
        GetChannelPeers,
    } from "../../wailsjs/go/main/App.js";
    import { EventsOn, EventsOff } from "../../wailsjs/runtime/runtime.js";
    import { onMount, onDestroy } from "svelte";
    import ChannelSettingsModal from "./ChannelSettingsModal.svelte";
    import MessageItem from "./MessageItem.svelte";
    import WidgetContainer from "./WidgetContainer.svelte";

    export let activePeerId;

    let messageInput = "";
    let chatContainer;
    let channelMetadata = null;
    let showSettingsModal = false;
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
                messages = [...messages, newMsg];
                scrollToBottom();
            }
        });

        EventsOn("P2P_EVENT", (event) => {
            if (event.type === "CHANNEL_METADATA") {
                const meta = event.payload;
                if (meta.channel_id === activePeerId) {
                    channelMetadata = meta;
                }
            }
        });
    });

    onDestroy(() => {
        EventsOff("P2P_MESSAGE");
        EventsOff("P2P_EVENT");
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
            }
        } catch (e) {
            console.error("Error sending message:", e);
        }
    }

    $: chatTitle =
        $appContext.peers.find((p) => p.user_id === activePeerId)?.username ||
        activePeerId ||
        "Chat";
</script>

<WidgetContainer title={chatTitle} className="h-full">
    <div
        class="flex flex-col h-full bg-[var(--background-primary)]/50 text-[var(--text-normal)] font-mono"
    >
        <!-- Sub-Header (Topic/Controls) -->
        {#if (channelMetadata && channelMetadata.topic) || activePeerId.startsWith("#")}
            <div
                class="flex items-center justify-between px-4 py-1 border-b border-[var(--border-subtle)] bg-[var(--background-secondary)]/50 text-xs"
            >
                <div class="truncate opacity-70">
                    {#if channelMetadata && channelMetadata.topic}
                        {channelMetadata.topic}
                    {:else}
                        <span class="italic text-[10px]">No topic set</span>
                    {/if}
                </div>
                <div class="flex items-center gap-2">
                    {#if channelMetadata && channelMetadata.owner_id === $auth.currentUser?.id}
                        <button
                            class="hover:text-[var(--brand-primary)]"
                            on:click={() => (showSettingsModal = true)}
                            title="Settings"
                        >
                            <svg
                                width="12"
                                height="12"
                                viewBox="0 0 24 24"
                                fill="none"
                                stroke="currentColor"
                                stroke-width="2"
                                stroke-linecap="round"
                                stroke-linejoin="round"
                                ><circle cx="12" cy="12" r="3"></circle><path
                                    d="M19.4 15a1.65 1.65 0 0 0 .33 1.82l.06.06a2 2 0 0 1 0 2.83 2 2 0 0 1-2.83 0l-.06-.06a1.65 1.65 0 0 0-1.82-.33 1.65 1.65 0 0 0-1 1.51V21a2 2 0 0 1-2 2 2 2 0 0 1-2-2v-.09A1.65 1.65 0 0 0 9 19.4a1.65 1.65 0 0 0-1.82.33l-.06.06a2 2 0 0 1-2.83 0 2 2 0 0 1 0-2.83l.06-.06a1.65 1.65 0 0 0 .33-1.82 1.65 1.65 0 0 0-1.51-1H3a2 2 0 0 1-2-2 2 2 0 0 1 2-2h.09A1.65 1.65 0 0 0 4.6 9a1.65 1.65 0 0 0-.33-1.82l-.06-.06a2 2 0 0 1 0-2.83 2 2 0 0 1 2.83 0l.06.06a1.65 1.65 0 0 0 1.82.33H9a1.65 1.65 0 0 0 1-1.51V3a2 2 0 0 1 2-2 2 2 0 0 1 2 2v.09a1.65 1.65 0 0 0 1 1.51 1.65 1.65 0 0 0 1.82-.33l.06-.06a2 2 0 0 1 2.83 0 2 2 0 0 1 0 2.83l-.06.06a1.65 1.65 0 0 0-.33 1.82V9a1.65 1.65 0 0 0 1.51 1H21a2 2 0 0 1 2 2 2 2 0 0 1-2 2h-.09a1.65 1.65 0 0 0-1.51 1z"
                                ></path></svg
                            >
                        </button>
                    {/if}
                </div>
            </div>
        {/if}

        <div class="flex flex-1 overflow-hidden">
            <!-- Messages Area -->
            <div class="flex-1 flex flex-col min-w-0">
                <div
                    bind:this={chatContainer}
                    class="flex-1 overflow-y-auto p-4 space-y-0.5 scrollbar-thin widget-content"
                >
                    {#each messages as msg, i}
                        <MessageItem
                            message={msg}
                            isGroupStart={isGroupStart(msg, messages[i - 1])}
                        />
                    {/each}
                </div>

                <!-- Input Area -->
                <div class="p-4 pt-2">
                    <div
                        class="relative bg-[var(--background-secondary)] border border-[var(--border-subtle)] rounded-sm transition-colors focus-within:border-[var(--brand-primary)]"
                    >
                        <input
                            type="text"
                            bind:value={messageInput}
                            on:keydown={handleKeydown}
                            placeholder={`Message ${activePeerId}...`}
                            class="w-full bg-transparent px-3 py-2 outline-none text-[var(--text-normal)] placeholder-[var(--text-muted)] font-mono text-sm"
                            autocomplete="off"
                        />
                    </div>
                </div>
            </div>
        </div>
    </div>

    {#if showSettingsModal && channelMetadata}
        <ChannelSettingsModal
            channelName={activePeerId}
            currentTopic={channelMetadata.topic}
            isPrivate={channelMetadata.is_private}
            on:close={() => (showSettingsModal = false)}
        />
    {/if}
</WidgetContainer>

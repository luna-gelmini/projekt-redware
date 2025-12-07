<script>
    import { appContext, appearance, auth } from "../stores";
    import { renderMarkdown } from "../utils/markdown";

    export let message;
    export let isGroupStart = true; // If previous message was from same author within timeframe

    $: compact = $appearance.density === "compact";

    function formatTime(date) {
        return date.toLocaleTimeString([], {
            hour: "2-digit",
            minute: "2-digit",
        });
    }

    // Helper to get display name
    function getDisplayName(msg) {
        if (msg.sender === $auth.currentUser?.id)
            return $auth.currentUser.username;
        if (msg.senderUsername) return msg.senderUsername;
        // Fallback to peer list lookup
        const peer = $appContext.peers.find((p) => p.user_id === msg.sender);
        return peer ? peer.username : "Unknown";
    }

    $: displayName = getDisplayName(message);
    $: timestamp = formatTime(message.timestamp);
</script>

{#if compact}
    <div
        class="flex items-baseline gap-2 px-4 py-0.5 hover:bg-[var(--background-modifier-hover)] -mx-4 group"
    >
        <span
            class="text-[10px] text-[var(--text-muted)] w-[50px] text-right shrink-0 select-none opacity-0 group-hover:opacity-100 font-mono"
        >
            {timestamp}
        </span>
        <div class="flex-1 min-w-0 break-words leading-[1.375rem]">
            {#if isGroupStart}
                <span
                    class="font-bold text-[var(--header-primary)] mr-1 cursor-pointer hover:underline"
                >
                    {displayName}
                </span>
            {/if}
            <span class="text-[var(--text-normal)]">
                {message.content}
            </span>
        </div>
    </div>
{:else}
    <div
        class="group flex pr-4 hover:bg-[var(--background-message-hover)] {$appearance.messageDensity ===
        'compact'
            ? 'py-0.5'
            : 'py-1'} {isGroupStart ? 'mt-4' : ''}"
    >
        {#if isGroupStart}
            <!-- Avatar -->
            <div
                class="flex-shrink-0 cursor-pointer select-none {$appearance.messageDensity ===
                'compact'
                    ? 'w-12 text-right pr-3'
                    : 'w-16 pl-4'}"
            >
                {#if $appearance.messageDensity === "cozy"}
                    <div
                        class="w-10 h-10 rounded-full bg-[var(--background-secondary)] overflow-hidden"
                    >
                        <img
                            src={`http://localhost:3000/user/avatar/${message.sender}`}
                            alt={message.senderUsername}
                            class="w-full h-full object-cover"
                            on:error={(e) =>
                                (e.target.src = `https://api.dicebear.com/7.x/avataaars/svg?seed=${message.sender}`)}
                        />
                    </div>
                {:else}
                    <span
                        class="text-[10px] text-[var(--text-muted)] opacity-0 group-hover:opacity-100"
                    >
                        {formatTime(message.timestamp)}
                    </span>
                {/if}
            </div>
        {:else if $appearance.messageDensity === "compact"}
            <div class="w-12 text-right pr-3 flex-shrink-0 select-none">
                <span
                    class="text-[10px] text-[var(--text-muted)] opacity-0 group-hover:opacity-100"
                >
                    {formatTime(message.timestamp)}
                </span>
            </div>
        {:else}
            <div class="w-16 flex-shrink-0"></div>
        {/if}

        <div class="flex-1 min-w-0">
            {#if isGroupStart && $appearance.messageDensity === "cozy"}
                <div class="flex items-center gap-2 mb-1">
                    <span
                        class="font-medium text-[var(--header-primary)] hover:underline cursor-pointer"
                    >
                        {message.senderUsername || message.sender}
                    </span>
                    <span class="text-xs text-[var(--text-muted)] select-none">
                        {formatTime(message.timestamp)}
                    </span>
                </div>
            {/if}

            <div
                class="text-[var(--text-normal)] whitespace-pre-wrap break-words leading-[1.375rem] markdown-content"
            >
                {#if $appearance.messageDensity === "compact" && isGroupStart}
                    <span
                        class="font-medium text-[var(--header-primary)] hover:underline cursor-pointer mr-1"
                    >
                        {message.senderUsername || message.sender}
                    </span>
                {/if}
                {@html renderMarkdown(message.content)}
            </div>
        </div>
    </div>
{/if}

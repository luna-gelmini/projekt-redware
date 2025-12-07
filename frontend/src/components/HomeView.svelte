<script>
    import { appContext } from "../stores";
    import { onMount } from "svelte";

    let activeTab = "online"; // online, all, pending, blocked
    let searchQuery = "";

    $: peers = $appContext.peers || [];
    $: filteredPeers = peers.filter((p) => {
        if (searchQuery) {
            const q = searchQuery.toLowerCase();
            if (
                !p.username.toLowerCase().includes(q) &&
                !p.user_id.toLowerCase().includes(q)
            ) {
                return false;
            }
        }

        if (activeTab === "online") {
            return (
                p.status === "online" ||
                p.status === "dnd" ||
                p.status === "idle"
            );
        } else if (activeTab === "all") {
            return true;
        } else if (activeTab === "pending") {
            return false; // TODO: Implement Friend Requests
        } else if (activeTab === "blocked") {
            return false; // TODO: Implement Blocking
        }
        return true;
    });

    function handleChat(peer) {
        appContext.setActiveChannel(peer.user_id);
    }

    function handleCall(peer) {
        console.log("Call not implemented yet");
    }
</script>

<div class="flex flex-col h-full bg-[var(--background-primary)]">
    <!-- Header -->
    <div
        class="h-12 flex items-center px-4 border-b border-[var(--border-subtle)] gap-4"
    >
        <div
            class="flex items-center gap-2 text-[var(--header-primary)] font-bold"
        >
            <svg
                xmlns="http://www.w3.org/2000/svg"
                width="24"
                height="24"
                viewBox="0 0 24 24"
                fill="none"
                stroke="currentColor"
                stroke-width="2"
                stroke-linecap="round"
                stroke-linejoin="round"
                class="text-[var(--text-muted)]"
            >
                <path d="M20 21v-2a4 4 0 0 0-4-4H8a4 4 0 0 0-4 4v2"></path>
                <circle cx="12" cy="7" r="4"></circle>
            </svg>
            <span>Friends</span>
        </div>

        <div
            class="h-6 w-[1px] bg-[var(--background-modifier-accent)] mx-2"
        ></div>

        <div class="flex items-center gap-4">
            <button
                class="px-2 py-1 rounded hover:bg-[var(--background-modifier-hover)] transition-colors font-medium
                {activeTab === 'online'
                    ? 'text-[var(--header-primary)] bg-[var(--background-modifier-selected)]'
                    : 'text-[var(--text-muted)]'}"
                on:click={() => (activeTab = "online")}
            >
                Online
            </button>
            <button
                class="px-2 py-1 rounded hover:bg-[var(--background-modifier-hover)] transition-colors font-medium
                {activeTab === 'all'
                    ? 'text-[var(--header-primary)] bg-[var(--background-modifier-selected)]'
                    : 'text-[var(--text-muted)]'}"
                on:click={() => (activeTab = "all")}
            >
                All
            </button>
            <button
                class="px-2 py-1 rounded hover:bg-[var(--background-modifier-hover)] transition-colors font-medium
                {activeTab === 'pending'
                    ? 'text-[var(--header-primary)] bg-[var(--background-modifier-selected)]'
                    : 'text-[var(--text-muted)]'}"
                on:click={() => (activeTab = "pending")}
            >
                Pending
            </button>
            <button
                class="px-2 py-1 rounded hover:bg-[var(--background-modifier-hover)] transition-colors font-medium
                {activeTab === 'blocked'
                    ? 'text-[var(--header-primary)] bg-[var(--background-modifier-selected)]'
                    : 'text-[var(--text-muted)]'}"
                on:click={() => (activeTab = "blocked")}
            >
                Blocked
            </button>
            <button
                class="px-2 py-1 rounded bg-green-600 text-white hover:bg-green-700 transition-colors font-medium text-sm ml-4"
            >
                Add Friend
            </button>
        </div>
    </div>

    <!-- Search -->
    <div class="p-4 border-b border-[var(--border-subtle)]">
        <div class="relative">
            <input
                type="text"
                bind:value={searchQuery}
                placeholder="Search"
                class="w-full bg-[var(--background-tertiary)] text-[var(--text-normal)] rounded px-3 py-2 pl-10 focus:outline-none focus:ring-2 focus:ring-[var(--brand-primary)] transition-all"
            />
            <svg
                class="absolute left-3 top-2.5 w-5 h-5 text-[var(--text-muted)]"
                viewBox="0 0 24 24"
                fill="none"
                stroke="currentColor"
                stroke-width="2"
                stroke-linecap="round"
                stroke-linejoin="round"
            >
                <circle cx="11" cy="11" r="8"></circle>
                <line x1="21" y1="21" x2="16.65" y2="16.65"></line>
            </svg>
        </div>
    </div>

    <!-- List -->
    <div class="flex-1 overflow-y-auto p-4 custom-scrollbar">
        <div class="text-xs font-bold text-[var(--text-muted)] uppercase mb-4">
            {activeTab.toUpperCase()} — {filteredPeers.length}
        </div>

        <div class="space-y-2">
            {#each filteredPeers as peer}
                <div
                    class="group flex items-center justify-between p-3 rounded-lg hover:bg-[var(--background-modifier-hover)] border-t border-[var(--background-modifier-accent)] hover:border-transparent transition-all cursor-pointer"
                    on:click={() => handleChat(peer)}
                >
                    <div class="flex items-center gap-3">
                        <div class="relative w-10 h-10">
                            {#if peer.avatar}
                                <img
                                    src={peer.avatar}
                                    alt={peer.username}
                                    class="w-full h-full rounded-full object-cover"
                                />
                            {:else}
                                <div
                                    class="w-full h-full rounded-full bg-[var(--background-accent)] flex items-center justify-center text-sm font-bold text-white"
                                >
                                    {peer.username[0].toUpperCase()}
                                </div>
                            {/if}
                            <div
                                class="absolute bottom-0 right-0 w-3.5 h-3.5 rounded-full border-2 border-[var(--background-primary)]
                                {peer.status === 'online'
                                    ? 'bg-green-500'
                                    : peer.status === 'idle'
                                      ? 'bg-yellow-500'
                                      : peer.status === 'dnd'
                                        ? 'bg-red-500'
                                        : 'bg-gray-500'}"
                            ></div>
                        </div>
                        <div>
                            <div
                                class="font-bold text-[var(--header-primary)] flex items-center gap-2"
                            >
                                {peer.username}
                                <span
                                    class="text-xs text-[var(--text-muted)] font-normal hidden group-hover:inline"
                                    >#{peer.user_id.substring(0, 4)}</span
                                >
                            </div>
                            <div class="text-xs text-[var(--text-muted)]">
                                {peer.statusMessage || peer.status}
                            </div>
                        </div>
                    </div>

                    <div
                        class="flex items-center gap-2 opacity-0 group-hover:opacity-100 transition-opacity"
                    >
                        <button
                            class="w-9 h-9 rounded-full bg-[var(--background-secondary)] flex items-center justify-center text-[var(--text-muted)] hover:text-[var(--text-normal)] hover:bg-[var(--background-tertiary)] transition-colors"
                            title="Message"
                            on:click|stopPropagation={() => handleChat(peer)}
                        >
                            <svg
                                width="20"
                                height="20"
                                viewBox="0 0 24 24"
                                fill="none"
                                stroke="currentColor"
                                stroke-width="2"
                                stroke-linecap="round"
                                stroke-linejoin="round"
                            >
                                <path
                                    d="M21 11.5a8.38 8.38 0 0 1-.9 3.8 8.5 8.5 0 0 1-7.6 4.7 8.38 8.38 0 0 1-3.8-.9L3 21l1.9-5.7a8.38 8.38 0 0 1-.9-3.8 8.5 8.5 0 0 1 4.7-7.6 8.38 8.38 0 0 1 3.8-.9h.5a8.48 8.48 0 0 1 8 8v.5z"
                                ></path>
                            </svg>
                        </button>
                        <button
                            class="w-9 h-9 rounded-full bg-[var(--background-secondary)] flex items-center justify-center text-[var(--text-muted)] hover:text-[var(--text-normal)] hover:bg-[var(--background-tertiary)] transition-colors"
                            title="Call"
                            on:click|stopPropagation={() => handleCall(peer)}
                        >
                            <svg
                                width="20"
                                height="20"
                                viewBox="0 0 24 24"
                                fill="none"
                                stroke="currentColor"
                                stroke-width="2"
                                stroke-linecap="round"
                                stroke-linejoin="round"
                            >
                                <path
                                    d="M22 16.92v3a2 2 0 0 1-2.18 2 19.79 19.79 0 0 1-8.63-3.07 19.5 19.5 0 0 1-6-6 19.79 19.79 0 0 1-3.07-8.67A2 2 0 0 1 4.11 2h3a2 2 0 0 1 2 1.72 12.84 12.84 0 0 0 .7 2.81 2 2 0 0 1-.45 2.11L8.09 9.91a16 16 0 0 0 6 6l1.27-1.27a2 2 0 0 1 2.11-.45 12.84 12.84 0 0 0 2.81.7A2 2 0 0 1 22 16.92z"
                                ></path>
                            </svg>
                        </button>
                    </div>
                </div>
            {/each}
        </div>
    </div>
</div>

<script>
    import { fade, scale } from "svelte/transition";
    import { modal } from "../stores";

    export let user; // { username, avatar, status, statusMessage, id }

    function close() {
        modal.close();
    }
</script>

<div
    class="fixed inset-0 z-[100] flex items-center justify-center bg-black/60 backdrop-blur-sm"
    transition:fade={{ duration: 200 }}
    on:click={close}
    on:keydown={(e) => e.key === "Escape" && close()}
    role="button"
    tabindex="0"
>
    <div
        class="w-full max-w-sm bg-[var(--background-primary)] border border-[var(--border-subtle)] rounded-2xl shadow-2xl overflow-hidden transform transition-all"
        transition:scale={{ duration: 200, start: 0.95 }}
        on:click|stopPropagation
        on:keydown={() => {}}
        role="document"
        tabindex="-1"
    >
        <!-- Header / Banner -->
        <div
            class="h-24 bg-gradient-to-r from-indigo-500 to-purple-600 relative"
        >
            <button
                class="absolute top-2 right-2 p-2 text-white/80 hover:text-white bg-black/20 hover:bg-black/40 rounded-full transition-colors"
                on:click={close}
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
                    <line x1="18" y1="6" x2="6" y2="18"></line>
                    <line x1="6" y1="6" x2="18" y2="18"></line>
                </svg>
            </button>
        </div>

        <!-- Avatar & Content -->
        <div class="px-6 pb-6 relative flex flex-col">
            <!-- Avatar Container (Negative margin to pull it up) -->
            <div class="relative -mt-12 mb-4 self-start">
                <div
                    class="w-24 h-24 rounded-full border-4 border-[var(--background-primary)] bg-[var(--background-secondary)] overflow-hidden shadow-lg flex items-center justify-center"
                >
                    {#if user.avatar}
                        <img
                            src={user.avatar}
                            alt={user.username}
                            class="w-full h-full object-cover"
                        />
                    {:else}
                        <span
                            class="text-3xl font-bold text-[var(--text-muted)]"
                        >
                            {(user.username || "?")[0].toUpperCase()}
                        </span>
                    {/if}
                </div>
                <!-- Status Indicator -->
                <div
                    class="absolute bottom-1 right-1 w-6 h-6 rounded-full border-4 border-[var(--background-primary)]
                    {user.status === 'online'
                        ? 'bg-green-500'
                        : user.status === 'idle'
                          ? 'bg-yellow-500'
                          : user.status === 'dnd'
                            ? 'bg-red-500'
                            : 'bg-gray-500'}"
                ></div>
            </div>

            <!-- Info -->
            <div class="space-y-1">
                <h2 class="text-2xl font-bold text-[var(--header-primary)]">
                    {user.username}
                </h2>
                <p class="text-sm text-[var(--text-muted)] font-mono">
                    ID: {user.id}
                </p>
            </div>

            <!-- Status Message -->
            {#if user.statusMessage}
                <div
                    class="mt-4 p-3 bg-[var(--background-secondary)] rounded-lg border border-[var(--border-subtle)]"
                >
                    <p class="text-sm text-[var(--text-normal)] italic">
                        "{user.statusMessage}"
                    </p>
                </div>
            {/if}

            <!-- Actions (Placeholder) -->
            <div class="mt-6 flex gap-2">
                <button
                    class="flex-1 py-2 bg-[var(--brand-primary)] hover:bg-[var(--brand-hover)] text-white rounded-md font-medium transition-colors"
                    on:click={() => {
                        // TODO: Start DM
                        close();
                    }}
                >
                    Message
                </button>
            </div>
        </div>
    </div>
</div>

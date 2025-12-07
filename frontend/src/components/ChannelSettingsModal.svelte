<script>
    import { createEventDispatcher } from "svelte";
    import { UpdateChannelInfo } from "../../wailsjs/go/main/App.js";

    export let channelName = "";
    export let currentTopic = "";
    export let isPrivate = false;

    const dispatch = createEventDispatcher();

    let topic = currentTopic;
    let privateChannel = isPrivate;
    let error = "";

    async function handleSave() {
        if (!topic.trim()) {
            error = "Topic cannot be empty";
            return;
        }

        try {
            const success = await UpdateChannelInfo(
                channelName,
                topic,
                privateChannel,
            );
            if (success) {
                dispatch("close");
            } else {
                error = "Failed to update channel settings";
            }
        } catch (e) {
            error = e.message || "An error occurred";
        }
    }
</script>

<div
    class="fixed inset-0 z-50 flex items-center justify-center bg-black/50 backdrop-blur-sm p-4"
    on:click={() => dispatch("close")}
    on:keydown={(e) => e.key === "Escape" && dispatch("close")}
    role="button"
    tabindex="0"
>
    <div
        class="bg-[var(--background-primary)] border border-[var(--border-subtle)] rounded-xl shadow-2xl w-full max-w-md overflow-hidden animate-in fade-in zoom-in duration-200"
    >
        <div class="p-6 space-y-6">
            <div class="flex items-center justify-between">
                <h2 class="text-xl font-bold text-[var(--header-primary)]">
                    Channel Settings
                </h2>
                <button
                    class="text-[var(--text-muted)] hover:text-[var(--text-normal)] transition-colors"
                    on:click={() => dispatch("close")}
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
                        ><line x1="18" y1="6" x2="6" y2="18"></line><line
                            x1="6"
                            y1="6"
                            x2="18"
                            y2="18"
                        ></line></svg
                    >
                </button>
            </div>

            <div class="space-y-4">
                <div class="space-y-2">
                    <label
                        for="channel-name-display"
                        class="text-xs font-bold text-[var(--text-muted)] uppercase tracking-wider"
                        >Channel Name</label
                    >
                    <div
                        id="channel-name-display"
                        class="text-[var(--text-normal)] font-mono px-3 py-2 bg-[var(--background-secondary)] rounded border border-[var(--border-subtle)] opacity-75 cursor-not-allowed"
                    >
                        {channelName}
                    </div>
                </div>

                <div class="space-y-2">
                    <label
                        for="channel-topic"
                        class="text-xs font-bold text-[var(--text-muted)] uppercase tracking-wider"
                        >Topic</label
                    >
                    <input
                        id="channel-topic"
                        type="text"
                        bind:value={topic}
                        class="w-full bg-[var(--background-secondary)] border border-[var(--border-subtle)] rounded px-3 py-2 text-[var(--text-normal)] focus:border-[var(--brand-primary)] focus:outline-none transition-colors"
                        placeholder="What is this channel about?"
                    />
                </div>

                <div
                    class="flex items-center justify-between p-3 bg-[var(--background-secondary)] rounded border border-[var(--border-subtle)]"
                >
                    <div class="flex flex-col">
                        <span
                            class="text-sm font-bold text-[var(--text-normal)]"
                            >Private Channel</span
                        >
                        <span class="text-xs text-[var(--text-muted)]"
                            >Only invited members can join</span
                        >
                    </div>
                    <button
                        class="w-10 h-6 rounded-full transition-colors relative {privateChannel
                            ? 'bg-[var(--brand-primary)]'
                            : 'bg-[var(--background-tertiary)]'}"
                        on:click={() => (privateChannel = !privateChannel)}
                    >
                        <div
                            class="absolute top-1 left-1 w-4 h-4 bg-white rounded-full transition-transform {privateChannel
                                ? 'translate-x-4'
                                : 'translate-x-0'}"
                        ></div>
                    </button>
                </div>

                {#if error}
                    <div
                        class="text-red-400 text-sm bg-red-400/10 p-3 rounded border border-red-400/20"
                    >
                        {error}
                    </div>
                {/if}
            </div>

            <div class="flex justify-end gap-3 pt-2">
                <button
                    class="px-4 py-2 text-sm font-medium text-[var(--text-normal)] hover:bg-[var(--background-modifier-hover)] rounded transition-colors"
                    on:click={() => dispatch("close")}
                >
                    Cancel
                </button>
                <button
                    class="px-4 py-2 text-sm font-medium bg-[var(--brand-primary)] text-white rounded hover:bg-[var(--brand-primary)]/90 transition-colors"
                    on:click={handleSave}
                >
                    Save Changes
                </button>
            </div>
        </div>
    </div>
</div>

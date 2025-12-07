<script>
    import { createEventDispatcher, onMount } from "svelte";
    import { UpdateChannelInfo } from "../../wailsjs/go/main/App.js";
    import { appContext } from "../stores";

    export let channel = {}; // The channel object to edit

    const dispatch = createEventDispatcher();

    let channelName = "";
    let topic = "";
    let isPrivate = false;
    let icon = "";
    let description = "";
    let channelType = "text";
    let error = "";

    onMount(() => {
        if (channel) {
            channelName = channel.channel_id || "";
            topic = channel.topic || "";
            isPrivate = channel.is_private || false;
            icon = channel.icon || "";
            description = channel.description || "";
            channelType = channel.type || "text";
        }
    });

    async function handleSave() {
        if (!channelName.trim()) {
            error = "Channel name is required";
            return;
        }

        const finalName = channelName.startsWith("#")
            ? channelName
            : "#" + channelName;

        try {
            await UpdateChannelInfo(
                finalName,
                topic,
                isPrivate,
                icon,
                description,
                channelType,
            );
            dispatch("close");
        } catch (e) {
            error =
                "Failed to update channel: " + (e.message || "Unknown error");
        }
    }
</script>

<div
    class="fixed inset-0 z-50 flex items-center justify-center bg-black/60 backdrop-blur-sm p-4"
    on:click={() => dispatch("close")}
    on:keydown={(e) => e.key === "Escape" && dispatch("close")}
    role="button"
    tabindex="0"
>
    <div
        class="bg-[var(--background-primary)] border border-[var(--border-subtle)] rounded-sm shadow-[0_0_20px_rgba(0,0,0,0.5)] w-full max-w-md overflow-hidden animate-in fade-in zoom-in duration-200"
    >
        <div class="p-4 border-b border-[var(--border-subtle)]">
            <h2
                class="text-sm font-bold uppercase tracking-wider text-[var(--header-primary)]"
            >
                Edit Channel
            </h2>
        </div>

        <div class="p-6 space-y-6">
            <div class="space-y-4">
                <div class="space-y-1">
                    <label
                        for="edit-name"
                        class="text-[10px] font-bold text-[var(--text-muted)] uppercase tracking-wider"
                        >Channel Name</label
                    >
                    <input
                        id="edit-name"
                        type="text"
                        bind:value={channelName}
                        class="w-full bg-[var(--background-tertiary)] border border-[var(--border-subtle)] rounded-sm px-3 py-2 text-sm text-[var(--text-normal)] focus:border-[var(--brand-primary)] focus:outline-none transition-colors placeholder-[var(--text-muted)]/50"
                        placeholder="#general"
                        disabled
                    />
                    <p class="text-[10px] text-[var(--text-muted)]">
                        Channel names cannot be changed yet.
                    </p>
                </div>
                <div class="space-y-1">
                    <label
                        for="edit-icon"
                        class="text-[10px] font-bold text-[var(--text-muted)] uppercase tracking-wider"
                        >Icon URL (Optional)</label
                    >
                    <input
                        id="edit-icon"
                        type="text"
                        bind:value={icon}
                        class="w-full bg-[var(--background-tertiary)] border border-[var(--border-subtle)] rounded-sm px-3 py-2 text-sm text-[var(--text-normal)] focus:border-[var(--brand-primary)] focus:outline-none transition-colors placeholder-[var(--text-muted)]/50"
                        placeholder="https://example.com/icon.png"
                    />
                </div>
                <div class="space-y-1">
                    <label
                        for="edit-topic"
                        class="text-[10px] font-bold text-[var(--text-muted)] uppercase tracking-wider"
                        >Topic (Optional)</label
                    >
                    <input
                        id="edit-topic"
                        type="text"
                        bind:value={topic}
                        class="w-full bg-[var(--background-tertiary)] border border-[var(--border-subtle)] rounded-sm px-3 py-2 text-sm text-[var(--text-normal)] focus:border-[var(--brand-primary)] focus:outline-none transition-colors placeholder-[var(--text-muted)]/50"
                        placeholder="Current discussion topic"
                    />
                </div>
                <div class="space-y-1">
                    <label
                        for="edit-desc"
                        class="text-[10px] font-bold text-[var(--text-muted)] uppercase tracking-wider"
                        >Description (Optional)</label
                    >
                    <textarea
                        id="edit-desc"
                        bind:value={description}
                        class="w-full bg-[var(--background-tertiary)] border border-[var(--border-subtle)] rounded-sm px-3 py-2 text-sm text-[var(--text-normal)] focus:border-[var(--brand-primary)] focus:outline-none transition-colors min-h-[60px] placeholder-[var(--text-muted)]/50"
                        placeholder="What is this channel about?"
                    ></textarea>
                </div>
                <div class="space-y-1">
                    <label class="flex items-center gap-2 cursor-pointer group">
                        <input
                            type="checkbox"
                            bind:checked={isPrivate}
                            class="accent-[var(--brand-primary)]"
                        />
                        <span
                            class="text-sm text-[var(--text-normal)] group-hover:text-[var(--header-primary)] transition-colors"
                            >Private Channel</span
                        >
                    </label>
                </div>
                <div class="space-y-1">
                    <span
                        class="text-[10px] font-bold text-[var(--text-muted)] uppercase tracking-wider"
                        >Channel Type</span
                    >
                    <div class="flex gap-4">
                        <label
                            class="flex items-center gap-2 cursor-pointer group"
                        >
                            <input
                                type="radio"
                                bind:group={channelType}
                                value="text"
                                class="accent-[var(--brand-primary)]"
                            />
                            <span
                                class="text-sm text-[var(--text-normal)] group-hover:text-[var(--header-primary)] transition-colors"
                                >Text</span
                            >
                        </label>
                        <label
                            class="flex items-center gap-2 cursor-pointer group"
                        >
                            <input
                                type="radio"
                                bind:group={channelType}
                                value="voice"
                                class="accent-[var(--brand-primary)]"
                            />
                            <span
                                class="text-sm text-[var(--text-normal)] group-hover:text-[var(--header-primary)] transition-colors"
                                >Voice</span
                            >
                        </label>
                    </div>
                </div>
            </div>

            {#if error}
                <div
                    class="text-red-400 text-xs bg-red-500/10 p-2 rounded-sm border border-red-500/20"
                >
                    {error}
                </div>
            {/if}

            <div
                class="flex justify-end gap-2 pt-2 border-t border-[var(--border-subtle)] mt-4"
            >
                <button
                    class="px-4 py-2 text-xs font-bold uppercase tracking-wide text-[var(--text-muted)] hover:text-[var(--text-normal)] hover:bg-[var(--background-modifier-hover)] rounded-sm transition-colors"
                    on:click={() => dispatch("close")}
                >
                    Cancel
                </button>
                <button
                    class="px-4 py-2 text-xs font-bold uppercase tracking-wide text-white bg-[var(--brand-primary)] rounded-sm hover:bg-[var(--brand-primary)]/90 transition-colors shadow-lg shadow-[var(--brand-primary)]/20"
                    on:click={handleSave}
                >
                    Save Changes
                </button>
            </div>
        </div>
    </div>
</div>

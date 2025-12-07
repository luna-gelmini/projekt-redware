<script>
    import { createEventDispatcher } from "svelte";
    import { CreateChannel } from "../../wailsjs/go/main/App.js";
    import { appContext, auth } from "../stores";
    import { get } from "svelte/store";

    const dispatch = createEventDispatcher();

    // export let initialTab = "create"; // Deprecated
    let channelName = "";
    let topic = "";
    let isPrivate = false;
    let icon = "";
    let description = "";
    let channelType = "text"; // 'text' or 'voice'
    let error = "";
    let successMessage = "";

    async function handleCreate() {
        console.log("AddChannelModal: handleCreate called", {
            channelName,
            topic,
            isPrivate,
            icon,
            description,
            channelType,
        });
        const authValue = get(auth);
        if (!channelName.trim()) {
            error = "Channel name is required";
            return;
        }

        const finalName = channelName.startsWith("#")
            ? channelName
            : "#" + channelName;

        try {
            await CreateChannel(
                finalName,
                topic,
                isPrivate,
                icon,
                description,
                channelType,
            );

            // After creation, we might want to set metadata if it's private or has a topic
            // The backend CreateChannel now handles basic metadata init

            // Switch to "My Server" view so the user sees their new channel
            // We need to access the auth store value. Since we are in a function, we can use the get() method if we imported it as a store,
            // but here we are using Svelte's auto-subscription syntax $auth in the template or we need to subscribe.
            // Actually, we can just import get from svelte/store or use the $auth value if we make it reactive.
            // Simpler: just use the auth store object and .subscribe or get.
            // But wait, we are in a .svelte file. We can use $auth if we declared it.
            // Let's add 'auth' to the imports and use $auth in the script? No, $auth is only for reactive statements.
            // Inside a function, we should use get(auth).

            // Let's import get from svelte/store

            appContext.setActivePeer(authValue.currentUser.id);
            appContext.setActiveChannel(finalName);
            dispatch("close");
        } catch (e) {
            error =
                "Failed to create channel: " + (e.message || "Unknown error");
        }
    }
</script>

<div
    class="fixed inset-0 z-50 flex items-center justify-center bg-black/60 backdrop-blur-sm p-4"
>
    <div
        class="bg-[var(--background-primary)] border border-[var(--border-subtle)] rounded-sm shadow-[0_0_20px_rgba(0,0,0,0.5)] w-full max-w-lg overflow-hidden animate-in fade-in zoom-in duration-200"
    >
        <!-- Header -->
        <div class="flex border-b border-[var(--border-subtle)] px-6 py-4">
            <h2
                class="text-sm font-bold uppercase tracking-wider text-[var(--header-primary)]"
            >
                Create Channel
            </h2>
        </div>

        <div class="p-6 space-y-6">
            <div class="space-y-4">
                <p class="text-[var(--text-muted)] text-xs">
                    Create a new channel for people to join.
                </p>
                <div class="space-y-2">
                    <label
                        for="create-name"
                        class="text-[10px] font-bold text-[var(--text-muted)] uppercase tracking-wider"
                        >Channel Name</label
                    >
                    <input
                        id="create-name"
                        type="text"
                        bind:value={channelName}
                        class="w-full bg-[var(--background-tertiary)] border border-[var(--border-subtle)] rounded-sm px-3 py-2 text-sm text-[var(--text-normal)] focus:border-[var(--brand-primary)] focus:outline-none transition-colors placeholder-[var(--text-muted)]/50"
                        placeholder="#my-awesome-channel"
                        on:keydown={(e) => e.key === "Enter" && handleCreate()}
                    />
                </div>
                <div class="space-y-2">
                    <label
                        for="create-icon"
                        class="text-[10px] font-bold text-[var(--text-muted)] uppercase tracking-wider"
                        >Icon URL (Optional)</label
                    >
                    <input
                        id="create-icon"
                        type="text"
                        bind:value={icon}
                        class="w-full bg-[var(--background-tertiary)] border border-[var(--border-subtle)] rounded-sm px-3 py-2 text-sm text-[var(--text-normal)] focus:border-[var(--brand-primary)] focus:outline-none transition-colors placeholder-[var(--text-muted)]/50"
                        placeholder="https://example.com/icon.png"
                    />
                </div>
                <div class="space-y-2">
                    <label
                        for="create-topic"
                        class="text-[10px] font-bold text-[var(--text-muted)] uppercase tracking-wider"
                        >Topic (Optional)</label
                    >
                    <input
                        id="create-topic"
                        type="text"
                        bind:value={topic}
                        class="w-full bg-[var(--background-tertiary)] border border-[var(--border-subtle)] rounded-sm px-3 py-2 text-sm text-[var(--text-normal)] focus:border-[var(--brand-primary)] focus:outline-none transition-colors placeholder-[var(--text-muted)]/50"
                        placeholder="Current discussion topic"
                    />
                </div>
                <div class="space-y-2">
                    <label
                        for="create-desc"
                        class="text-[10px] font-bold text-[var(--text-muted)] uppercase tracking-wider"
                        >Description (Optional)</label
                    >
                    <textarea
                        id="create-desc"
                        bind:value={description}
                        class="w-full bg-[var(--background-tertiary)] border border-[var(--border-subtle)] rounded-sm px-3 py-2 text-sm text-[var(--text-normal)] focus:border-[var(--brand-primary)] focus:outline-none transition-colors min-h-[60px] placeholder-[var(--text-muted)]/50"
                        placeholder="What is this server about?"
                    ></textarea>
                </div>
                <div class="space-y-2">
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
                <div class="space-y-2">
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
                    on:click={handleCreate}
                >
                    Create Channel
                </button>
            </div>
        </div>
    </div>
</div>

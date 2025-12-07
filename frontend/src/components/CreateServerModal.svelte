<script>
    import { createEventDispatcher } from "svelte";
    import { UpdateServerMetadata } from "../../wailsjs/go/main/App.js";
    import { auth } from "../stores";

    const dispatch = createEventDispatcher();

    let serverName = $auth.currentUser?.username || "";
    let icon = $auth.currentUser?.avatar || "";
    let description = "";
    let error = "";

    async function handleCreate() {
        if (!serverName.trim()) {
            error = "Server name is required";
            return;
        }

        try {
            await UpdateServerMetadata(serverName, description, icon);

            // Update local auth store to reflect changes immediately if needed
            // But ideally the backend events would propagate this.
            // For now, let's just close.
            dispatch("close");
        } catch (e) {
            error =
                "Failed to create server: " + (e.message || "Unknown error");
        }
    }
</script>

<div
    class="fixed inset-0 z-50 flex items-center justify-center bg-black/60 backdrop-blur-sm p-4"
>
    <div
        class="bg-[var(--background-primary)] border border-[var(--border-subtle)] rounded-sm shadow-[0_0_20px_rgba(0,0,0,0.5)] w-full max-w-md overflow-hidden animate-in fade-in zoom-in duration-200"
    >
        <div class="p-6 space-y-6">
            <div class="space-y-2">
                <h2 class="text-lg font-bold text-[var(--header-primary)]">
                    Customize Your Server
                </h2>
                <p class="text-[var(--text-muted)] text-xs">
                    Give your new server a personality with a name and an icon.
                    You can always change it later.
                </p>
            </div>

            <div class="space-y-4">
                <div class="space-y-1">
                    <label
                        for="server-name"
                        class="text-[10px] font-bold text-[var(--text-muted)] uppercase tracking-wider"
                        >Server Name</label
                    >
                    <input
                        id="server-name"
                        type="text"
                        bind:value={serverName}
                        class="w-full bg-[var(--background-tertiary)] border border-[var(--border-subtle)] rounded-sm px-3 py-2 text-sm text-[var(--text-normal)] focus:border-[var(--brand-primary)] focus:outline-none transition-colors placeholder-[var(--text-muted)]/50"
                        placeholder="My Awesome Server"
                    />
                </div>

                <div class="space-y-1">
                    <label
                        for="server-icon"
                        class="text-[10px] font-bold text-[var(--text-muted)] uppercase tracking-wider"
                        >Icon URL (Optional)</label
                    >
                    <input
                        id="server-icon"
                        type="text"
                        bind:value={icon}
                        class="w-full bg-[var(--background-tertiary)] border border-[var(--border-subtle)] rounded-sm px-3 py-2 text-sm text-[var(--text-normal)] focus:border-[var(--brand-primary)] focus:outline-none transition-colors placeholder-[var(--text-muted)]/50"
                        placeholder="https://example.com/icon.png"
                    />
                </div>

                <div class="space-y-1">
                    <label
                        for="server-desc"
                        class="text-[10px] font-bold text-[var(--text-muted)] uppercase tracking-wider"
                        >Description (Optional)</label
                    >
                    <textarea
                        id="server-desc"
                        bind:value={description}
                        class="w-full bg-[var(--background-tertiary)] border border-[var(--border-subtle)] rounded-sm px-3 py-2 text-sm text-[var(--text-normal)] focus:border-[var(--brand-primary)] focus:outline-none transition-colors min-h-[60px] placeholder-[var(--text-muted)]/50"
                        placeholder="Tell us about your server!"
                    ></textarea>
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
                    Create
                </button>
            </div>
        </div>
    </div>
</div>

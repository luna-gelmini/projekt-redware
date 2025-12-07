<script>
    import { createEventDispatcher, onMount } from "svelte";
    import { appContext } from "../stores";
    import {
        UpdateNodeSettings,
        UpdateChannelInfo,
        CreateChannel,
    } from "../../wailsjs/go/main/App.js";

    const dispatch = createEventDispatcher();

    let activeTab = "overview"; // overview, channels
    let serverName = $appContext.serverName || "";
    let serverIcon = $appContext.serverIcon || "";
    let serverDescription = $appContext.serverDescription || "";

    let successMessage = "";
    let errorMessage = "";

    // Local channels management
    // We need to fetch channels owned by us.
    // We can use GetPeerChannels($auth.currentUser.id) if implemented, or just use what we know.
    // But for now, let's focus on Server Overview.

    async function saveOverview() {
        try {
            await UpdateNodeSettings(serverName, serverIcon, serverDescription);
            appContext.setServerMetadata(
                serverName,
                serverIcon,
                serverDescription,
            );
            successMessage = "Server settings saved!";
            setTimeout(() => (successMessage = ""), 3000);
        } catch (e) {
            errorMessage = "Failed to save settings: " + e;
            setTimeout(() => (errorMessage = ""), 3000);
        }
    }

    function close() {
        dispatch("close");
    }
</script>

<div
    class="fixed inset-0 z-50 flex items-center justify-center bg-black/50 backdrop-blur-sm p-4"
>
    <div
        class="bg-[var(--background-primary)] w-full max-w-4xl h-[80vh] rounded-lg shadow-2xl flex overflow-hidden animate-in fade-in zoom-in duration-200"
    >
        <!-- Sidebar -->
        <div class="w-60 bg-[var(--background-secondary)] flex flex-col">
            <div
                class="p-4 text-xs font-bold text-[var(--text-muted)] uppercase"
            >
                Server Settings
            </div>
            <div class="flex-1 px-2 space-y-0.5">
                <button
                    class="w-full text-left px-2 py-1.5 rounded text-[var(--text-normal)] hover:bg-[var(--background-modifier-hover)] hover:text-[var(--interactive-hover)]
                    {activeTab === 'overview'
                        ? 'bg-[var(--background-modifier-selected)] text-[var(--interactive-active)]'
                        : ''}"
                    on:click={() => (activeTab = "overview")}
                >
                    Overview
                </button>
                <button
                    class="w-full text-left px-2 py-1.5 rounded text-[var(--text-normal)] hover:bg-[var(--background-modifier-hover)] hover:text-[var(--interactive-hover)]
                    {activeTab === 'channels'
                        ? 'bg-[var(--background-modifier-selected)] text-[var(--interactive-active)]'
                        : ''}"
                    on:click={() => (activeTab = "channels")}
                >
                    Channels
                </button>
            </div>
        </div>

        <!-- Content -->
        <div class="flex-1 flex flex-col bg-[var(--background-primary)]">
            <div class="flex-1 overflow-y-auto p-10 custom-scrollbar">
                {#if activeTab === "overview"}
                    <h2
                        class="text-xl font-bold text-[var(--header-primary)] mb-6"
                    >
                        Server Overview
                    </h2>

                    <div class="flex gap-8">
                        <!-- Icon Upload Placeholder -->
                        <div class="flex flex-col gap-2 items-center">
                            <div
                                class="w-24 h-24 rounded-full bg-[var(--background-accent)] flex items-center justify-center overflow-hidden border-4 border-[var(--background-secondary)] shadow-sm"
                            >
                                {#if serverIcon}
                                    <img
                                        src={serverIcon}
                                        alt="Server Icon"
                                        class="w-full h-full object-cover"
                                    />
                                {:else}
                                    <span class="text-2xl font-bold text-white">
                                        {serverName
                                            ? serverName
                                                  .substring(0, 2)
                                                  .toUpperCase()
                                            : "??"}
                                    </span>
                                {/if}
                            </div>
                            <div class="text-xs text-[var(--text-muted)]">
                                Minimum Size: 128x128
                            </div>
                        </div>

                        <!-- Form -->
                        <div class="flex-1 space-y-6">
                            <div class="space-y-2">
                                <label
                                    class="text-xs font-bold text-[var(--text-muted)] uppercase"
                                    >Server Name</label
                                >
                                <input
                                    type="text"
                                    bind:value={serverName}
                                    class="w-full bg-[var(--background-tertiary)] border border-[var(--border-subtle)] rounded p-2 text-[var(--text-normal)] focus:border-[var(--brand-primary)] focus:outline-none"
                                />
                            </div>

                            <div class="space-y-2">
                                <label
                                    class="text-xs font-bold text-[var(--text-muted)] uppercase"
                                    >Icon URL</label
                                >
                                <input
                                    type="text"
                                    bind:value={serverIcon}
                                    class="w-full bg-[var(--background-tertiary)] border border-[var(--border-subtle)] rounded p-2 text-[var(--text-normal)] focus:border-[var(--brand-primary)] focus:outline-none"
                                />
                            </div>

                            <div class="space-y-2">
                                <label
                                    class="text-xs font-bold text-[var(--text-muted)] uppercase"
                                    >Description</label
                                >
                                <textarea
                                    bind:value={serverDescription}
                                    class="w-full bg-[var(--background-tertiary)] border border-[var(--border-subtle)] rounded p-2 text-[var(--text-normal)] focus:border-[var(--brand-primary)] focus:outline-none min-h-[100px]"
                                ></textarea>
                            </div>

                            {#if successMessage}
                                <div class="text-green-400 text-sm">
                                    {successMessage}
                                </div>
                            {/if}
                            {#if errorMessage}
                                <div class="text-red-400 text-sm">
                                    {errorMessage}
                                </div>
                            {/if}
                        </div>
                    </div>
                {:else if activeTab === "channels"}
                    <h2
                        class="text-xl font-bold text-[var(--header-primary)] mb-6"
                    >
                        Channels
                    </h2>
                    <div class="text-[var(--text-muted)]">
                        Channel management coming soon...
                    </div>
                {/if}
            </div>

            <!-- Footer -->
            <div
                class="p-4 bg-[var(--background-secondary)] flex justify-end gap-4"
            >
                <button
                    class="px-4 py-2 text-[var(--text-normal)] hover:underline"
                    on:click={close}
                >
                    Cancel
                </button>
                <button
                    class="px-6 py-2 bg-[var(--brand-primary)] text-white rounded hover:bg-[var(--brand-primary)]/90 transition-colors font-medium"
                    on:click={saveOverview}
                >
                    Save Changes
                </button>
            </div>
        </div>

        <!-- Close Button (Absolute) -->
        <button
            class="absolute top-4 right-4 text-[var(--text-muted)] hover:text-[var(--text-normal)]"
            on:click={close}
        >
            <svg
                width="24"
                height="24"
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
</div>

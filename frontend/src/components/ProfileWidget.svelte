<script>
    import { auth } from "../stores";
    import SettingsView from "./SettingsView.svelte";
    import { SetStatus, GetLibp2pID } from "../../wailsjs/go/main/App.js";
    import WidgetContainer from "./WidgetContainer.svelte";
    import { onMount } from "svelte";

    // Mock user data
    $: user = $auth.currentUser || { username: "Ghost", id: "0000" };

    let showStatusMenu = false;
    let currentStatus = "online";
    let statusMessage = "";
    let libp2pID = "";

    onMount(async () => {
        try {
            libp2pID = await GetLibp2pID();
        } catch (e) {
            console.error("Failed to get Libp2p ID:", e);
        }
    });

    const statuses = [
        { id: "online", color: "bg-green-500", label: "Online" },
        { id: "idle", color: "bg-yellow-500", label: "Idle" },
        { id: "dnd", color: "bg-red-500", label: "Do Not Disturb" },
        { id: "offline", color: "bg-gray-500", label: "Invisible" },
    ];

    async function updateStatus(status) {
        currentStatus = status;
        await SetStatus(status, statusMessage);
        showStatusMenu = false;
    }

    async function updateStatusMessage() {
        await SetStatus(currentStatus, statusMessage);
    }

    function openSettings() {
        import("../stores").then(({ modal }) => modal.open(SettingsView));
    }
</script>

<WidgetContainer title="Profile" className="h-full">
    <div class="flex flex-col h-full p-4 gap-4">
        <!-- Avatar & Status -->
        <div class="flex items-center gap-4">
            <div
                class="relative group cursor-pointer shrink-0"
                on:click={() => (showStatusMenu = !showStatusMenu)}
                on:keydown={(e) =>
                    e.key === "Enter" && (showStatusMenu = !showStatusMenu)}
                role="button"
                tabindex="0"
            >
                <div
                    class="w-16 h-16 rounded-sm bg-gradient-to-br from-indigo-500 to-purple-600 flex items-center justify-center text-white font-bold shadow-inner transition-transform duration-300 group-hover:scale-105 text-2xl overflow-hidden"
                >
                    {#if user.avatar}
                        <img
                            src={user.avatar}
                            alt="Avatar"
                            class="w-full h-full object-cover"
                        />
                    {:else}
                        {user.username && user.username.length > 0
                            ? user.username[0].toUpperCase()
                            : "?"}
                    {/if}
                </div>
                <!-- Status Dot -->
                <div
                    class="absolute -bottom-1 -right-1 w-4 h-4 border-[2px] border-[var(--background-primary)] rounded-full
                    {currentStatus === 'online'
                        ? 'bg-green-500'
                        : currentStatus === 'idle'
                          ? 'bg-yellow-500'
                          : currentStatus === 'dnd'
                            ? 'bg-red-500'
                            : 'bg-gray-500'}"
                ></div>
            </div>

            <div class="flex flex-col min-w-0">
                <span
                    class="text-lg font-bold text-[var(--header-primary)] truncate leading-tight"
                >
                    {user.username || "Unknown"}
                </span>
                <span class="text-xs text-[var(--text-muted)] truncate">
                    {statusMessage || currentStatus}
                </span>
                {#if libp2pID}
                    <span
                        class="text-[10px] text-[var(--text-muted)] truncate font-mono mt-1"
                        title={libp2pID}
                    >
                        ID: {libp2pID.substring(0, 8)}...
                    </span>
                {/if}
            </div>
        </div>

        <!-- Status Menu (Inline if open) -->
        {#if showStatusMenu}
            <div
                class="bg-[var(--background-secondary)] border border-[var(--border-subtle)] rounded-sm p-2 space-y-1 animate-in fade-in zoom-in duration-200"
            >
                {#each statuses as status}
                    <button
                        class="w-full flex items-center gap-2 px-2 py-1.5 rounded-sm hover:bg-[var(--background-modifier-hover)] text-left text-xs"
                        on:click={() => updateStatus(status.id)}
                    >
                        <div class="w-2 h-2 rounded-full {status.color}"></div>
                        <span class="text-[var(--text-normal)]"
                            >{status.label}</span
                        >
                    </button>
                {/each}
                <div class="mt-2 pt-2 border-t border-[var(--border-subtle)]">
                    <input
                        type="text"
                        placeholder="Set a custom status..."
                        bind:value={statusMessage}
                        on:change={updateStatusMessage}
                        class="w-full bg-[var(--background-tertiary)] text-[var(--text-normal)] text-xs px-2 py-1 rounded-sm border border-[var(--border-subtle)] focus:border-[var(--brand-primary)] outline-none"
                    />
                </div>
            </div>
        {/if}

        <!-- Actions -->
        <div class="grid grid-cols-3 gap-2 mt-auto">
            <button
                class="flex flex-col items-center justify-center gap-1 p-2 rounded-sm bg-[var(--background-secondary)] hover:bg-[var(--background-modifier-hover)] transition-colors text-[var(--text-muted)] hover:text-[var(--text-normal)]"
                title="Mute"
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
                    ><path
                        d="M12 1a3 3 0 0 0-3 3v8a3 3 0 0 0 6 0V4a3 3 0 0 0-3-3z"
                    ></path><path d="M19 10v2a7 7 0 0 1-14 0v-2"></path><line
                        x1="12"
                        y1="19"
                        x2="12"
                        y2="23"
                    ></line><line x1="8" y1="23" x2="16" y2="23"></line></svg
                >
                <span class="text-[10px] uppercase font-bold">Mute</span>
            </button>
            <button
                class="flex flex-col items-center justify-center gap-1 p-2 rounded-sm bg-[var(--background-secondary)] hover:bg-[var(--background-modifier-hover)] transition-colors text-[var(--text-muted)] hover:text-[var(--text-normal)]"
                title="Deafen"
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
                    ><path d="M3 18v-6a9 9 0 0 1 18 0v6"></path><path
                        d="M21 19a2 2 0 0 1-2 2h-1a2 2 0 0 1-2-2v-3a2 2 0 0 1 2-2h3zM3 19a2 2 0 0 0 2 2h1a2 2 0 0 0 2-2v-3a2 2 0 0 0-2-2H3z"
                    ></path></svg
                >
                <span class="text-[10px] uppercase font-bold">Deafen</span>
            </button>
            <button
                class="flex flex-col items-center justify-center gap-1 p-2 rounded-sm bg-[var(--background-secondary)] hover:bg-[var(--background-modifier-hover)] transition-colors text-[var(--text-muted)] hover:text-[var(--text-normal)]"
                title="Settings"
                on:click={openSettings}
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
                    ><circle cx="12" cy="12" r="3"></circle><path
                        d="M19.4 15a1.65 1.65 0 0 0 .33 1.82l.06.06a2 2 0 0 1 0 2.83 2 2 0 0 1-2.83 0l-.06-.06a1.65 1.65 0 0 0-1.82-.33 1.65 1.65 0 0 0-1 1.51V21a2 2 0 0 1-2 2 2 2 0 0 1-2-2v-.09A1.65 1.65 0 0 0 9 19.4a1.65 1.65 0 0 0-1.82.33l-.06.06a2 2 0 0 1-2.83 0 2 2 0 0 1 0-2.83l.06-.06a1.65 1.65 0 0 0 .33-1.82 1.65 1.65 0 0 0-1.51-1H3a2 2 0 0 1-2-2 2 2 0 0 1 2-2h.09A1.65 1.65 0 0 0 4.6 9a1.65 1.65 0 0 0-.33-1.82l-.06-.06a2 2 0 0 1 0-2.83 2 2 0 0 1 2.83 0l.06.06a1.65 1.65 0 0 0 1.82.33H9a1.65 1.65 0 0 0 1-1.51V3a2 2 0 0 1 2-2 2 2 0 0 1 2 2v.09a1.65 1.65 0 0 0 1 1.51 1.65 1.65 0 0 0 1.82-.33l.06-.06a2 2 0 0 1 2.83 0 2 2 0 0 1 0 2.83l-.06.06a1.65 1.65 0 0 0-.33 1.82V9a1.65 1.65 0 0 0 1.51 1H21a2 2 0 0 1 2 2 2 2 0 0 1-2 2h-.09a1.65 1.65 0 0 0-1.51 1z"
                    ></path></svg
                >
                <span class="text-[10px] uppercase font-bold">Settings</span>
            </button>
        </div>
    </div>
</WidgetContainer>

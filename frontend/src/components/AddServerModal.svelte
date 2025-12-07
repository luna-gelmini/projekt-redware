<script>
    import { modal, appContext } from "../stores";
    import {
        UpdateServerMetadata,
        JoinNode,
        GetPublicNodes,
    } from "../../wailsjs/go/main/App.js";
    import { onMount } from "svelte";

    // Tabs
    let activeTab = "create"; // 'create' or 'join'

    // Create State
    let serverName = "";
    let serverIcon = "";
    let serverDescription = "";
    let createError = "";

    // Join State
    let joinCode = ""; // Can be IP:Port or UserID
    let joinError = "";
    let discoveredNodes = [];
    let isJoining = false;

    async function handleCreate() {
        if (!serverName.trim()) {
            createError = "Server name is required";
            return;
        }
        try {
            await UpdateServerMetadata(
                serverName,
                serverDescription,
                serverIcon,
            );
            // The backend now auto-saves this as a joined node.
            // We can close the modal.
            modal.close();
        } catch (e) {
            createError = "Failed to create server: " + e;
        }
    }

    async function handleJoin(nodeID = null) {
        const target = nodeID || joinCode;
        if (!target.trim()) {
            joinError = "Please enter a Join Code or Address";
            return;
        }
        isJoining = true;
        joinError = "";
        try {
            // JoinNode in backend now should handle the connection.
            // Note: JoinNode in app.go takes (nodeID, joinCode).
            // If we are joining by IP, nodeID is the IP.
            // If we are joining by UserID (from discovery), nodeID is UserID.
            // For now, we assume public join (empty code) or the input IS the code/address.

            // If joining from list, we pass the ID.
            // If joining from input, we pass the input as ID (and empty code for now).
            // TODO: Support separate Join Code input if needed.

            const success = await JoinNode(target, "");
            if (success) {
                modal.close();
            } else {
                joinError = "Failed to join. Check the address or try again.";
            }
        } catch (e) {
            joinError = "Error joining node: " + e;
        } finally {
            isJoining = false;
        }
    }

    async function fetchDiscovered() {
        try {
            const nodes = await GetPublicNodes();
            // Filter out nodes we already joined? Maybe later.
            discoveredNodes = nodes || [];
        } catch (e) {
            console.error("Failed to fetch public nodes:", e);
        }
    }

    onMount(() => {
        fetchDiscovered();
    });
</script>

<div
    class="fixed inset-0 z-[100] flex items-center justify-center bg-black/60 backdrop-blur-sm"
    on:click|self={() => modal.close()}
    on:keydown={(e) => e.key === "Escape" && modal.close()}
    role="button"
    tabindex="0"
>
    <div
        class="w-full max-w-md bg-[var(--background-primary)] rounded-sm shadow-2xl border border-[var(--border-subtle)] overflow-hidden flex flex-col max-h-[80vh]"
    >
        <!-- Header -->
        <div class="p-6 text-center space-y-2">
            <h2 class="text-2xl font-bold text-[var(--header-primary)]">
                Add a Server
            </h2>
            <p class="text-sm text-[var(--text-muted)]">
                Create your own community or join an existing one.
            </p>
        </div>

        <!-- Tabs -->
        <div class="flex border-b border-[var(--border-subtle)]">
            <button
                class="flex-1 py-3 text-sm font-bold uppercase tracking-wide transition-colors
                {activeTab === 'create'
                    ? 'text-[var(--brand-primary)] border-b-2 border-[var(--brand-primary)] bg-[var(--background-secondary)]'
                    : 'text-[var(--text-muted)] hover:text-[var(--text-normal)] hover:bg-[var(--background-modifier-hover)]'}"
                on:click={() => (activeTab = "create")}
            >
                Create My Own
            </button>
            <button
                class="flex-1 py-3 text-sm font-bold uppercase tracking-wide transition-colors
                {activeTab === 'join'
                    ? 'text-[var(--brand-primary)] border-b-2 border-[var(--brand-primary)] bg-[var(--background-secondary)]'
                    : 'text-[var(--text-muted)] hover:text-[var(--text-normal)] hover:bg-[var(--background-modifier-hover)]'}"
                on:click={() => (activeTab = "join")}
            >
                Join a Server
            </button>
        </div>

        <!-- Content -->
        <div class="p-6 overflow-y-auto custom-scrollbar">
            {#if activeTab === "create"}
                <div class="space-y-4">
                    <div class="space-y-1">
                        <label
                            for="server-name"
                            class="text-xs font-bold text-[var(--text-muted)] uppercase"
                            >Server Name</label
                        >
                        <input
                            id="server-name"
                            type="text"
                            bind:value={serverName}
                            class="w-full bg-[var(--background-tertiary)] border border-[var(--border-subtle)] rounded-sm px-3 py-2 text-sm text-[var(--text-normal)] focus:border-[var(--brand-primary)] focus:outline-none"
                            placeholder="My Awesome Guild"
                        />
                    </div>
                    <div class="space-y-1">
                        <label
                            for="server-icon"
                            class="text-xs font-bold text-[var(--text-muted)] uppercase"
                            >Icon URL (Optional)</label
                        >
                        <input
                            id="server-icon"
                            type="text"
                            bind:value={serverIcon}
                            class="w-full bg-[var(--background-tertiary)] border border-[var(--border-subtle)] rounded-sm px-3 py-2 text-sm text-[var(--text-normal)] focus:border-[var(--brand-primary)] focus:outline-none"
                            placeholder="https://..."
                        />
                    </div>
                    <div class="space-y-1">
                        <label
                            for="server-desc"
                            class="text-xs font-bold text-[var(--text-muted)] uppercase"
                            >Description (Optional)</label
                        >
                        <textarea
                            id="server-desc"
                            bind:value={serverDescription}
                            class="w-full bg-[var(--background-tertiary)] border border-[var(--border-subtle)] rounded-sm px-3 py-2 text-sm text-[var(--text-normal)] focus:border-[var(--brand-primary)] focus:outline-none min-h-[80px] resize-none"
                            placeholder="A place for..."
                        ></textarea>
                    </div>

                    {#if createError}
                        <div class="text-xs text-red-400">{createError}</div>
                    {/if}

                    <button
                        class="w-full py-2.5 bg-[var(--brand-primary)] hover:bg-[var(--brand-hover)] text-white font-bold rounded-sm shadow-lg transition-all"
                        on:click={handleCreate}
                    >
                        Create Server
                    </button>
                </div>
            {:else}
                <div class="space-y-6">
                    <!-- Manual Join -->
                    <div class="space-y-2">
                        <label
                            for="join-code"
                            class="text-xs font-bold text-[var(--text-muted)] uppercase"
                            >Join by Address / ID</label
                        >
                        <div class="flex gap-2">
                            <input
                                id="join-code"
                                type="text"
                                bind:value={joinCode}
                                class="flex-1 bg-[var(--background-tertiary)] border border-[var(--border-subtle)] rounded-sm px-3 py-2 text-sm text-[var(--text-normal)] focus:border-[var(--brand-primary)] focus:outline-none"
                                placeholder="IP:Port or UserID"
                            />
                            <button
                                class="px-4 bg-[var(--brand-primary)] hover:bg-[var(--brand-hover)] text-white font-bold rounded-sm transition-all disabled:opacity-50"
                                on:click={() => handleJoin()}
                                disabled={isJoining}
                            >
                                {isJoining ? "..." : "Join"}
                            </button>
                        </div>
                        {#if joinError}
                            <div class="text-xs text-red-400">{joinError}</div>
                        {/if}
                    </div>

                    <!-- Discovery List -->
                    <div class="space-y-2">
                        <div
                            class="flex items-center justify-between border-b border-[var(--border-subtle)] pb-1"
                        >
                            <h3
                                class="text-xs font-bold text-[var(--header-secondary)] uppercase"
                            >
                                Discovered Servers
                            </h3>
                            <button
                                class="text-xs text-[var(--brand-primary)] hover:underline"
                                on:click={fetchDiscovered}
                            >
                                Refresh
                            </button>
                        </div>
                        <div class="space-y-2">
                            {#if discoveredNodes.length === 0}
                                <div
                                    class="text-sm text-[var(--text-muted)] text-center py-4"
                                >
                                    No public servers found.
                                </div>
                            {:else}
                                {#each discoveredNodes as node}
                                    <div
                                        class="flex items-center gap-3 p-2 rounded-sm hover:bg-[var(--background-modifier-hover)] transition-colors group"
                                    >
                                        <div
                                            class="w-10 h-10 rounded-full bg-[var(--background-secondary)] flex items-center justify-center overflow-hidden border border-[var(--border-subtle)]"
                                        >
                                            {#if node.server_icon || node.avatar}
                                                <img
                                                    src={node.server_icon ||
                                                        node.avatar}
                                                    alt={node.server_name}
                                                    class="w-full h-full object-cover"
                                                />
                                            {:else}
                                                <span
                                                    class="text-xs font-bold text-[var(--text-normal)]"
                                                >
                                                    {(
                                                        node.server_name ||
                                                        node.username ||
                                                        "?"
                                                    )
                                                        .substring(0, 2)
                                                        .toUpperCase()}
                                                </span>
                                            {/if}
                                        </div>
                                        <div class="flex-1 min-w-0">
                                            <div
                                                class="text-sm font-bold text-[var(--header-primary)] truncate"
                                            >
                                                {node.server_name ||
                                                    node.username}
                                            </div>
                                            <div
                                                class="text-xs text-[var(--text-muted)] truncate"
                                            >
                                                {node.server_description ||
                                                    node.status_message ||
                                                    "No description"}
                                            </div>
                                        </div>
                                        <button
                                            class="px-3 py-1 bg-[var(--button-secondary-background)] hover:bg-[var(--button-secondary-hover)] text-[var(--text-normal)] text-xs font-bold rounded-sm transition-colors"
                                            on:click={() =>
                                                handleJoin(node.user_id)}
                                        >
                                            Join
                                        </button>
                                    </div>
                                {/each}
                            {/if}
                        </div>
                    </div>
                </div>
            {/if}
        </div>
    </div>
</div>

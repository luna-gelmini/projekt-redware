<script>
    import { createEventDispatcher, onMount } from "svelte";
    import { JoinNode, GetPublicNodes } from "../../wailsjs/go/main/App.js";
    import { appContext } from "../stores";

    const dispatch = createEventDispatcher();

    let nodeId = "";
    let joinCode = "";
    let publicNodes = [];
    let error = "";
    let isLoading = false;

    async function fetchPublicNodes() {
        try {
            publicNodes = (await GetPublicNodes()) || [];
        } catch (e) {
            console.error("Failed to fetch public nodes:", e);
        }
    }

    async function handleJoin() {
        if (!nodeId.trim()) {
            error = "Node ID or Address is required";
            return;
        }

        isLoading = true;
        error = "";

        try {
            const success = await JoinNode(nodeId, joinCode);
            if (success) {
                appContext.setActivePeer(nodeId); // Set as active if joined
                dispatch("close");
            } else {
                error = "Failed to join node. Check ID/Code.";
            }
        } catch (e) {
            error = "Error joining node: " + (e.message || "Unknown error");
        } finally {
            isLoading = false;
        }
    }

    function selectNode(node) {
        nodeId = node.user_id;
        // If we had a way to know if it requires a code, we could focus that input
    }

    onMount(() => {
        fetchPublicNodes();
    });
</script>

<div
    class="fixed inset-0 z-50 flex items-center justify-center bg-black/50 backdrop-blur-sm p-4"
>
    <div
        class="bg-[var(--background-primary)] border border-[var(--border-subtle)] rounded-sm shadow-2xl w-full max-w-md overflow-hidden animate-in fade-in zoom-in duration-200"
    >
        <div
            class="p-4 border-b border-[var(--border-subtle)] bg-[var(--background-secondary)]"
        >
            <h2
                class="text-lg font-bold text-[var(--header-primary)] uppercase tracking-wider"
            >
                Join a Node
            </h2>
            <p class="text-xs text-[var(--text-muted)]">
                Connect to a P2P server
            </p>
        </div>

        <div class="p-6 space-y-6">
            <!-- Public Nodes List -->
            <div class="space-y-2">
                <label
                    class="text-xs font-bold text-[var(--text-muted)] uppercase tracking-wider"
                    >Discovered Nodes</label
                >
                <div
                    class="max-h-40 overflow-y-auto border border-[var(--border-subtle)] rounded-sm bg-[var(--background-tertiary)] p-2 space-y-1"
                >
                    {#if publicNodes.length === 0}
                        <div
                            class="text-xs text-[var(--text-muted)] text-center py-4"
                        >
                            No public nodes discovered yet.
                        </div>
                    {:else}
                        {#each publicNodes as node}
                            <button
                                class="w-full flex items-center gap-3 p-2 hover:bg-[var(--background-modifier-hover)] rounded-sm transition-colors text-left"
                                on:click={() => selectNode(node)}
                            >
                                <div
                                    class="w-8 h-8 rounded-sm bg-[var(--brand-primary)] flex items-center justify-center text-white font-bold text-xs shrink-0"
                                >
                                    {(node.server_name || node.username)
                                        .substring(0, 2)
                                        .toUpperCase()}
                                </div>
                                <div class="overflow-hidden">
                                    <div
                                        class="text-sm font-bold text-[var(--text-normal)] truncate"
                                    >
                                        {node.server_name || node.username}
                                    </div>
                                    <div
                                        class="text-xs text-[var(--text-muted)] truncate"
                                    >
                                        {node.server_description ||
                                            "No description"}
                                    </div>
                                </div>
                            </button>
                        {/each}
                    {/if}
                </div>
            </div>

            <!-- Manual Input -->
            <div class="space-y-4 border-t border-[var(--border-subtle)] pt-4">
                <div class="space-y-2">
                    <label
                        for="node-id"
                        class="text-xs font-bold text-[var(--text-muted)] uppercase tracking-wider"
                        >Node ID / Address</label
                    >
                    <input
                        id="node-id"
                        type="text"
                        bind:value={nodeId}
                        class="w-full bg-[var(--background-secondary)] border border-[var(--border-subtle)] rounded-sm px-3 py-2 text-[var(--text-normal)] focus:border-[var(--brand-primary)] focus:outline-none transition-colors font-mono text-sm"
                        placeholder="User ID or IP:Port"
                    />
                </div>
                <div class="space-y-2">
                    <label
                        for="join-code"
                        class="text-xs font-bold text-[var(--text-muted)] uppercase tracking-wider"
                        >Join Code (Optional)</label
                    >
                    <input
                        id="join-code"
                        type="text"
                        bind:value={joinCode}
                        class="w-full bg-[var(--background-secondary)] border border-[var(--border-subtle)] rounded-sm px-3 py-2 text-[var(--text-normal)] focus:border-[var(--brand-primary)] focus:outline-none transition-colors font-mono text-sm"
                        placeholder="Enter code if private"
                    />
                </div>
            </div>

            {#if error}
                <div
                    class="text-red-400 text-xs bg-red-400/10 p-2 rounded-sm border border-red-400/20"
                >
                    {error}
                </div>
            {/if}

            <div class="flex justify-end gap-3 pt-2">
                <button
                    class="px-4 py-2 text-sm font-medium text-[var(--text-normal)] hover:bg-[var(--background-modifier-hover)] rounded-sm transition-colors"
                    on:click={() => dispatch("close")}
                >
                    Cancel
                </button>
                <button
                    class="px-4 py-2 text-sm font-bold text-white bg-[var(--brand-primary)] rounded-sm hover:bg-[var(--brand-primary)]/90 transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
                    on:click={handleJoin}
                    disabled={isLoading}
                >
                    {isLoading ? "Joining..." : "Join Node"}
                </button>
            </div>
        </div>
    </div>
</div>

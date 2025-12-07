<script>
    import { appContext, appearance, auth, modal } from "../stores";
    import {
        GetJoinedNodes,
        ConnectPeer,
        GetLocalNodeMetadata,
    } from "../../wailsjs/go/main/App.js";
    import { onMount, onDestroy } from "svelte";
    import { EventsOn } from "../../wailsjs/runtime/runtime.js";
    import logoDark from "../assets/images/logo_dark.png";
    import logoLight from "../assets/images/logo_light.png";
    import ProfileModal from "./ProfileModal.svelte";
    import AddServerModal from "./AddServerModal.svelte";
    import WidgetContainer from "./WidgetContainer.svelte";
    import { flip } from "svelte/animate";

    let joinedGuilds = [];
    let pollInterval;
    let localNodeMetadata = null;

    // Drag & Drop State
    let draggingItem = null;
    let hoveringItem = null;

    $: guilds = [
        {
            id: "home",
            name: "Home",
            icon: "🔥",
            type: "home",
            avatar: $appearance.theme === "light" ? logoLight : logoDark,
        },
        ...joinedGuilds.map((g) => ({
            id: g.node_id,
            name: g.name || "Unknown Server",
            icon: (g.name || "??").substring(0, 2).toUpperCase(),
            type: "node",
            avatar: g.icon,
            color: stringToColor(g.name || "Unknown"),
            isLocal: g.owner_id === $auth.currentUser?.id,
            description: g.description,
        })),
    ];

    $: activeId = $appContext.activePeerId || "home";

    function stringToColor(str) {
        let hash = 0;
        for (let i = 0; i < str.length; i++) {
            hash = str.charCodeAt(i) + ((hash << 5) - hash);
        }
        const c = (hash & 0x00ffffff).toString(16).toUpperCase();
        return "#" + "00000".substring(0, 6 - c.length) + c;
    }

    async function fetchGuilds() {
        try {
            const nodes = (await GetJoinedNodes()) || [];
            // Sort: Local Node first, then by Join Date (descending)
            // But we want to respect custom order if we implement it later.
            // For now, let's just use the DB order (JoinedAt DESC).
            // But wait, "My Server" should probably be at the top if it's in the list.

            // Filter out duplicates if any (though DB constraint handles it)
            // Filter to only show servers (is_server = true)
            joinedGuilds = nodes.filter((n) => n.is_server);
        } catch (e) {
            console.error("Failed to fetch guilds:", e);
        }
    }

    async function handleGuildClick(guild) {
        if (guild.id === "home") {
            appContext.setActivePeer(null);
        } else {
            // If it's a remote node, ensure we are connected
            if (!guild.isLocal) {
                // We need the address. JoinedNode has it.
                // Find the full guild object from joinedGuilds to get address
                const fullGuild = joinedGuilds.find(
                    (g) => g.node_id === guild.id,
                );
                if (fullGuild && fullGuild.address) {
                    await ConnectPeer(guild.id, fullGuild.address);
                }
            }
            appContext.setActivePeer(guild.id);
        }
    }

    function handleContextMenu(e, guild) {
        if (guild.type === "node") {
            e.preventDefault();
            // TODO: Guild Context Menu (Leave Server, etc.)
            // For now, maybe just Profile? Or nothing.
        }
    }

    function openAddServerModal() {
        modal.open(AddServerModal, {});
    }

    // Drag & Drop Handlers
    function handleDragStart(e, item) {
        draggingItem = item;
        e.dataTransfer.effectAllowed = "move";
        e.dataTransfer.setData("text/plain", item.id);
    }

    function handleDragOver(e, item) {
        if (!draggingItem || draggingItem === item) return;
        e.preventDefault();
        hoveringItem = item;
    }

    function handleDrop(e, targetItem) {
        if (!draggingItem) return;
        e.preventDefault();

        // Reorder logic (Client-side only for now, or persist to localStorage)
        const fromIndex = guilds.indexOf(draggingItem);
        const toIndex = guilds.indexOf(targetItem);

        if (fromIndex !== toIndex) {
            // We need to reorder 'joinedGuilds' because 'guilds' is derived
            // This is tricky because 'guilds' includes 'home'.
            // Let's just reorder the derived array for display? No, that won't persist.
            // We should reorder joinedGuilds.

            // Map back to joinedGuilds index
            // draggingItem is from 'guilds'.
            // If dragging 'home', ignore.
            if (draggingItem.id === "home" || targetItem.id === "home") return;

            const realFrom = joinedGuilds.findIndex(
                (g) => g.node_id === draggingItem.id,
            );
            const realTo = joinedGuilds.findIndex(
                (g) => g.node_id === targetItem.id,
            );

            if (realFrom !== -1 && realTo !== -1) {
                const newOrder = [...joinedGuilds];
                const [moved] = newOrder.splice(realFrom, 1);
                newOrder.splice(realTo, 0, moved);
                joinedGuilds = newOrder;
                // TODO: Persist order
            }
        }

        draggingItem = null;
        hoveringItem = null;
    }

    onMount(() => {
        fetchGuilds();
        pollInterval = setInterval(fetchGuilds, 5000); // Poll for new joins

        EventsOn("localNodeUpdated", () => {
            fetchGuilds(); // Refresh if local node changes
        });
    });

    onDestroy(() => {
        if (pollInterval) clearInterval(pollInterval);
    });
</script>

<WidgetContainer title="Guilds" className="h-full">
    <div
        class="flex flex-col gap-2 p-2 overflow-y-auto h-full content-start items-center custom-scrollbar"
    >
        {#each guilds as guild (guild.id)}
            {@const isActive =
                (guild.id === "home" && !activeId) || activeId === guild.id}

            <div
                class="w-full flex flex-col items-center gap-1"
                animate:flip={{ duration: 200 }}
            >
                <div
                    class="w-full aspect-square"
                    draggable={guild.id !== "home"}
                    on:dragstart={(e) => handleDragStart(e, guild)}
                    on:dragover={(e) => handleDragOver(e, guild)}
                    on:drop={(e) => handleDrop(e, guild)}
                >
                    <button
                        class="group relative flex flex-col items-center justify-center w-full h-full transition-all duration-200 outline-none"
                        on:click={() => handleGuildClick(guild)}
                        on:contextmenu={(e) => handleContextMenu(e, guild)}
                        title={guild.name}
                    >
                        <!-- Icon/Avatar -->
                        <div
                            class="w-full h-full flex items-center justify-center overflow-hidden border transition-all duration-200
                            {isActive
                                ? 'border-[var(--brand-primary)] rounded-sm shadow-[0_0_10px_var(--brand-primary)]'
                                : 'border-[var(--border-subtle)] rounded-sm hover:border-[var(--text-muted)]'}
                            {draggingItem === guild ? 'opacity-50' : ''}
                            {hoveringItem === guild
                                ? 'border-2 border-dashed border-[var(--brand-primary)]'
                                : ''}"
                            style={!guild.avatar && guild.type !== "home"
                                ? `background-color: ${guild.color}; color: white;`
                                : guild.type === "home"
                                  ? "background-color: var(--background-secondary);"
                                  : ""}
                        >
                            {#if guild.avatar}
                                <img
                                    src={guild.avatar}
                                    alt={guild.name}
                                    class="w-full h-full object-cover"
                                />
                            {:else}
                                <span
                                    class="text-lg font-bold text-[var(--text-normal)]"
                                >
                                    {guild.icon}
                                </span>
                            {/if}
                        </div>

                        <!-- Pill Indicator -->
                        {#if isActive}
                            <div
                                class="absolute -left-3 top-1/2 -translate-y-1/2 w-1 h-8 bg-[var(--text-normal)] rounded-r-full"
                            ></div>
                        {:else}
                            <div
                                class="absolute -left-3 top-1/2 -translate-y-1/2 w-1 h-2 bg-[var(--text-normal)] rounded-r-full opacity-0 group-hover:opacity-100 transition-all duration-200 group-hover:h-5"
                            ></div>
                        {/if}
                    </button>
                </div>

                {#if guild.id === "home"}
                    <div
                        class="w-8 h-[2px] bg-[var(--background-modifier-accent)] rounded-lg shrink-0 my-1"
                    ></div>
                {/if}
            </div>
        {/each}

        <!-- Add Server Button -->
        <button
            class="flex items-center justify-center w-full aspect-square border border-dashed border-[var(--border-subtle)] rounded-sm text-[var(--text-muted)] hover:text-[var(--brand-primary)] hover:border-[var(--brand-primary)] transition-all duration-200 group bg-[var(--background-secondary)]/50"
            on:click={openAddServerModal}
            title="Add a Server"
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
                class="group-hover:rotate-90 transition-transform duration-200"
            >
                <line x1="12" y1="5" x2="12" y2="19"></line>
                <line x1="5" y1="12" x2="19" y2="12"></line>
            </svg>
        </button>
    </div>
</WidgetContainer>

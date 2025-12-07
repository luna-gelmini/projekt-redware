<script>
    import { appContext, appearance, auth } from "../stores";
    import {
        GetDiscoveredPeers,
        ConnectPeer,
        JoinChannel,
    } from "../../wailsjs/go/main/App.js";
    import { onMount, onDestroy } from "svelte";
    import logoDark from "../assets/images/logo_dark.png";
    import logoLight from "../assets/images/logo_light.png";
    import ProfileModal from "./ProfileModal.svelte";
    import AddChannelModal from "./AddChannelModal.svelte";
    import { modal } from "../stores";

    let currentPeerNodes = [];
    let joinedChannels = [];
    let pollInterval;

    $: nodes = [
        {
            id: "home",
            name: "Home",
            icon: "🔥",
            type: "home",
            avatar: $appearance.theme === "light" ? logoLight : logoDark,
        },
        // Local Node (My Server)
        ...($auth.currentUser
            ? [
                  {
                      id: $auth.currentUser.id,
                      name: $auth.currentUser.username || "My Server",
                      icon: ($auth.currentUser.username || "Me")
                          .substring(0, 2)
                          .toUpperCase(),
                      type: "node",
                      avatar: $auth.currentUser.avatar,
                      color: stringToColor($auth.currentUser.username || "Me"),
                      isLocal: true,
                  },
              ]
            : []),
        ...currentPeerNodes,
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

    async function fetchPeers() {
        try {
            const peers = (await GetDiscoveredPeers()) || [];
            // Map backend peers to UI nodes
            const peerNodes = peers.map((p) => ({
                id: p.user_id,
                name: p.server_name || p.username, // Use Server Name if available
                icon: (p.server_name || p.username)
                    .substring(0, 2)
                    .toUpperCase(),
                avatar: p.server_icon || p.avatar, // Use Server Icon if available
                type: "node",
                ip: p.ip,
                port: p.port,
                status: p.status || "offline", // Add status
                statusMessage: p.status_message,
                color: stringToColor(p.server_name || p.username),
            }));

            currentPeerNodes = peerNodes;
            appContext.setPeers(peers);
        } catch (e) {
            console.error("Failed to fetch peers:", e);
        }
    }

    async function handleNodeClick(node) {
        console.log("Node clicked:", node);
        if (node.id === "home") {
            appContext.setActivePeer(null);
        } else if (node.type === "channel" || node.isLocal) {
            // For channels or local node, just set as active
            appContext.setActivePeer(node.id);
        } else {
            // Connect to peer
            const address = `${node.ip}:${node.port}`;
            await ConnectPeer(node.id, address);
            appContext.setActivePeer(node.id);
        }
    }

    function handleContextMenu(e, node) {
        if (node.type === "node") {
            e.preventDefault();
            modal.open(ProfileModal, {
                user: {
                    username: node.name,
                    avatar: node.avatar,
                    status: node.status,
                    statusMessage: node.statusMessage,
                    id: node.id,
                },
            });
        }
    }

    function openAddChannelModal() {
        modal.open(AddChannelModal, {});
    }

    onMount(() => {
        fetchPeers();
        pollInterval = setInterval(fetchPeers, 5000);
    });

    onDestroy(() => {
        if (pollInterval) clearInterval(pollInterval);
    });
</script>

<div
    class="w-[80px] h-full flex flex-col items-center py-6 bg-[var(--background-tertiary)]/50 backdrop-blur-sm border-[var(--border-subtle)] space-y-4 overflow-y-auto scrollbar-hide"
>
    {#each nodes as node}
        {@const isActive =
            (node.id === "home" && !activeId) || activeId === node.id}

        <button
            class="relative group flex items-center justify-center w-12 h-12 transition-all duration-300 outline-none"
            on:click={() => handleNodeClick(node)}
            on:contextmenu={(e) => handleContextMenu(e, node)}
        >
            <!-- Active Glow/Background -->
            <div
                class="absolute inset-0 transition-all duration-300
                {isActive
                    ? 'bg-[var(--brand-primary)]/20'
                    : 'bg-transparent group-hover:bg-[var(--background-modifier-hover)]'}"
            ></div>

            <!-- Active Border Indicator (Right side) -->
            <div
                class="absolute right-0 w-1 h-6 bg-[var(--brand-primary)] rounded-l-full transition-all duration-300
                {isActive
                    ? 'opacity-100 translate-x-0'
                    : 'opacity-0 translate-x-2'}"
            ></div>

            <!-- Icon/Avatar -->
            <div
                class="relative z-10 w-12 h-12 flex items-center justify-center overflow-hidden transition-all duration-300
                {isActive ? 'rounded-sm' : 'rounded-sm group-hover:rounded-sm'}"
                style={!node.avatar && node.type !== "home"
                    ? `background-color: ${node.color}; color: white;`
                    : node.type === "home"
                      ? "background-color: var(--background-secondary);"
                      : ""}
            >
                {#if node.avatar}
                    <img
                        src={node.avatar}
                        alt={node.name}
                        class="w-full h-full object-cover"
                    />
                {:else}
                    <span
                        class="text-lg font-bold transition-colors duration-300
                        {isActive
                            ? 'text-[var(--brand-primary)]'
                            : 'text-[var(--text-muted)] group-hover:text-[var(--text-normal)]'}"
                    >
                        {node.icon}
                    </span>
                {/if}
            </div>

            <!-- Tooltip -->
            <div
                class="absolute left-full ml-4 px-3 py-1.5 bg-[var(--background-floating)] border border-[var(--border-subtle)] text-[var(--text-normal)] text-xs font-medium rounded shadow-xl opacity-0 group-hover:opacity-100 transition-opacity duration-200 pointer-events-none whitespace-nowrap z-50 backdrop-blur-md"
            >
                <div class="flex flex-col items-start">
                    <span>{node.name}</span>
                </div>
            </div>
        </button>
    {/each}

    <!-- Add Channel Button (Join/Create) -->
    <button
        class="group flex items-center justify-center w-12 h-12 mt-2 text-[var(--status-positive)] hover:text-white hover:bg-[var(--status-positive)] rounded-full transition-all duration-200 bg-[var(--background-primary)]"
        on:click={openAddChannelModal}
        title="Add Channel"
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
            class="transition-transform duration-300 group-hover:rotate-90"
        >
            <line x1="12" y1="5" x2="12" y2="19"></line>
            <line x1="5" y1="12" x2="19" y2="12"></line>
        </svg>
    </button>
</div>

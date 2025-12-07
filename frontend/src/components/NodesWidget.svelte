<script>
    import { appContext, appearance, auth, modal } from "../stores";
    import {
        GetDiscoveredPeers,
        ConnectPeer,
        GetPublicNodes,
        GetLocalNodeMetadata,
    } from "../../wailsjs/go/main/App.js";
    import { onMount, onDestroy } from "svelte";
    import { EventsOn } from "../../wailsjs/runtime/runtime.js";
    import logoDark from "../assets/images/logo_dark.png";
    import logoLight from "../assets/images/logo_light.png";
    import ProfileModal from "./ProfileModal.svelte";
    import JoinNodeModal from "./JoinNodeModal.svelte";
    import CreateServerModal from "./CreateServerModal.svelte";
    import WidgetContainer from "./WidgetContainer.svelte";

    let currentPeerNodes = [];
    let pollInterval;
    let localNodeMetadata = null;

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
                      name:
                          localNodeMetadata?.name ||
                          $auth.currentUser.username ||
                          "My Server",
                      icon: (
                          localNodeMetadata?.name ||
                          $auth.currentUser.username ||
                          "Me"
                      )
                          .substring(0, 2)
                          .toUpperCase(),
                      type: "node",
                      avatar:
                          localNodeMetadata?.icon || $auth.currentUser.avatar,
                      color: stringToColor(
                          localNodeMetadata?.name ||
                              $auth.currentUser.username ||
                              "Me",
                      ),
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
            // Use GetPublicNodes (currently same as Discovered)
            const peers = (await GetPublicNodes()) || [];
            const peerNodes = peers.map((p) => ({
                id: p.user_id,
                name: p.server_name || p.username,
                icon: (p.server_name || p.username)
                    .substring(0, 2)
                    .toUpperCase(),
                avatar: p.server_icon || p.avatar,
                type: "node",
                ip: p.ip,
                port: p.port,
                status: p.status || "offline",
                statusMessage: p.status_message,
                color: stringToColor(p.server_name || p.username),
            }));

            currentPeerNodes = peerNodes;
            appContext.setPeers(peers);
        } catch (e) {
            console.error("Failed to fetch peers:", e);
        }
    }

    async function fetchLocalMetadata() {
        try {
            const meta = await GetLocalNodeMetadata();
            if (meta && meta.name) {
                localNodeMetadata = meta;
            }
        } catch (e) {
            console.error("Failed to fetch local metadata:", e);
        }
    }

    async function handleNodeClick(node) {
        if (node.id === "home") {
            appContext.setActivePeer(null);
        } else if (node.type === "channel" || node.isLocal) {
            appContext.setActivePeer(node.id);
        } else {
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

    function openJoinNodeModal() {
        modal.open(JoinNodeModal, {});
    }

    function openCreateServerModal() {
        modal.open(CreateServerModal, {});
    }

    onMount(() => {
        fetchPeers();
        fetchLocalMetadata();
        pollInterval = setInterval(fetchPeers, 5000);

        EventsOn("localNodeUpdated", (meta) => {
            console.log("Local Node Updated:", meta);
            localNodeMetadata = meta;
        });
    });

    onDestroy(() => {
        if (pollInterval) clearInterval(pollInterval);
    });
</script>

<WidgetContainer title="Nodes" className="h-full">
    <div
        class="flex flex-col gap-2 p-2 overflow-y-auto h-full content-start items-center"
    >
        {#each nodes as node}
            {@const isActive =
                (node.id === "home" && !activeId) || activeId === node.id}

            <button
                class="group relative flex flex-col items-center justify-center w-full aspect-square transition-all duration-200 outline-none"
                on:click={() => handleNodeClick(node)}
                on:contextmenu={(e) => handleContextMenu(e, node)}
                title={node.name}
            >
                <!-- Icon/Avatar -->
                <div
                    class="w-full h-full flex items-center justify-center overflow-hidden border transition-all duration-200
                    {isActive
                        ? 'border-[var(--brand-primary)] rounded-sm shadow-[0_0_10px_var(--brand-primary)]'
                        : 'border-[var(--border-subtle)] rounded-sm hover:border-[var(--text-muted)]'}"
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
                            class="text-lg font-bold text-[var(--text-normal)]"
                        >
                            {node.icon}
                        </span>
                    {/if}
                </div>

                <!-- Status Indicator (Dot) -->
                {#if node.type === "node" && !node.isLocal}
                    <div
                        class="absolute bottom-0 right-0 w-2.5 h-2.5 rounded-full border border-[var(--background-primary)]
                        {node.status === 'online'
                            ? 'bg-[var(--status-online)]'
                            : node.status === 'idle'
                              ? 'bg-[var(--status-idle)]'
                              : node.status === 'dnd'
                                ? 'bg-[var(--status-dnd)]'
                                : 'bg-[var(--status-offline)]'}"
                    ></div>
                {/if}
            </button>
        {/each}

        <!-- Create Server Button -->
        <button
            class="flex items-center justify-center w-full aspect-square border border-dashed border-[var(--border-subtle)] rounded-sm text-[var(--text-muted)] hover:text-[var(--brand-primary)] hover:border-[var(--brand-primary)] transition-all duration-200"
            on:click={openCreateServerModal}
            title="Create Server"
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
            >
                <line x1="12" y1="5" x2="12" y2="19"></line>
                <line x1="5" y1="12" x2="19" y2="12"></line>
            </svg>
        </button>

        <!-- Join Node Button -->
        <button
            class="flex items-center justify-center w-full aspect-square border border-dashed border-[var(--border-subtle)] rounded-sm text-[var(--text-muted)] hover:text-[var(--brand-primary)] hover:border-[var(--brand-primary)] transition-all duration-200"
            on:click={openJoinNodeModal}
            title="Join Node"
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
            >
                <path
                    d="M10 13a5 5 0 0 0 7.54.54l3-3a5 5 0 0 0-7.07-7.07l-1.72 1.71"
                ></path>
                <path
                    d="M14 11a5 5 0 0 0-7.54-.54l-3 3a5 5 0 0 0 7.07 7.07l1.71-1.71"
                ></path>
            </svg>
        </button>
    </div>
</WidgetContainer>

<script>
    import WidgetContainer from "./WidgetContainer.svelte";

    let updates = [
        {
            appId: 1091500,
            name: "Cyberpunk 2077",
            changeNumber: 21456789,
            lastUpdated: "2 hours ago",
            branch: "public",
            status: "ok",
        },
        {
            appId: 1245620,
            name: "Elden Ring",
            changeNumber: 21456123,
            lastUpdated: "5 hours ago",
            branch: "public",
            status: "ok",
        },
        {
            appId: 1172470,
            name: "Apex Legends",
            changeNumber: 21459999,
            lastUpdated: "10 mins ago",
            branch: "staging",
            status: "updating",
        },
        {
            appId: 427520,
            name: "Factorio",
            changeNumber: 21450000,
            lastUpdated: "1 day ago",
            branch: "experimental",
            status: "ok",
        },
        {
            appId: 570,
            name: "Dota 2",
            changeNumber: 21460001,
            lastUpdated: "Just now",
            branch: "public",
            status: "ok",
        },
    ];
</script>

<WidgetContainer title="DATABASE UPDATES" className="h-full">
    <div
        class="h-full overflow-y-auto custom-scrollbar bg-[var(--background-tertiary)] font-mono text-xs"
    >
        <table class="w-full text-left border-collapse">
            <thead
                class="bg-[var(--background-secondary)] text-[var(--text-muted)] sticky top-0 z-10"
            >
                <tr>
                    <th
                        class="p-2 border-b border-[var(--border-subtle)] font-normal"
                        >AppID</th
                    >
                    <th
                        class="p-2 border-b border-[var(--border-subtle)] font-normal"
                        >Name</th
                    >
                    <th
                        class="p-2 border-b border-[var(--border-subtle)] font-normal hidden xl:table-cell"
                        >Change #</th
                    >
                    <th
                        class="p-2 border-b border-[var(--border-subtle)] font-normal text-right"
                        >Updated</th
                    >
                </tr>
            </thead>
            <tbody class="divide-y divide-[var(--border-subtle)]">
                {#each updates as update}
                    <tr
                        class="hover:bg-[var(--background-modifier-hover)] transition-colors cursor-pointer group"
                    >
                        <td class="p-2 text-[var(--brand-primary)]"
                            >{update.appId}</td
                        >
                        <td class="p-2 text-[var(--text-normal)]">
                            <div class="flex items-center gap-2">
                                <span class="truncate max-w-[100px]"
                                    >{update.name}</span
                                >
                                {#if update.branch !== "public"}
                                    <span
                                        class="px-1 py-0.5 rounded-[2px] bg-[var(--background-modifier-selected)] text-[var(--text-muted)] text-[10px] uppercase"
                                        >{update.branch}</span
                                    >
                                {/if}
                            </div>
                        </td>
                        <td
                            class="p-2 text-[var(--text-muted)] hidden xl:table-cell"
                            >{update.changeNumber}</td
                        >
                        <td class="p-2 text-right">
                            <span
                                class={update.lastUpdated === "Just now"
                                    ? "text-[var(--status-online)]"
                                    : "text-[var(--text-muted)]"}
                            >
                                {update.lastUpdated}
                            </span>
                        </td>
                    </tr>
                {/each}
            </tbody>
        </table>

        <div
            class="p-2 text-[10px] text-[var(--text-muted)] text-center opacity-50 mt-4"
        >
            Connected to SteamDB (Mock) • v3.4.1
        </div>
    </div>
</WidgetContainer>

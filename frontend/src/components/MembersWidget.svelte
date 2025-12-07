<script>
    import { onMount, onDestroy } from "svelte";
    import { GetGuildMembers } from "../../wailsjs/go/main/App";
    import { EventsOn } from "../../wailsjs/runtime/runtime";
    import { appContext } from "../stores";
    import WidgetContainer from "./WidgetContainer.svelte";

    let members = [];
    let activePeerId;

    // Subscribe to activePeerId to fetch members when guild changes
    $: {
        activePeerId = $appContext.activePeerId;
        fetchMembers();
    }

    async function fetchMembers() {
        if (!activePeerId || activePeerId === "home") {
            members = [];
            return;
        }

        try {
            // GetGuildMembers returns []*proto.Member
            members = await GetGuildMembers(activePeerId);
        } catch (e) {
            console.error("Failed to fetch members:", e);
        }
    }

    onMount(() => {
        fetchMembers();

        // Listen for real-time updates
        EventsOn("GUILD_MEMBERS_UPDATED", (guildId) => {
            if (guildId === activePeerId) {
                fetchMembers();
            }
        });
    });
</script>

<WidgetContainer title="MEMBERS">
    <div slot="content" class="h-full overflow-y-auto p-2 space-y-1">
        {#if members.length === 0}
            <div class="text-xs text-[var(--text-muted)] text-center py-4">
                No other members connected.
            </div>
        {:else}
            {#each members as member}
                <div
                    class="flex items-center gap-2 p-1.5 hover:bg-[var(--background-modifier-hover)] rounded-sm cursor-pointer group"
                >
                    <div
                        class="w-8 h-8 rounded-sm bg-[var(--brand-primary)] flex items-center justify-center text-white font-bold text-xs shrink-0"
                    >
                        {(member.username || member.user_id)
                            .substring(0, 2)
                            .toUpperCase()}
                    </div>
                    <div class="overflow-hidden">
                        <div
                            class="text-sm font-bold text-[var(--text-normal)] truncate group-hover:text-[var(--header-primary)]"
                        >
                            {member.username || "Unknown"}
                        </div>
                        <div
                            class="text-[10px] text-[var(--text-muted)] truncate font-mono opacity-50"
                        >
                            {member.user_id.substring(0, 8)}...
                        </div>
                    </div>
                </div>
            {/each}
        {/if}
    </div>
</WidgetContainer>

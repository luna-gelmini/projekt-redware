<script>
    import { voice, appContext } from "../stores";
    import { webRTCService } from "../utils/webrtc";
    import { onDestroy } from "svelte";

    let isMuted = false;
    let isDeafened = false;

    function toggleMute() {
        isMuted = !isMuted;
        webRTCService.setMute(isMuted);
    }

    function toggleDeaf() {
        isDeafened = !isDeafened;
        webRTCService.setDeaf(isDeafened);
    }

    function disconnect() {
        webRTCService.leaveChannel();
        voice.update((v) => ({ ...v, channelId: null, connected: false }));
    }

    // Find channel name
    $: channel = $appContext.channels?.find(
        (c) => c.channel_id === $voice.channelId,
    );
    $: channelName = channel ? channel.channel_id : "Voice Channel";
</script>

<div
    class="bg-[var(--background-secondary)] border-t border-[var(--border-subtle)] p-2"
>
    <div class="flex items-center justify-between">
        <div class="flex flex-col min-w-0 mr-2">
            <div
                class="flex items-center gap-1 text-[var(--status-online)] font-bold text-xs uppercase tracking-wide"
            >
                <span class="w-2 h-2 rounded-full bg-[var(--status-online)]"
                ></span>
                Voice Connected
            </div>
            <div class="text-[var(--text-normal)] text-sm font-medium truncate">
                {channelName} / Server
            </div>
        </div>

        <div class="flex items-center gap-1">
            <button
                class="p-2 rounded-md hover:bg-[var(--background-modifier-hover)] {isMuted
                    ? 'text-red-400'
                    : 'text-[var(--text-normal)]'}"
                on:click={toggleMute}
                title={isMuted ? "Unmute" : "Mute"}
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
                    {#if isMuted}
                        <line x1="1" y1="1" x2="23" y2="23"></line>
                        <path
                            d="M9 9v3a3 3 0 0 0 5.12 2.12M15 9.34V4a3 3 0 0 0-5.94-.6"
                        ></path>
                        <path
                            d="M17 16.95A7 7 0 0 1 5 12v-2m14 0v2a7 7 0 0 1-.11 1.23"
                        ></path>
                        <line x1="12" y1="19" x2="12" y2="23"></line>
                        <line x1="8" y1="23" x2="16" y2="23"></line>
                    {:else}
                        <path
                            d="M12 1a3 3 0 0 0-3 3v8a3 3 0 0 0 6 0V4a3 3 0 0 0-3-3z"
                        ></path>
                        <path d="M19 10v2a7 7 0 0 1-14 0v-2"></path>
                        <line x1="12" y1="19" x2="12" y2="23"></line>
                        <line x1="8" y1="23" x2="16" y2="23"></line>
                    {/if}
                </svg>
            </button>

            <button
                class="p-2 rounded-md hover:bg-[var(--background-modifier-hover)] text-[var(--text-normal)]"
                on:click={disconnect}
                title="Disconnect"
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
                    <path d="M18.36 6.64a9 9 0 1 1-12.73 0"></path>
                    <line x1="12" y1="2" x2="12" y2="12"></line>
                </svg>
            </button>
        </div>
    </div>
</div>

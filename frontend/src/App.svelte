<script>
  import { onMount } from "svelte";
  import LoginView from "./components/LoginView.svelte";
  import LayoutEditor from "./components/LayoutEditor.svelte";
  import {
    isAuthenticated,
    appContext,
    modal,
    appearance,
    voice,
  } from "./stores";
  import { EventsOn } from "../wailsjs/runtime/runtime.js";
  import VoiceConnectionPanel from "./components/VoiceConnectionPanel.svelte";
  import { webRTCService } from "./utils/webrtc";
  import { GetDiscoveredPeers } from "../wailsjs/go/main/App";

  // --- Theme Engine ---
  $: {
    const root = document.documentElement;
    const theme = $appearance.theme;

    if (theme === "light") {
      root.style.setProperty("--background-primary", "#ffffff");
      root.style.setProperty("--background-secondary", "#f2f3f5");
      root.style.setProperty("--background-tertiary", "#e3e5e8");
      root.style.setProperty("--background-floating", "#ffffff");
      root.style.setProperty("--header-primary", "#060607");
      root.style.setProperty("--header-secondary", "#4f5660");
      root.style.setProperty("--text-normal", "#313338");
      root.style.setProperty("--text-muted", "#5c5e66");
      root.style.setProperty("--border-subtle", "#e3e5e8");
      root.style.setProperty("--channeltextarea-background", "#e3e5e8");
    } else if (theme === "oled") {
      root.style.setProperty("--background-primary", "#000000");
      root.style.setProperty("--background-secondary", "#000000");
      root.style.setProperty("--background-tertiary", "#000000");
      root.style.setProperty("--background-floating", "#000000");
      root.style.setProperty("--header-primary", "#ffffff");
      root.style.setProperty("--header-secondary", "#b9bbbe");
      root.style.setProperty("--text-normal", "#ffffff");
      root.style.setProperty("--text-muted", "#a0a0a0");
      root.style.setProperty("--border-subtle", "#333333");
      root.style.setProperty("--channeltextarea-background", "#1a1a1a");
    } else {
      // Midnight (Default)
      root.style.setProperty("--background-primary", "#313338");
      root.style.setProperty("--background-secondary", "#2b2d31");
      root.style.setProperty("--background-tertiary", "#1e1f22");
      root.style.setProperty("--background-floating", "#18191c");
      root.style.setProperty("--header-primary", "#f2f3f5");
      root.style.setProperty("--header-secondary", "#b9bbbe");
      root.style.setProperty("--text-normal", "#dbdee1");
      root.style.setProperty("--text-muted", "#949ba4");
      root.style.setProperty("--border-subtle", "rgba(255,255,255,0.05)");
      root.style.setProperty("--channeltextarea-background", "#383a40");
    }

    // Accent Color
    root.style.setProperty("--brand-primary", $appearance.accentColor);
    // Font Size
    root.style.setProperty("font-size", `${$appearance.fontSize}px`);
  }

  onMount(() => {
    // Inject Custom CSS
    const style = document.createElement("style");
    style.id = "redware-custom-css";
    document.head.appendChild(style);

    appearance.subscribe((settings) => {
      style.textContent = settings.customCSS;
    });

    // Listen for backend debug logs
    EventsOn("DEBUG_LOG", (msg) => {
      console.log("%c[BACKEND]", "color: #ff00ff; font-weight: bold;", msg);
    });

    EventsOn("WEBRTC_SIGNAL", (signal) => {
      webRTCService.handleSignal(signal);
    });

    // Global Peer Polling
    const fetchPeers = async () => {
      try {
        const peers = await GetDiscoveredPeers();
        if (peers) {
          appContext.setPeers(peers);
        }
      } catch (err) {
        console.error("Failed to fetch peers:", err);
      }
    };

    fetchPeers();
    const interval = setInterval(fetchPeers, 5000);

    return () => {
      clearInterval(interval);
    };
  });
</script>

<main
  class="w-screen h-screen bg-[var(--background-tertiary)] text-[var(--text-normal)] overflow-hidden font-sans selection:bg-[var(--brand-experiment)] selection:text-white flex"
>
  {#if $isAuthenticated}
    <div class="flex-1 flex flex-col min-h-0 overflow-hidden">
      <LayoutEditor />
      {#if $voice.connected}
        <VoiceConnectionPanel />
      {/if}
    </div>
  {:else}
    <LoginView />
  {/if}

  <!-- Global Modal Container -->
  {#if $modal.isOpen}
    <svelte:component
      this={$modal.component}
      {...$modal.props}
      on:close={modal.close}
    />
  {/if}
</main>

<style>
  :global(body) {
    margin: 0;
    padding: 0;
    background-color: var(--background-tertiary);
    font-family: "gg sans", "Noto Sans", "Helvetica Neue", Helvetica, Arial,
      sans-serif;
  }
</style>

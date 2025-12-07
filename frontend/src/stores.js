import { writable, derived } from 'svelte/store';
import { Log } from '../wailsjs/go/main/App';

// --- Auth Store ---
function createAuthStore() {
    const { subscribe, set, update } = writable({
        isAuthenticated: false,
        token: null,
        currentUser: null,
    });

    return {
        subscribe,
        login: (token, username, avatar) => {
            // In a real app, decode JWT to get user ID/details
            set({
                isAuthenticated: true,
                token,
                currentUser: { id: token, username, avatar }
            });
            // Save to localStorage if needed (Wails handles persistence differently usually, but for now in-memory)
        },
        logout: () => set({ isAuthenticated: false, token: null, currentUser: null }),
        updateAvatar: (avatar) => update(s => ({
            ...s,
            currentUser: { ...s.currentUser, avatar }
        })),
    };
}
export const auth = createAuthStore();
export const isAuthenticated = derived(auth, $auth => $auth.isAuthenticated);

// --- App Context Store ---
// Holds global app state like active peer, sidebar state, etc.
function createAppContextStore() {
    const { subscribe, update } = writable({
        activePeerId: null, // Primary Sidebar Selection (Server or Home)
        activeChannelId: null, // Secondary Sidebar Selection (Channel or DM)
        isSidebarOpen: true,
        messages: {}, // Map<peerId, Message[]>
        peers: [], // List of online peers
        serverName: "",
        serverIcon: "",
        serverDescription: "",
    });

    return {
        subscribe,
        setActivePeer: (peerId) => update(s => {
            console.log("Store: setActivePeer called with:", peerId);
            // When switching servers, clear active channel?
            // Or keep it if we want to remember state per server?
            // For now, clear it to avoid confusion.
            return { ...s, activePeerId: peerId, activeChannelId: null };
        }),
        setActiveChannel: (channelId) => update(s => {
            console.log("Store: setActiveChannel called with:", channelId);
            return { ...s, activeChannelId: channelId };
        }),
        toggleSidebar: () => update(s => ({ ...s, isSidebarOpen: !s.isSidebarOpen })),
        addMessage: (peerId, message) => update(s => {
            const peerMessages = s.messages[peerId] || [];
            return {
                ...s,
                messages: {
                    ...s.messages,
                    [peerId]: [...peerMessages, message]
                }
            };
        }),
        setPeers: (peers) => update(s => ({ ...s, peers })),
        setServerMetadata: (name, icon, description) => update(s => ({
            ...s,
            serverName: name,
            serverIcon: icon,
            serverDescription: description
        })),
    };
}
export const appContext = createAppContextStore();

// --- Appearance Store ---
function createAppearanceStore() {
    const { subscribe, update } = writable({
        theme: 'dark', // 'dark', 'light', 'oled'
        density: 'comfortable', // 'comfortable', 'compact'
        fontSize: 16,
        accentColor: '#5865F2', // Default blurple, user can change
        customCSS: '', // User-defined CSS overrides
    });

    return {
        subscribe,
        setTheme: (theme) => update(s => ({ ...s, theme })),
        setDensity: (density) => update(s => ({ ...s, density })),
        setFontSize: (size) => update(s => ({ ...s, fontSize: size })),
        setAccentColor: (color) => update(s => ({ ...s, accentColor: color })),
        setCustomCSS: (css) => update(s => ({ ...s, customCSS: css })),
    };
}
export const appearance = createAppearanceStore();

// --- Modal Store ---
function createModalStore() {
    const { subscribe, set } = writable({
        isOpen: false,
        component: null,
        props: {}
    });

    return {
        subscribe,
        open: (component, props = {}) => set({ isOpen: true, component, props }),
        close: () => set({ isOpen: false, component: null, props: {} })
    };
}
export const modal = createModalStore();

// --- Voice Store ---
export const voice = writable({
    channelId: null,
    connected: false,
    peers: []
});

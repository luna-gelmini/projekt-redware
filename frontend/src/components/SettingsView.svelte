<script>
    import { appearance, modal, auth, appContext } from "../stores";
    import { UpdateAvatar } from "../../wailsjs/go/main/App.js";
    import Cropper from "svelte-easy-crop";
    import { getCroppedImg } from "../utils/canvasUtils"; // We need to create this

    // Available themes
    const themes = [
        { id: "dark", name: "Midnight", color: "#1e1e2e" },
        { id: "light", name: "Daylight", color: "#ffffff" },
        { id: "oled", name: "Void", color: "#000000" },
    ];

    // Accent colors
    const accents = [
        { id: "#5865F2", name: "Blurple" },
        { id: "#23a559", name: "Emerald" },
        { id: "#eb459e", name: "Fuchsia" },
        { id: "#f0b232", name: "Gold" },
        { id: "#ed4245", name: "Crimson" },
    ];

    const cssPlaceholder = "/* .sidebar { width: 100px !important; } */";

    let crop = { x: 0, y: 0 };
    let zoom = 1;
    let image = null;
    let croppedAreaPixels = null;
    let isCropping = false;

    function handleClose() {
        modal.close();
    }

    function onCropComplete(e) {
        console.log("Crop complete:", e.detail);
        croppedAreaPixels = e.detail.pixels;
    }

    async function handleFileChange(e) {
        const file = e.target.files[0];
        if (!file) return;
        const reader = new FileReader();
        reader.onload = (event) => {
            image = /** @type {string} */ (event.target.result);
            isCropping = true;
        };
        reader.readAsDataURL(file);
    }

    async function saveCroppedImage() {
        console.log("Saving cropped image...", { image, croppedAreaPixels });
        if (!image || !croppedAreaPixels) {
            console.error("Missing image or crop data");
            return;
        }
        try {
            const croppedImage = await getCroppedImg(image, croppedAreaPixels);
            console.log("Image cropped successfully, uploading...");
            await UpdateAvatar(croppedImage);
            auth.updateAvatar(croppedImage);
            isCropping = false;
            image = null;
        } catch (e) {
            console.error("Failed to crop/upload:", e);
        }
    }

    function cancelCrop() {
        isCropping = false;
        image = null;
    }
</script>

{#if isCropping}
    <div class="fixed inset-0 z-[200] bg-black flex flex-col">
        <div class="relative flex-1 bg-black">
            <Cropper
                {image}
                bind:crop
                bind:zoom
                aspect={1}
                on:cropcomplete={onCropComplete}
            />
        </div>
        <div
            class="h-20 bg-[var(--background-tertiary)] flex items-center justify-between px-8 border-t border-[var(--border-subtle)]"
        >
            <div class="flex items-center gap-4">
                <span class="text-sm font-medium text-[var(--text-muted)]"
                    >Zoom</span
                >
                <input
                    type="range"
                    min="1"
                    max="3"
                    step="0.1"
                    bind:value={zoom}
                    class="w-32 h-1 bg-[var(--background-secondary)] rounded-lg appearance-none cursor-pointer accent-[var(--brand-primary)]"
                />
            </div>
            <div class="flex gap-4">
                <button
                    on:click={cancelCrop}
                    class="px-4 py-2 text-sm font-medium text-[var(--text-normal)] hover:text-[var(--text-header)] transition-colors"
                >
                    Cancel
                </button>
                <button
                    on:click={saveCroppedImage}
                    class="px-6 py-2 text-sm font-bold text-white bg-[var(--brand-primary)] rounded-lg hover:bg-[var(--brand-hover)] transition-colors shadow-lg shadow-indigo-500/20"
                >
                    Save
                </button>
            </div>
        </div>
    </div>
{/if}

<div
    class="fixed inset-0 z-[100] flex items-center justify-center bg-black/60 backdrop-blur-sm"
    on:click|self={handleClose}
    on:keydown={(e) => e.key === "Escape" && handleClose()}
    role="button"
    tabindex="0"
>
    <div
        class="w-full max-w-2xl bg-[var(--background-primary)] rounded-sm shadow-[0_0_30px_rgba(0,0,0,0.6)] border border-[var(--border-subtle)] overflow-hidden flex flex-col max-h-[85vh]"
    >
        <!-- Header -->
        <div
            class="h-14 px-6 flex items-center justify-between border-b border-[var(--border-subtle)] bg-[var(--background-tertiary)]/80 backdrop-blur-md"
        >
            <h2
                class="text-sm font-bold uppercase tracking-wider text-[var(--header-primary)]"
            >
                Interface Settings
            </h2>
            <button
                on:click={handleClose}
                class="p-1.5 rounded-sm hover:bg-[var(--background-modifier-hover)] text-[var(--text-muted)] hover:text-[var(--text-normal)] transition-colors"
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
                    ><line x1="18" y1="6" x2="6" y2="18"></line><line
                        x1="6"
                        y1="6"
                        x2="18"
                        y2="18"
                    ></line></svg
                >
            </button>
        </div>

        <!-- Content -->
        <div class="flex-1 overflow-y-auto p-6 space-y-8 custom-scrollbar">
            <!-- Node Settings Removed as per user request -->

            <!-- Profile Picture Section -->
            <section class="space-y-3">
                <h3
                    class="text-xs font-bold uppercase tracking-wider text-[var(--header-secondary)] border-b border-[var(--border-subtle)] pb-1"
                >
                    Profile Picture
                </h3>
                <div class="flex items-center gap-4">
                    <div
                        class="relative group w-20 h-20 rounded-sm overflow-hidden bg-[var(--background-secondary)] border border-[var(--border-subtle)]"
                    >
                        {#if $auth.currentUser && $auth.currentUser.avatar}
                            <img
                                src={$auth.currentUser.avatar}
                                alt="Avatar"
                                class="w-full h-full object-cover"
                            />
                        {:else}
                            <div
                                class="w-full h-full flex items-center justify-center text-2xl font-bold text-[var(--text-muted)]"
                            >
                                {$auth.currentUser?.username?.[0]?.toUpperCase() ||
                                    "?"}
                            </div>
                        {/if}
                        <label
                            class="absolute inset-0 flex items-center justify-center bg-black/60 opacity-0 group-hover:opacity-100 transition-opacity cursor-pointer text-white text-[10px] font-bold uppercase tracking-widest"
                        >
                            CHANGE
                            <input
                                type="file"
                                accept="image/*"
                                class="hidden"
                                on:change={handleFileChange}
                            />
                        </label>
                    </div>
                    <div class="text-xs text-[var(--text-muted)]">
                        Click to upload a new profile picture.<br />
                        Recommended size: 256x256.
                    </div>
                </div>
            </section>
            <!-- Theme Section -->
            <section class="space-y-3">
                <h3
                    class="text-xs font-bold uppercase tracking-wider text-[var(--header-secondary)] border-b border-[var(--border-subtle)] pb-1"
                >
                    Theme
                </h3>
                <div class="grid grid-cols-3 gap-3">
                    {#each themes as theme}
                        <button
                            class="relative h-20 rounded-sm border transition-all overflow-hidden group
                            {$appearance.theme === theme.id
                                ? 'border-[var(--brand-primary)] shadow-[0_0_10px_rgba(var(--brand-primary-rgb),0.3)]'
                                : 'border-[var(--border-subtle)] hover:border-[var(--text-muted)]'}"
                            style="background-color: {theme.color}"
                            on:click={() => appearance.setTheme(theme.id)}
                        >
                            <div
                                class="absolute inset-0 flex items-center justify-center"
                            >
                                <span
                                    class="text-xs font-bold uppercase tracking-wide {theme.id ===
                                    'light'
                                        ? 'text-black'
                                        : 'text-white'}"
                                    style="color: {theme.id === 'light'
                                        ? 'black'
                                        : 'white'} !important"
                                    >{theme.name}</span
                                >
                            </div>
                            {#if $appearance.theme === theme.id}
                                <div
                                    class="absolute top-1 right-1 w-4 h-4 bg-[var(--brand-primary)] rounded-sm flex items-center justify-center text-white"
                                >
                                    <svg
                                        width="10"
                                        height="10"
                                        viewBox="0 0 24 24"
                                        fill="none"
                                        stroke="currentColor"
                                        stroke-width="4"
                                        stroke-linecap="round"
                                        stroke-linejoin="round"
                                        ><polyline points="20 6 9 17 4 12"
                                        ></polyline></svg
                                    >
                                </div>
                            {/if}
                        </button>
                    {/each}
                </div>
            </section>

            <!-- Accent Color Section -->
            <section class="space-y-3">
                <h3
                    class="text-xs font-bold uppercase tracking-wider text-[var(--header-secondary)] border-b border-[var(--border-subtle)] pb-1"
                >
                    Accent Color
                </h3>
                <div class="flex flex-wrap gap-3">
                    {#each accents as color}
                        <button
                            class="w-10 h-10 rounded-sm transition-transform hover:scale-105 focus:outline-none ring-offset-2 ring-offset-[var(--background-primary)]
                            {$appearance.accentColor === color.id
                                ? 'ring-2 ring-[var(--text-normal)] scale-105'
                                : ''}"
                            style="background-color: {color.id}"
                            on:click={() => appearance.setAccentColor(color.id)}
                            title={color.name}
                        >
                        </button>
                    {/each}

                    <!-- Custom Picker Placeholder -->
                    <div
                        class="w-10 h-10 rounded-sm border border-dashed border-[var(--text-muted)] flex items-center justify-center text-[var(--text-muted)] hover:text-[var(--text-normal)] hover:border-[var(--text-normal)] cursor-pointer transition-colors"
                    >
                        <svg
                            width="16"
                            height="16"
                            viewBox="0 0 24 24"
                            fill="none"
                            stroke="currentColor"
                            stroke-width="2"
                            stroke-linecap="round"
                            stroke-linejoin="round"
                            ><line x1="12" y1="5" x2="12" y2="19"></line><line
                                x1="5"
                                y1="12"
                                x2="19"
                                y2="12"
                            ></line></svg
                        >
                    </div>
                </div>
            </section>

            <!-- Density Section -->
            <section class="space-y-3">
                <h3
                    class="text-xs font-bold uppercase tracking-wider text-[var(--header-secondary)] border-b border-[var(--border-subtle)] pb-1"
                >
                    Density
                </h3>
                <div
                    class="flex bg-[var(--background-secondary)] p-1 rounded-sm border border-[var(--border-subtle)]"
                >
                    <button
                        class="flex-1 py-1.5 rounded-sm text-xs font-bold uppercase tracking-wide transition-all
                        {$appearance.density === 'comfortable'
                            ? 'bg-[var(--background-primary)] text-[var(--text-normal)] shadow-sm'
                            : 'text-[var(--text-muted)] hover:text-[var(--text-normal)]'}"
                        on:click={() => appearance.setDensity("comfortable")}
                    >
                        Comfortable
                    </button>
                    <button
                        class="flex-1 py-1.5 rounded-sm text-xs font-bold uppercase tracking-wide transition-all
                        {$appearance.density === 'compact'
                            ? 'bg-[var(--background-primary)] text-[var(--text-normal)] shadow-sm'
                            : 'text-[var(--text-muted)] hover:text-[var(--text-normal)]'}"
                        on:click={() => appearance.setDensity("compact")}
                    >
                        Compact
                    </button>
                </div>
            </section>

            <!-- Font Size Section -->
            <section class="space-y-3">
                <h3
                    class="text-xs font-bold uppercase tracking-wider text-[var(--header-secondary)] border-b border-[var(--border-subtle)] pb-1"
                >
                    Font Size ({$appearance.fontSize}px)
                </h3>
                <input
                    type="range"
                    min="12"
                    max="24"
                    step="1"
                    value={$appearance.fontSize}
                    on:input={(e) =>
                        appearance.setFontSize(parseInt(e.target.value))}
                    class="w-full h-1 bg-[var(--background-secondary)] rounded-sm appearance-none cursor-pointer accent-[var(--brand-primary)]"
                />
                <div
                    class="flex justify-between text-[10px] text-[var(--text-muted)] font-mono"
                >
                    <span>12px</span>
                    <span>24px</span>
                </div>
            </section>

            <!-- Advanced: Custom CSS -->
            <section class="space-y-3">
                <div
                    class="flex items-center justify-between border-b border-[var(--border-subtle)] pb-1"
                >
                    <h3
                        class="text-xs font-bold uppercase tracking-wider text-[var(--header-secondary)]"
                    >
                        Advanced: Custom CSS
                    </h3>
                    <span class="text-[10px] text-[var(--text-muted)]"
                        >Override anything.</span
                    >
                </div>
                <textarea
                    class="w-full h-48 bg-[var(--background-secondary)] text-[var(--text-normal)] font-mono text-xs p-3 rounded-sm outline-none border border-[var(--border-subtle)] focus:border-[var(--brand-primary)] resize-none placeholder-[var(--text-muted)]/50"
                    placeholder={cssPlaceholder}
                    value={$appearance.customCSS}
                    on:input={(e) =>
                        appearance.setCustomCSS(e.currentTarget.value)}
                ></textarea>
            </section>
        </div>
    </div>
</div>

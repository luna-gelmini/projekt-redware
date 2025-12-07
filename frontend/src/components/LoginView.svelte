<script>
    import { auth, appearance } from "../stores";
    import { Login, Register, StartP2P } from "../../wailsjs/go/main/App";
    import logoDark from "../assets/images/logo_dark.png";
    import logoLight from "../assets/images/logo_light.png";

    let isRegistering = false;
    let username = "";
    let password = "";
    let error = "";
    let isLoading = false;

    async function handleSubmit() {
        if (!username || !password) {
            error = "Please fill in all fields.";
            return;
        }
        isLoading = true;
        error = "";

        try {
            if (isRegistering) {
                const success = await Register(username, password);
                if (success) {
                    // Auto login after register
                    const loginResult = await Login(username, password);
                    if (loginResult && loginResult.token) {
                        auth.login(
                            loginResult.token,
                            username,
                            loginResult.avatar,
                        );
                        console.log("Attempting to start P2P...");
                        const p2pStarted = await StartP2P();
                        console.log("P2P Started:", p2pStarted);
                    } else {
                        console.log("Attempting to start P2P (fallback)...");
                        await StartP2P();
                        error = "Registration successful, but login failed.";
                    }
                } else {
                    error = "Registration failed. Username might be taken.";
                }
            } else {
                const result = await Login(username, password);
                if (result && result.token) {
                    auth.login(result.token, username, result.avatar);
                    // Start P2P
                    await StartP2P();
                } else {
                    error = result.error || "Login failed";
                }
            }
        } catch (e) {
            error = "An error occurred: " + e;
        } finally {
            isLoading = false;
        }
    }

    function toggleMode() {
        isRegistering = !isRegistering;
        error = "";
    }
</script>

<div
    class="flex items-center justify-center h-screen bg-[var(--background-tertiary)]"
>
    <div
        class="w-full max-w-md p-8 space-y-6 bg-[var(--background-primary)] rounded-xl shadow-2xl border border-[var(--border-subtle)]"
    >
        <div class="text-center space-y-2 flex flex-col items-center">
            <img
                src={$appearance.theme === "light" ? logoLight : logoDark}
                alt="Projekt Redware Logo"
                class="w-20 h-20 mb-4 object-contain"
            />
            <h1
                class="text-3xl font-bold tracking-tight text-[var(--header-primary)]"
            >
                {isRegistering ? "Create Account" : "Welcome Back"}
            </h1>
            <p class="text-[var(--text-muted)]">
                {isRegistering ? "Join the network." : "Enter Projekt Redware."}
            </p>
        </div>

        <form on:submit|preventDefault={handleSubmit} class="space-y-4">
            <div class="space-y-2">
                <label
                    for="username"
                    class="text-sm font-medium text-[var(--header-secondary)]"
                    >Username</label
                >
                <input
                    id="username"
                    type="text"
                    bind:value={username}
                    class="w-full px-4 py-2 bg-[var(--background-secondary)] border border-[var(--border-subtle)] rounded-lg focus:outline-none focus:ring-2 focus:ring-[var(--brand-primary)] text-[var(--text-normal)] placeholder-[var(--text-muted)] transition-all"
                    placeholder="username"
                />
            </div>
            <div class="space-y-2">
                <label
                    for="password"
                    class="text-sm font-medium text-[var(--header-secondary)]"
                    >Password</label
                >
                <input
                    id="password"
                    type="password"
                    bind:value={password}
                    class="w-full px-4 py-2 bg-[var(--background-secondary)] border border-[var(--border-subtle)] rounded-lg focus:outline-none focus:ring-2 focus:ring-[var(--brand-primary)] text-[var(--text-normal)] placeholder-[var(--text-muted)] transition-all"
                    placeholder="••••••••"
                />
            </div>

            {#if error}
                <div
                    class="p-3 text-sm text-red-400 bg-red-900/20 rounded-lg border border-red-900/50"
                >
                    {error}
                </div>
            {/if}

            <button
                type="submit"
                disabled={isLoading}
                class="w-full py-2.5 text-sm font-semibold text-white bg-[var(--brand-primary)] rounded-lg hover:bg-[var(--brand-hover)] focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-[var(--brand-primary)] disabled:opacity-50 disabled:cursor-not-allowed transition-all shadow-lg shadow-indigo-500/20"
            >
                {isLoading
                    ? "Processing..."
                    : isRegistering
                      ? "Sign Up"
                      : "Sign In"}
            </button>
        </form>

        <div class="text-center">
            <button
                on:click={toggleMode}
                class="text-sm text-[var(--text-link)] hover:underline hover:text-[var(--text-link-hover)] transition-colors"
            >
                {isRegistering
                    ? "Already have an account? Sign in"
                    : "Need an account? Sign up"}
            </button>
        </div>
    </div>
</div>

<style>
    /* Minimalist styles handled by Tailwind classes and CSS variables in style.css */
</style>

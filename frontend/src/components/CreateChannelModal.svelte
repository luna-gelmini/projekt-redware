<script>
    import { CreateChannel } from "../../wailsjs/go/main/App";
    import { modal, appContext } from "../stores";

    let channelName = "";
    let error = "";

    async function handleCreate() {
        if (!channelName) {
            error = "Channel name is required";
            return;
        }
        if (!channelName.startsWith("#")) {
            channelName = "#" + channelName;
        }

        try {
            await CreateChannel(channelName);
            modal.close();
            appContext.setActivePeer(channelName);
        } catch (e) {
            error = "Error: " + e;
        }
    }

    function handleKeydown(e) {
        if (e.key === "Enter") {
            handleCreate();
        }
    }
</script>

<div class="p-6 bg-gray-800 rounded-lg shadow-xl w-96 border border-gray-700">
    <h2 class="text-2xl font-bold mb-4 text-white">Create Channel</h2>

    {#if error}
        <div class="bg-red-500/20 text-red-200 p-3 rounded mb-4 text-sm">
            {error}
        </div>
    {/if}

    <div class="mb-4">
        <label
            class="block text-gray-400 text-sm font-bold mb-2"
            for="channelName"
        >
            Channel Name
        </label>
        <input
            id="channelName"
            type="text"
            bind:value={channelName}
            on:keydown={handleKeydown}
            placeholder="#general"
            class="w-full bg-gray-900 text-white border border-gray-600 rounded p-2 focus:outline-none focus:border-blue-500 transition-colors"
            autoFocus
        />
    </div>

    <div class="flex justify-end gap-2">
        <button
            on:click={() => modal.close()}
            class="px-4 py-2 text-gray-300 hover:text-white transition-colors"
        >
            Cancel
        </button>
        <button
            on:click={handleCreate}
            class="px-4 py-2 bg-blue-600 hover:bg-blue-500 text-white rounded font-medium transition-colors"
        >
            Create Channel
        </button>
    </div>
</div>

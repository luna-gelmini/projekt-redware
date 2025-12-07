<script>
  import { onMount } from "svelte";
  import { GetActiveNow } from "../../wailsjs/go/main/App";

  const openPopout = (userId, targetRect) => {
    console.log(`TODO: Open popout for user ${userId}`, targetRect);
  };

  let activities = [];
  let loading = true;

  onMount(async () => {
    try {
      activities = await GetActiveNow();
    } catch (err) {
      console.error(err);
    } finally {
      loading = false;
    }
  });
</script>

<div>
  <h2 class="text-lg font-bold text-[var(--header-primary)] mb-2">
    Active Now
  </h2>
  {#if loading}
    <p class="text-[var(--text-muted)]">Loading activities...</p>
  {:else if activities.length === 0}
    <div class="p-4 bg-[var(--background-secondary)] rounded-lg text-center">
      <p class="font-semibold text-[var(--header-primary)]">
        It's quiet... for now
      </p>
      <p class="text-sm text-[var(--text-muted)] mt-1">
        When a friend starts an activity, we'll show it here!
      </p>
    </div>
  {:else}
    <div class="space-y-3">
      {#each activities as activity, index (index)}
        <div
          class="flex items-start p-3 rounded-lg bg-[var(--background-modifier-hover)] hover:bg-[var(--background-modifier-active)] transition-colors duration-150 group"
        >
          <img
            src={activity.imageUrl}
            alt={activity.game}
            class="w-20 h-20 object-cover rounded-lg flex-shrink-0"
          />
          <div class="ml-4 overflow-hidden flex-grow">
            <p
              class="font-semibold text-[var(--header-primary)] truncate group-hover:underline cursor-pointer"
              on:click={(e) =>
                openPopout(
                  activity.user.id,
                  e.currentTarget.getBoundingClientRect(),
                )}
            >
              {activity.user.name}
            </p>
            <p class="text-sm text-[var(--text-normal)] truncate">
              {activity.game}
            </p>
            <p class="text-sm text-[var(--text-muted)] truncate">
              {activity.details}
            </p>
          </div>
          <button
            class="ml-4 text-xs font-semibold bg-[var(--background-secondary-alt)] hover:bg-[var(--background-accent)] text-[var(--text-normal)] px-3 py-1.5 rounded self-center flex-shrink-0"
          >
            Join
          </button>
        </div>
      {/each}
    </div>
  {/if}
</div>

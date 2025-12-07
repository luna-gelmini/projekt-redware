<script>
  export let embedData;
</script>

<div
  class="mt-2 bg-[var(--background-secondary)] border-l-4 border-blue-500 rounded p-3 max-w-lg"
>
  {#if embedData.provider}
    <div class="text-xs text-[var(--text-muted)] mb-1">
      {embedData.provider.name}
    </div>
  {/if}
  <a
    href={embedData.url}
    target="_blank"
    rel="noopener noreferrer"
    class="text-[var(--text-link)] font-semibold hover:underline"
  >
    {embedData.title}
  </a>
  <p class="text-sm text-[var(--text-normal)] mt-1">{embedData.description}</p>
  {#if embedData.thumbnail}
    <img
      src={embedData.thumbnail.url}
      alt={embedData.title}
      class="mt-2 rounded-md max-h-60"
    />
  {/if}
</div>

<script lang="ts">
  let query = "";
  let searching = false;
  let results: any = null;
  let error = "";

  async function handleSearch() {
    if (!query.trim()) {
      error = "Please enter a search query";
      return;
    }

    searching = true;
    error = "";
    results = null;

    try {
      const response = await fetch(`/api/search?q=${encodeURIComponent(query)}`);
      
      if (!response.ok) {
        const errorText = await response.text();
        throw new Error(errorText || "Search failed");
      }

      results = await response.json();
      
      if (!results.documents || !results.documents[0] || results.documents[0].length === 0) {
        error = "No results found";
        results = null;
      }
    } catch (err) {
      error = err instanceof Error ? err.message : "Search failed";
      results = null;
    } finally {
      searching = false;
    }
  }

  function handleKeyPress(event: KeyboardEvent) {
    if (event.key === "Enter") {
      handleSearch();
    }
  }
</script>

<div class="bg-white rounded-2xl shadow-lg p-8 border border-indigo-100">
  <h2 class="text-2xl font-bold text-slate-800 mb-6 flex items-center gap-2">
    <svg class="w-6 h-6 text-indigo-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
    </svg>
    Search Documents
  </h2>

  <div class="space-y-6">
    <!-- Search Input -->
    <div class="flex gap-3">
      <input
        type="text"
        bind:value={query}
        on:keypress={handleKeyPress}
        placeholder="Enter your search query..."
        class="flex-1 px-4 py-3 border border-slate-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-indigo-500 focus:border-transparent"
      />
      <button
        on:click={handleSearch}
        disabled={searching || !query.trim()}
        class="bg-gradient-to-r from-indigo-600 to-purple-600 text-white font-semibold px-8 py-3 rounded-lg hover:from-indigo-700 hover:to-purple-700 disabled:opacity-50 disabled:cursor-not-allowed transition-all duration-200 shadow-md hover:shadow-lg"
      >
        {#if searching}
          <svg class="animate-spin h-5 w-5" fill="none" viewBox="0 0 24 24">
            <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
            <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
          </svg>
        {:else}
          Search
        {/if}
      </button>
    </div>

    <!-- Error Message -->
    {#if error}
      <div class="p-4 rounded-lg bg-amber-50 border border-amber-200">
        <p class="text-sm font-medium text-amber-800">{error}</p>
      </div>
    {/if}

    <!-- Results -->
    {#if results && results.documents && results.documents[0]}
      <div class="space-y-4">
        <div class="flex items-center justify-between">
          <h3 class="text-lg font-semibold text-slate-800">
            Found {results.documents[0].length} result{results.documents[0].length !== 1 ? 's' : ''}
          </h3>
        </div>

        <div class="space-y-3">
          {#each results.documents[0] as doc, i}
            <div class="p-5 bg-gradient-to-br from-slate-50 to-indigo-50/30 rounded-lg border border-slate-200 hover:border-indigo-300 transition-colors">
              <div class="flex items-start justify-between mb-3">
                <div class="flex items-center gap-2">
                  <span class="inline-flex items-center justify-center w-7 h-7 bg-indigo-600 text-white text-xs font-bold rounded-full">
                    {i + 1}
                  </span>
                  {#if results.metadatas && results.metadatas[0] && results.metadatas[0][i]}
                    <div class="flex items-center gap-2 text-xs text-slate-600">
                      <span class="font-medium">{results.metadatas[0][i].filename || 'Unknown'}</span>
                      {#if results.metadatas[0][i].chunk_num}
                        <span class="text-slate-400">•</span>
                        <span class="bg-slate-200 px-2 py-0.5 rounded">Chunk {results.metadatas[0][i].chunk_num}</span>
                      {/if}
                    </div>
                  {/if}
                </div>
                {#if results.distances && results.distances[0] && results.distances[0][i] !== undefined}
                  <span class="text-xs font-semibold text-indigo-600 bg-indigo-100 px-2 py-1 rounded">
                    Score: {(1 - results.distances[0][i]).toFixed(3)}
                  </span>
                {/if}
              </div>
              <p class="text-sm text-slate-700 leading-relaxed">{doc}</p>
            </div>
          {/each}
        </div>
      </div>
    {/if}
  </div>
</div>

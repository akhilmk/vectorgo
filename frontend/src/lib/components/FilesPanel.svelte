<script lang="ts">
  import { onMount } from "svelte";
  import { api } from "../api";
  import type { StatsResult } from "../api";

  let stats: StatsResult | null = null;
  let loading = false;
  let error = "";
  let showDeleteDialog = false;
  let fileToDelete = "";
  let deleting = false;

  async function loadStats() {
    loading = true;
    error = "";
    try {
      stats = await api.getStats();
    } catch (err) {
      error = err instanceof Error ? err.message : "Failed to load stats";
      stats = null;
    } finally {
      loading = false;
    }
  }

  onMount(() => {
    loadStats();
  });

  function openDeleteDialog(filename: string) {
    fileToDelete = filename;
    showDeleteDialog = true;
  }

  function closeDeleteDialog() {
    showDeleteDialog = false;
    fileToDelete = "";
  }

  async function handleDeleteFile() {
    deleting = true;
    try {
      await api.deleteFile(fileToDelete);
      showDeleteDialog = false;
      fileToDelete = "";
      await loadStats(); // Reload stats after delete
    } catch (err) {
      error = err instanceof Error ? err.message : "Delete failed";
    } finally {
      deleting = false;
    }
  }

  // Export refresh function so parent can call it
  export function refresh() {
    loadStats();
  }
</script>

<div class="space-y-6">
  <!-- Stats Cards -->
  <div class="grid grid-cols-2 gap-4">
    <div class="bg-gradient-to-br from-indigo-50 to-indigo-100/50 rounded-xl p-5 border border-indigo-200">
      <div class="flex items-center gap-3">
        <div class="w-12 h-12 bg-indigo-600 rounded-lg flex items-center justify-center">
          <svg class="w-6 h-6 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
          </svg>
        </div>
        <div>
          <p class="text-sm font-medium text-indigo-700">Total Files</p>
          <p class="text-3xl font-bold text-indigo-900">{stats?.total_files || 0}</p>
        </div>
      </div>
    </div>

    <div class="bg-gradient-to-br from-purple-50 to-purple-100/50 rounded-xl p-5 border border-purple-200">
      <div class="flex items-center gap-3">
        <div class="w-12 h-12 bg-purple-600 rounded-lg flex items-center justify-center">
          <svg class="w-6 h-6 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 11H5m14 0a2 2 0 012 2v6a2 2 0 01-2 2H5a2 2 0 01-2-2v-6a2 2 0 012-2m14 0V9a2 2 0 00-2-2M5 11V9a2 2 0 012-2m0 0V5a2 2 0 012-2h6a2 2 0 012 2v2M7 7h10" />
          </svg>
        </div>
        <div>
          <p class="text-sm font-medium text-purple-700">Total Chunks</p>
          <p class="text-3xl font-bold text-purple-900">{stats?.total_chunks || 0}</p>
        </div>
      </div>
    </div>
  </div>

  {#if loading}
    <div class="flex items-center justify-center py-12">
      <svg class="animate-spin h-8 w-8 text-indigo-600" fill="none" viewBox="0 0 24 24">
        <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
        <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
      </svg>
    </div>
  {:else if error}
    <div class="p-4 rounded-lg bg-red-50 border border-red-200">
      <p class="text-sm font-medium text-red-800">{error}</p>
    </div>
  {:else if stats}
    <!-- Files List -->
    <div class="bg-white rounded-xl border border-slate-200 overflow-hidden">
      <div class="px-6 py-4 bg-slate-50 border-b border-slate-200">
        <h3 class="text-lg font-semibold text-slate-800">Uploaded Files</h3>
      </div>

      {#if stats.files.length > 0}
        <div class="divide-y divide-slate-200">
          {#each stats.files as file}
            <div class="flex items-center justify-between p-4 hover:bg-slate-50 transition-colors">
              <div class="flex items-center gap-3 flex-1 min-w-0">
                <svg class="w-5 h-5 text-slate-500 flex-shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M7 21h10a2 2 0 002-2V9.414a1 1 0 00-.293-.707l-5.414-5.414A1 1 0 0012.586 3H7a2 2 0 00-2 2v14a2 2 0 002 2z" />
                </svg>
                <div class="flex-1 min-w-0">
                  <span class="text-sm font-medium text-slate-700 truncate block">{file}</span>
                  <span class="text-xs text-slate-500">
                    {stats.file_chunk_counts[file] || 0} chunk{stats.file_chunk_counts[file] !== 1 ? 's' : ''}
                  </span>
                </div>
              </div>
              <button
                on:click={() => openDeleteDialog(file)}
                class="flex-shrink-0 ml-4 text-red-600 hover:text-red-700 hover:bg-red-50 p-2 rounded-lg transition-colors"
                title="Delete file"
              >
                <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
                </svg>
              </button>
            </div>
          {/each}
        </div>
      {:else}
        <div class="text-center py-12">
          <svg class="w-16 h-16 text-slate-300 mx-auto mb-3" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
          </svg>
          <p class="text-slate-500 text-sm">No files uploaded yet</p>
        </div>
      {/if}
    </div>

  {/if}
</div>

<!-- Delete File Confirmation Dialog -->
{#if showDeleteDialog}
  <div class="fixed inset-0 bg-black/50 flex items-center justify-center z-50 p-4">
    <div class="bg-white rounded-2xl shadow-2xl max-w-md w-full p-6">
      <div class="flex items-center gap-3 mb-4">
        <div class="w-12 h-12 bg-red-100 rounded-full flex items-center justify-center">
          <svg class="w-6 h-6 text-red-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z" />
          </svg>
        </div>
        <h3 class="text-xl font-bold text-slate-800">Delete File</h3>
      </div>

      <p class="text-slate-600 mb-4">
        Are you sure you want to delete <strong>{fileToDelete}</strong>? This will remove all chunks associated with this file.
      </p>

      <p class="text-sm text-slate-500 mb-6">This action cannot be undone.</p>

      <div class="flex gap-3">
        <button
          on:click={closeDeleteDialog}
          disabled={deleting}
          class="flex-1 bg-slate-200 text-slate-700 font-semibold py-2.5 px-4 rounded-lg hover:bg-slate-300 disabled:opacity-50 transition-colors"
        >
          Cancel
        </button>
        <button
          on:click={handleDeleteFile}
          disabled={deleting}
          class="flex-1 bg-red-600 text-white font-semibold py-2.5 px-4 rounded-lg hover:bg-red-700 disabled:opacity-50 transition-colors flex items-center justify-center gap-2"
        >
          {#if deleting}
            <svg class="animate-spin h-5 w-5" fill="none" viewBox="0 0 24 24">
              <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
              <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
            </svg>
            Deleting...
          {:else}
            Delete File
          {/if}
        </button>
      </div>
    </div>
  </div>
{/if}



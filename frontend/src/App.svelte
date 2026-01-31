<script lang="ts">
  import { onMount } from "svelte";
  import UploadPanel from "./lib/components/UploadPanel.svelte";
  import SearchPanel from "./lib/components/SearchPanel.svelte";
  import FilesPanel from "./lib/components/FilesPanel.svelte";
  import Login from "./lib/components/Login.svelte";
  import { api } from "./lib/api";
  import { uploadStore } from "./lib/uploadStore";
  import type { UploadState } from "./lib/uploadStore";

  let loggedIn = false;
  let filesPanel: any;
  let activeTab: "upload" | "search" | "files" = "upload";
  let uploadState: UploadState;

  uploadStore.subscribe(state => {
    uploadState = state;
  });

  onMount(() => {
    loggedIn = api.isLoggedIn();
  });

  function handleLoginSuccess() {
    loggedIn = true;
  }

  function handleLogout() {
    api.logout();
    loggedIn = false;
  }

  function handleUploadComplete() {
    // Refresh files panel after successful upload
    if (filesPanel) {
      filesPanel.refresh();
    }
  }

  function setTab(tab: "upload" | "search" | "files") {
    activeTab = tab;
  }

  function handleCancelUpload() {
    uploadStore.cancelUpload();
  }
</script>

{#if !loggedIn}
  <Login onLoginSuccess={handleLoginSuccess} />
{:else}
  <main class="min-h-screen bg-gradient-to-br from-slate-50 via-indigo-50/30 to-purple-50/20">
    <!-- Header -->
    <div class="bg-white/80 backdrop-blur-sm border-b border-indigo-100 sticky top-0 z-10">
      <div class="max-w-5xl mx-auto px-6 py-4">
        <div class="flex items-center justify-between gap-3">
          <div class="flex items-center gap-3">
            <div class="w-10 h-10 bg-gradient-to-br from-indigo-600 to-purple-600 rounded-xl flex items-center justify-center shadow-lg">
              <svg class="w-6 h-6 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
              </svg>
            </div>
            <div>
              <h1 class="text-2xl font-black text-transparent bg-clip-text bg-gradient-to-r from-indigo-600 to-purple-600 tracking-tight">
                VectorGo
              </h1>
              <p class="text-xs text-slate-500 font-medium">PDF Vector Search & Embedding</p>
            </div>
          </div>

          <button
            on:click={handleLogout}
            class="text-sm font-semibold text-slate-600 hover:text-indigo-600 py-2 px-4 rounded-lg hover:bg-indigo-50 transition-colors"
          >
            Logout
          </button>
        </div>
      </div>
    </div>

    <!-- Global Upload Status Bar -->
    {#if uploadState?.uploading}
      <div class="bg-indigo-600 text-white sticky top-[73px] z-10 shadow-lg">
        <div class="max-w-5xl mx-auto px-6 py-3">
          <div class="flex items-center justify-between gap-4">
            <div class="flex items-center gap-3 flex-1 min-w-0">
              <svg class="w-5 h-5 animate-spin flex-shrink-0" fill="none" viewBox="0 0 24 24">
                <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
                <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
              </svg>
              <div class="flex-1 min-w-0">
                <p class="text-sm font-semibold truncate">Uploading: {uploadState.fileName}</p>
                <p class="text-xs text-indigo-100 truncate">{uploadState.progress}</p>
              </div>
            </div>
            <button
              on:click={handleCancelUpload}
              class="flex-shrink-0 bg-white/20 hover:bg-white/30 text-white font-semibold py-2 px-4 rounded-lg transition-colors text-sm"
            >
              Cancel
            </button>
          </div>
        </div>
      </div>
    {/if}

    <!-- Main Content -->
    <div class="max-w-5xl mx-auto px-6 py-8">
      <!-- Tab Navigation -->
      <div class="mb-6">
        <div class="border-b border-slate-200">
          <nav class="-mb-px flex space-x-8" aria-label="Tabs">
            <button
              on:click={() => setTab("upload")}
              class="
                {activeTab === 'upload' 
                  ? 'border-indigo-500 text-indigo-600' 
                  : 'border-transparent text-slate-500 hover:text-slate-700 hover:border-slate-300'}
                whitespace-nowrap py-4 px-1 border-b-2 font-semibold text-sm transition-colors flex items-center gap-2
              "
            >
              <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M7 16a4 4 0 01-.88-7.903A5 5 0 1115.9 6L16 6a5 5 0 011 9.9M15 13l-3-3m0 0l-3 3m3-3v12" />
              </svg>
              Upload & Process
            </button>

            <button
              on:click={() => setTab("search")}
              class="
                {activeTab === 'search' 
                  ? 'border-indigo-500 text-indigo-600' 
                  : 'border-transparent text-slate-500 hover:text-slate-700 hover:border-slate-300'}
                whitespace-nowrap py-4 px-1 border-b-2 font-semibold text-sm transition-colors flex items-center gap-2
              "
            >
              <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
              </svg>
              Search
            </button>

            <button
              on:click={() => setTab("files")}
              class="
                {activeTab === 'files' 
                  ? 'border-indigo-500 text-indigo-600' 
                  : 'border-transparent text-slate-500 hover:text-slate-700 hover:border-slate-300'}
                whitespace-nowrap py-4 px-1 border-b-2 font-semibold text-sm transition-colors flex items-center gap-2
              "
            >
              <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
              </svg>
              Manage Files
            </button>
          </nav>
        </div>
      </div>

      <!-- Tab Content -->
      <div class="bg-white rounded-2xl shadow-lg p-8 border border-indigo-100">
        {#if activeTab === "upload"}
          <div>
            <h2 class="text-2xl font-bold text-slate-800 mb-2">Upload PDF Documents</h2>
            <p class="text-slate-600 mb-6">Select a PDF file and configure chunking parameters for processing.</p>
            <UploadPanel onUploadComplete={handleUploadComplete} />
          </div>
        {:else if activeTab === "search"}
          <div>
            <h2 class="text-2xl font-bold text-slate-800 mb-2">Search Documents</h2>
            <p class="text-slate-600 mb-6">Perform semantic search to find relevant content across all uploaded documents.</p>
            <SearchPanel />
          </div>
        {:else if activeTab === "files"}
          <div>
            <h2 class="text-2xl font-bold text-slate-800 mb-2">Manage Files</h2>
            <p class="text-slate-600 mb-6">View collection statistics and manage your uploaded files.</p>
            <FilesPanel bind:this={filesPanel} />
          </div>
        {/if}
      </div>

      <!-- Info Section -->
      <div class="mt-8 bg-white/50 backdrop-blur-sm rounded-xl p-6 border border-indigo-100">
        <h3 class="text-lg font-bold text-slate-800 mb-3">How It Works</h3>
        <div class="grid md:grid-cols-3 gap-4 text-sm">
          <div class="flex gap-3">
            <div class="flex-shrink-0 w-8 h-8 bg-indigo-100 rounded-lg flex items-center justify-center">
              <span class="text-indigo-600 font-bold">1</span>
            </div>
            <div>
              <p class="font-semibold text-slate-800">Upload PDF</p>
              <p class="text-slate-600">Select a PDF file and configure chunk size and stride parameters.</p>
            </div>
          </div>
          <div class="flex gap-3">
            <div class="flex-shrink-0 w-8 h-8 bg-purple-100 rounded-lg flex items-center justify-center">
              <span class="text-purple-600 font-bold">2</span>
            </div>
            <div>
              <p class="font-semibold text-slate-800">Process & Embed</p>
              <p class="text-slate-600">Text is chunked and embedded using Ollama, then stored in ChromaDB.</p>
            </div>
          </div>
          <div class="flex gap-3">
            <div class="flex-shrink-0 w-8 h-8 bg-indigo-100 rounded-lg flex items-center justify-center">
              <span class="text-indigo-600 font-bold">3</span>
            </div>
            <div>
              <p class="font-semibold text-slate-800">Search & Manage</p>
              <p class="text-slate-600">Search through documents and manage your file collection.</p>
            </div>
          </div>
        </div>
      </div>
    </div>
  </main>
{/if}

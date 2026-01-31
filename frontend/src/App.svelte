<script lang="ts">
  import UploadPanel from "./lib/components/UploadPanel.svelte";
  import SearchPanel from "./lib/components/SearchPanel.svelte";
</script>

<main class="min-h-screen bg-gradient-to-br from-slate-50 via-indigo-50/30 to-purple-50/20">
  <!-- Header -->
  <div class="bg-white/80 backdrop-blur-sm border-b border-indigo-100 sticky top-0 z-10">
    <div class="max-w-6xl mx-auto px-6 py-4">
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
    </div>
  </div>

  <!-- Main Content -->
  <div class="max-w-6xl mx-auto px-6 py-8">
    <div class="mb-6">
      <h2 class="text-3xl font-bold text-slate-800 mb-2">Document Processing & Search</h2>
      <p class="text-slate-600">Upload PDF files, configure chunking parameters, and search through your documents using semantic search.</p>
    </div>

    <div class="grid lg:grid-cols-2 gap-6">
      <!-- Upload Panel -->
      <UploadPanel />

      <!-- Search Panel -->
      <SearchPanel />
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
            <p class="font-semibold text-slate-800">Search</p>
            <p class="text-slate-600">Perform semantic search to find relevant content across all documents.</p>
          </div>
        </div>
      </div>
    </div>
  </div>
</main>

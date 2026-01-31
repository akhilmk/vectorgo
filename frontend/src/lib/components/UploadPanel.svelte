<script lang="ts">
  import { api } from "../api";
  import { uploadStore } from "../uploadStore";

  export let onUploadComplete: (() => void) | undefined = undefined;

  let file: File | null = null;
  let chunkSize = 100;
  let chunkStride = 80;
  let message = "";
  let messageType: "success" | "error" | "" = "";
  
  // Subscribe to upload store to get uploading state
  let uploading = false;
  uploadStore.subscribe(state => {
    uploading = state.uploading;
  });

  function handleFileChange(event: Event) {
    const target = event.target as HTMLInputElement;
    if (target.files && target.files[0]) {
      file = target.files[0];
      message = "";
    }
  }

  async function handleUpload() {
    if (!file) {
      message = "Please select a PDF file";
      messageType = "error";
      return;
    }

    if (chunkSize <= 0 || chunkStride <= 0) {
      message = "Chunk size and stride must be positive numbers";
      messageType = "error";
      return;
    }

    message = "";
    const controller = new AbortController();
    uploadStore.startUpload(file.name, controller);

    try {
      const formData = new FormData();
      formData.append("file", file);
      formData.append("chunkSize", chunkSize.toString());
      formData.append("chunkStride", chunkStride.toString());

      const result = await api.uploadPDF(
        formData,
        (status) => {
          uploadStore.updateProgress(status);
        },
        controller.signal
      );
      
      message = `Successfully processed ${result.filename}`;
      messageType = "success";
      uploadStore.completeUpload();
      file = null;
      
      // Reset file input
      const fileInput = document.getElementById("file-input") as HTMLInputElement;
      if (fileInput) fileInput.value = "";
      
      // Notify parent component
      if (onUploadComplete) {
        onUploadComplete();
      }
      
    } catch (error) {
      const errorMsg = error instanceof Error ? error.message : "Upload failed";
      message = errorMsg;
      messageType = "error";
      uploadStore.reset();
    }
  }
</script>

  <h2 class="hidden">Upload PDF</h2>

  <div class="space-y-6">
    <!-- File Input -->
    <div>
      <label for="file-input" class="block text-sm font-semibold text-slate-700 mb-2">
        Select PDF File
      </label>
      <input
        id="file-input"
        type="file"
        accept=".pdf"
        on:change={handleFileChange}
        disabled={uploading}
        class="block w-full text-sm text-slate-600 file:mr-4 file:py-2.5 file:px-4 file:rounded-lg file:border-0 file:text-sm file:font-semibold file:bg-indigo-50 file:text-indigo-700 hover:file:bg-indigo-100 cursor-pointer border border-slate-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-indigo-500 disabled:opacity-50 disabled:cursor-not-allowed"
      />
      {#if file}
        <p class="mt-2 text-sm text-slate-600">
          Selected: <span class="font-medium">{file.name}</span>
          <span class="text-slate-500">({(file.size / (1024 * 1024)).toFixed(2)} MB)</span>
        </p>
      {/if}
    </div>

    <!-- Chunk Configuration -->
    <div class="grid grid-cols-2 gap-4">
      <div>
        <label for="chunk-size" class="block text-sm font-semibold text-slate-700 mb-2">
          Chunk Size (words)
        </label>
        <input
          id="chunk-size"
          type="number"
          bind:value={chunkSize}
          disabled={uploading}
          min="10"
          max="1000"
          class="w-full px-4 py-2.5 border border-slate-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-indigo-500 focus:border-transparent disabled:opacity-50 disabled:cursor-not-allowed"
        />
        <p class="mt-1 text-xs text-slate-500">Number of words per chunk</p>
      </div>

      <div>
        <label for="chunk-stride" class="block text-sm font-semibold text-slate-700 mb-2">
          Chunk Stride (words)
        </label>
        <input
          id="chunk-stride"
          type="number"
          bind:value={chunkStride}
          disabled={uploading}
          min="1"
          max="1000"
          class="w-full px-4 py-2.5 border border-slate-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-indigo-500 focus:border-transparent disabled:opacity-50 disabled:cursor-not-allowed"
        />
        <p class="mt-1 text-xs text-slate-500">Step size between chunks (overlap = size - stride)</p>
      </div>
    </div>

    <!-- Upload Button -->
    <button
      on:click={handleUpload}
      disabled={!file || uploading}
      class="w-full bg-gradient-to-r from-indigo-600 to-purple-600 text-white font-semibold py-3 px-6 rounded-lg hover:from-indigo-700 hover:to-purple-700 disabled:opacity-50 disabled:cursor-not-allowed transition-all duration-200 shadow-md hover:shadow-lg"
    >
      {#if uploading}
        <span class="flex items-center justify-center gap-2">
          <svg class="animate-spin h-5 w-5" fill="none" viewBox="0 0 24 24">
            <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
            <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
          </svg>
          Processing...
        </span>
      {:else}
        Upload & Process
      {/if}
    </button>

    <!-- Message Display -->
    {#if message}
      <div class="p-4 rounded-lg {messageType === 'success' ? 'bg-green-50 border border-green-200' : 'bg-red-50 border border-red-200'}">
        <p class="text-sm font-medium {messageType === 'success' ? 'text-green-800' : 'text-red-800'}">
          {message}
        </p>
      </div>
    {/if}
  </div>

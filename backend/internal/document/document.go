package document

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/ledongthuc/pdf"
)

type Config struct {
	OllamaURL     string
	ChromaURL     string
	ChromaAPIBase string
	DefaultModel  string
	Collection    string
}

type Handler struct {
	config Config
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func NewHandler() *Handler {
	return &Handler{
		config: Config{
			OllamaURL:     getEnv("OLLAMA_URL", "http://localhost:11434"),
			ChromaURL:     getEnv("CHROMA_URL", "http://localhost:8000"),
			ChromaAPIBase: "/api/v2/tenants/default_tenant/databases/default_database/collections",
			DefaultModel:  getEnv("EMBEDDING_MODEL", "embeddinggemma:300m"),
			Collection:    getEnv("COLLECTION_NAME", "documents"),
		},
	}
}

func (h *Handler) RegisterRoutes(mux *http.ServeMux, mw func(http.HandlerFunc) http.HandlerFunc) {
	mux.HandleFunc("/api/reset", mw(h.HandleReset))
	mux.HandleFunc("/api/upload", mw(h.HandleUpload))
	mux.HandleFunc("/api/search", mw(h.HandleSearch))
	mux.HandleFunc("/api/stats", mw(h.HandleStats))
	mux.HandleFunc("/api/files/", mw(h.HandleDeleteFile))
}

// Request/Response Structs
type EmbeddingRequest struct {
	Model  string `json:"model"`
	Prompt string `json:"prompt"`
}

type EmbeddingResponse struct {
	Embedding []float32 `json:"embedding"`
}

type ChromaAddRequest struct {
	Documents  []string      `json:"documents"`
	Metadatas  []interface{} `json:"metadatas"`
	Ids        []string      `json:"ids"`
	Embeddings [][]float32   `json:"embeddings"`
}

type ChromaQueryRequest struct {
	QueryEmbeddings [][]float32 `json:"query_embeddings"`
	NResults        int         `json:"n_results"`
}

type ChromaQueryResponse struct {
	Ids       [][]string      `json:"ids"`
	Documents [][]string      `json:"documents"`
	Metadatas [][]interface{} `json:"metadatas"`
	Distances [][]float32     `json:"distances"`
}

type ChromaGetResponse struct {
	ID       string      `json:"id"`
	Name     string      `json:"name"`
	Metadata interface{} `json:"metadata"`
}

type ChromaCountResponse struct {
	Count int `json:"count"`
}

type ChromaDeleteRequest struct {
	Where map[string]interface{} `json:"where"`
}

type ChromaDeleteResponse struct {
	DeletedCount int `json:"deleted_count"`
}

type StatsResponse struct {
	TotalChunks     int            `json:"total_chunks"`
	TotalFiles      int            `json:"total_files"`
	Files           []string       `json:"files"`
	FileChunkCounts map[string]int `json:"file_chunk_counts"`
}

// Handlers

func (h *Handler) HandleReset(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost && r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	log.Printf("Resetting collection: %s", h.config.Collection)

	url := fmt.Sprintf("%s%s/%s", h.config.ChromaURL, h.config.ChromaAPIBase, h.config.Collection)
	req, err := http.NewRequest(http.MethodDelete, url, nil)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to create request: %v", err), http.StatusInternalServerError)
		return
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to delete collection: %v", err), http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNotFound {
		body, _ := io.ReadAll(resp.Body)
		http.Error(w, fmt.Sprintf("chroma reset error: %s", string(body)), http.StatusInternalServerError)
		return
	}

	log.Printf("Collection reset successful")
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "reset successful", "collection": h.config.Collection})
}

func (h *Handler) HandleUpload(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse multipart form
	err := r.ParseMultipartForm(32 << 20) // 32 MB max
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to parse form: %v", err), http.StatusBadRequest)
		return
	}

	// Get file
	file, header, err := r.FormFile("file")
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to get file: %v", err), http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Log upload start with file details
	log.Printf("[UPLOAD START] File: %s | Size: %d bytes (%.2f MB)",
		header.Filename, header.Size, float64(header.Size)/(1024*1024))

	// Get chunk parameters
	chunkSize := 100
	chunkStride := 80

	if cs := r.FormValue("chunkSize"); cs != "" {
		if parsed, err := strconv.Atoi(cs); err == nil && parsed > 0 {
			chunkSize = parsed
		}
	}

	if cst := r.FormValue("chunkStride"); cst != "" {
		if parsed, err := strconv.Atoi(cst); err == nil && parsed > 0 {
			chunkStride = parsed
		}
	}

	log.Printf("[UPLOAD CONFIG] File: %s | Chunk size: %d words | Stride: %d words | Overlap: %d words",
		header.Filename, chunkSize, chunkStride, chunkSize-chunkStride)

	// Save file temporarily
	tmpFile, err := os.CreateTemp("", "upload-*.pdf")
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to create temp file: %v", err), http.StatusInternalServerError)
		return
	}
	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()

	_, err = io.Copy(tmpFile, file)
	if err != nil {
		log.Printf("[UPLOAD ERROR] File: %s | Failed to save: %v", header.Filename, err)
		http.Error(w, fmt.Sprintf("failed to save file: %v", err), http.StatusInternalServerError)
		return
	}

	log.Printf("[UPLOAD SAVED] File: %s | Temp path: %s", header.Filename, tmpFile.Name())

	// Process PDF
	// Process PDF with progress updates
	w.Header().Set("Content-Type", "application/x-ndjson")
	w.Header().Set("Transfer-Encoding", "chunked")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming not supported", http.StatusInternalServerError)
		return
	}

	progressFunc := func(msg string) {
		json.NewEncoder(w).Encode(map[string]string{"status": msg})
		flusher.Flush()
	}

	err = h.processPDF(tmpFile.Name(), header.Filename, chunkSize, chunkStride, progressFunc)
	if err != nil {
		log.Printf("Error processing PDF: %v", err)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	log.Printf("[UPLOAD COMPLETE] File: %s | Processing finished successfully", header.Filename)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":      "completed",
		"filename":    header.Filename,
		"chunkSize":   chunkSize,
		"chunkStride": chunkStride,
	})
}

func (h *Handler) HandleSearch(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	query := r.URL.Query().Get("q")
	if query == "" {
		http.Error(w, "Missing query parameter 'q'", http.StatusBadRequest)
		return
	}

	log.Printf("Searching for: %s", query)

	embedding, err := h.getEmbedding(query)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to get embedding: %v", err), http.StatusInternalServerError)
		return
	}

	results, err := h.queryChroma(embedding)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to query chroma: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(results)
}

func (h *Handler) HandleStats(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	log.Printf("Fetching collection statistics")

	// Get or create collection to ensure it exists
	colID, err := h.getOrCreateCollection(h.config.Collection)
	if err != nil {
		log.Printf("Failed to get collection: %v", err)
		// Return empty stats if collection doesn't exist
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(StatsResponse{
			TotalChunks:     0,
			TotalFiles:      0,
			Files:           []string{},
			FileChunkCounts: make(map[string]int),
		})
		return
	}

	// Get collection count
	countURL := fmt.Sprintf("%s%s/%s/count", h.config.ChromaURL, h.config.ChromaAPIBase, colID)
	resp, err := http.Get(countURL)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to get count: %v", err), http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		http.Error(w, fmt.Sprintf("chroma count error: %s", string(body)), http.StatusInternalServerError)
		return
	}

	var count int
	if err := json.NewDecoder(resp.Body).Decode(&count); err != nil {
		http.Error(w, fmt.Sprintf("failed to decode count: %v", err), http.StatusInternalServerError)
		return
	}

	// Get all documents to extract unique filenames and count chunks per file
	files := []string{}
	fileChunkCounts := make(map[string]int)

	if count > 0 {

		getURL := fmt.Sprintf("%s%s/%s/get", h.config.ChromaURL, h.config.ChromaAPIBase, colID)

		// Request all metadata to find unique files and count chunks
		reqBody, _ := json.Marshal(map[string]interface{}{
			"limit":   count,
			"include": []string{"metadatas"},
		})

		getResp, err := http.Post(getURL, "application/json", bytes.NewBuffer(reqBody))
		if err == nil {
			defer getResp.Body.Close()
			if getResp.StatusCode == http.StatusOK {
				var data struct {
					Metadatas []map[string]interface{} `json:"metadatas"`
				}
				if err := json.NewDecoder(getResp.Body).Decode(&data); err == nil {
					fileSet := make(map[string]bool)
					for _, meta := range data.Metadatas {
						if filename, ok := meta["filename"].(string); ok {
							fileSet[filename] = true
							fileChunkCounts[filename]++
						}
					}
					for filename := range fileSet {
						files = append(files, filename)
					}
				}
			}
		}
	}

	log.Printf("Collection stats: %d chunks, %d files", count, len(files))
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(StatsResponse{
		TotalChunks:     count,
		TotalFiles:      len(files),
		Files:           files,
		FileChunkCounts: fileChunkCounts,
	})
}

func (h *Handler) HandleDeleteFile(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract filename from URL path
	filename := strings.TrimPrefix(r.URL.Path, "/api/files/")
	if filename == "" {
		http.Error(w, "Filename required", http.StatusBadRequest)
		return
	}

	log.Printf("Deleting file: %s", filename)

	// Get collection ID
	colID, err := h.getOrCreateCollection(h.config.Collection)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to get collection: %v", err), http.StatusInternalServerError)
		return
	}

	// Delete all chunks with matching filename
	reqBody, _ := json.Marshal(ChromaDeleteRequest{
		Where: map[string]interface{}{
			"filename": filename,
		},
	})

	url := fmt.Sprintf("%s%s/%s/delete", h.config.ChromaURL, h.config.ChromaAPIBase, colID)
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(reqBody))
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to delete: %v", err), http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		http.Error(w, fmt.Sprintf("chroma delete error: %s", string(body)), http.StatusInternalServerError)
		return
	}

	log.Printf("Successfully deleted file: %s", filename)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":   "deleted",
		"filename": filename,
	})
}

// Helpers

func (h *Handler) processPDF(path, filename string, chunkSize, chunkStride int, progress func(string)) error {
	log.Printf("[PDF PROCESSING START] File: %s | Path: %s", filename, path)

	if progress != nil {
		progress("Reading PDF file...")
	}

	content, err := ReadPDF(path, filename, progress)
	if err != nil {
		log.Printf("[PDF ERROR] File: %s | Failed to read: %v", filename, err)
		return fmt.Errorf("failed to read PDF: %v", err)
	}

	// Report extracted content size
	contentLen := len(content)
	trimmedLen := len(strings.TrimSpace(content))
	log.Printf("[PDF EXTRACTION] File: %s | Extracted: %d chars | Trimmed: %d chars",
		filename, contentLen, trimmedLen)

	if progress != nil {
		progress(fmt.Sprintf("Extracted %d characters from PDF", contentLen))
	}

	if trimmedLen == 0 {
		log.Printf("[PDF ERROR] File: %s | No text content extracted (possibly scanned/image-based PDF)", filename)
		return fmt.Errorf("no text content extracted from PDF (file might be scanned or image-based)")
	}

	if progress != nil {
		progress("Splitting text into chunks...")
	}

	chunks := ChunkText(content, chunkSize, chunkStride)
	log.Printf("[PDF CHUNKING] File: %s | Total chunks: %d | Chunk size: %d words | Stride: %d words",
		filename, len(chunks), chunkSize, chunkStride)

	if len(chunks) == 0 {
		log.Printf("[PDF ERROR] File: %s | Resulted in 0 chunks (text too short)", filename)
		return fmt.Errorf("resulted in 0 chunks (text might be too short)")
	}

	if progress != nil {
		progress(fmt.Sprintf("Created %d chunks - Starting embedding...", len(chunks)))
	}

	for i, chunk := range chunks {
		msg := fmt.Sprintf("Processing chunk %d/%d", i+1, len(chunks))
		if progress != nil {
			progress(msg)
		}
		log.Printf("[CHUNK PROCESSING] File: %s | Chunk: %d/%d | Length: %d chars",
			filename, i+1, len(chunks), len(chunk))

		embedding, err := h.getEmbedding(chunk)
		if err != nil {
			log.Printf("[CHUNK WARNING] File: %s | Chunk: %d/%d | Embedding failed: %v",
				filename, i+1, len(chunks), err)
			continue
		}

		err = h.addToChroma(chunk, embedding, filename, i+1)
		if err != nil {
			log.Printf("[CHUNK WARNING] File: %s | Chunk: %d/%d | Storage failed: %v",
				filename, i+1, len(chunks), err)
			continue
		}

		log.Printf("[CHUNK SUCCESS] File: %s | Stored chunk: %d/%d", filename, i+1, len(chunks))
	}

	log.Printf("[PDF PROCESSING COMPLETE] File: %s | Total chunks: %d", filename, len(chunks))
	return nil
}

func (h *Handler) getEmbedding(text string) ([]float32, error) {
	reqBody, _ := json.Marshal(EmbeddingRequest{
		Model:  h.config.DefaultModel,
		Prompt: text,
	})

	resp, err := http.Post(h.config.OllamaURL+"/api/embeddings", "application/json", bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, fmt.Errorf("http post error: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("ollama returned status %d: %s", resp.StatusCode, string(body))
	}

	var res EmbeddingResponse
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return res.Embedding, nil
}

func (h *Handler) addToChroma(text string, embedding []float32, filename string, chunkNum int) error {
	colID, err := h.getOrCreateCollection(h.config.Collection)
	if err != nil {
		return fmt.Errorf("getOrCreateCollection failed: %w", err)
	}

	id := uuid.New().String()
	reqBody, _ := json.Marshal(ChromaAddRequest{
		Documents: []string{text},
		Metadatas: []interface{}{map[string]interface{}{
			"source":      "pdf",
			"filename":    filename,
			"chunk_num":   chunkNum,
			"uploaded_at": time.Now().Format(time.RFC3339),
		}},
		Ids:        []string{id},
		Embeddings: [][]float32{embedding},
	})

	url := fmt.Sprintf("%s%s/%s/add", h.config.ChromaURL, h.config.ChromaAPIBase, colID)
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(reqBody))
	if err != nil {
		return fmt.Errorf("http post to %s failed: %w", url, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("chroma add returned status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

func (h *Handler) queryChroma(embedding []float32) (*ChromaQueryResponse, error) {
	colID, err := h.getOrCreateCollection(h.config.Collection)
	if err != nil {
		return nil, err
	}

	reqBody, _ := json.Marshal(ChromaQueryRequest{
		QueryEmbeddings: [][]float32{embedding},
		NResults:        5,
	})

	url := fmt.Sprintf("%s%s/%s/query", h.config.ChromaURL, h.config.ChromaAPIBase, colID)
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("chroma query returned status %d: %s", resp.StatusCode, string(body))
	}

	var res ChromaQueryResponse
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return nil, err
	}

	return &res, nil
}

func (h *Handler) getOrCreateCollection(name string) (string, error) {
	// 1. Try to get
	getURL := fmt.Sprintf("%s%s/%s", h.config.ChromaURL, h.config.ChromaAPIBase, name)
	resp, err := http.Get(getURL)
	if err == nil {
		defer resp.Body.Close()
		if resp.StatusCode == http.StatusOK {
			var res struct {
				ID string `json:"id"`
			}
			if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
				return "", fmt.Errorf("failed to decode get collection response: %w", err)
			}
			return res.ID, nil
		}
	}

	// 2. Create if not found or status not OK
	createURL := fmt.Sprintf("%s%s", h.config.ChromaURL, h.config.ChromaAPIBase)
	reqBody, _ := json.Marshal(map[string]string{"name": name})
	resp, err = http.Post(createURL, "application/json", bytes.NewBuffer(reqBody))
	if err != nil {
		return "", fmt.Errorf("failed to POST to %s: %w", createURL, err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return "", fmt.Errorf("create collection at %s returned status %d: %s", createURL, resp.StatusCode, string(body))
	}

	var res struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal(body, &res); err != nil {
		return "", fmt.Errorf("failed to decode create collection response: %w", err)
	}

	if res.ID == "" {
		return "", fmt.Errorf("received empty collection ID from ChromaDB")
	}

	return res.ID, nil
}

// ReadPDF extracts plain text from a PDF file at the given path.
func ReadPDF(path, filename string, progress func(string)) (string, error) {
	f, r, err := pdf.Open(path)
	if err != nil {
		log.Printf("[PDF OPEN ERROR] File: %s | Error: %v", filename, err)
		return "", err
	}
	defer f.Close()

	total := r.NumPage()
	log.Printf("[PDF READING] File: %s | Total pages: %d", filename, total)

	var buf bytes.Buffer

	for i := 1; i <= total; i++ {
		// Report progress more frequently for large PDFs
		if progress != nil {
			if total < 20 || i%5 == 0 || i == 1 || i == total {
				progress(fmt.Sprintf("Reading PDF page %d/%d", i, total))
			}
		}

		p := r.Page(i)
		if p.V.IsNull() {
			log.Printf("[PDF PAGE SKIP] File: %s | Page: %d/%d | Reason: null page", filename, i, total)
			continue
		}

		type pageResult struct {
			text string
			err  error
		}
		ch := make(chan pageResult, 1)
		go func() {
			text, err := p.GetPlainText(nil)
			ch <- pageResult{text, err}
		}()

		select {
		case res := <-ch:
			if res.err != nil {
				log.Printf("[PDF PAGE ERROR] File: %s | Page: %d/%d | Error: %v", filename, i, total, res.err)
				continue
			}
			buf.WriteString(res.text)
		case <-time.After(10 * time.Second):
			log.Printf("[PDF PAGE TIMEOUT] File: %s | Page: %d/%d | Skipping after 10s", filename, i, total)
			if progress != nil {
				progress(fmt.Sprintf("Skipped page %d (timeout)", i))
			}
			continue
		}
	}

	log.Printf("[PDF READING COMPLETE] File: %s | Pages processed: %d | Text length: %d chars",
		filename, total, buf.Len())
	return buf.String(), nil
}

// ChunkText splits the text into chunks of `size` words with a `stride`.
func ChunkText(text string, size int, stride int) []string {
	var chunks []string
	words := strings.Fields(text)
	if len(words) == 0 {
		return nil
	}

	for i := 0; i < len(words); i += stride {
		end := i + size
		if end > len(words) {
			end = len(words)
		}
		chunks = append(chunks, strings.Join(words[i:end], " "))
		if end == len(words) {
			break
		}
	}
	return chunks
}

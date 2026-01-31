package main

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

	"github.com/dslipak/pdf"
	"github.com/google/uuid"
)

var (
	OllamaURL     = getEnv("OLLAMA_URL", "http://localhost:11434")
	ChromaURL     = getEnv("CHROMA_URL", "http://localhost:8000")
	ChromaAPIBase = "/api/v2/tenants/default_tenant/databases/default_database/collections"
	DefaultModel  = getEnv("EMBEDDING_MODEL", "embeddinggemma:300m")
	Collection    = getEnv("COLLECTION_NAME", "documents")
	Port          = getEnv("PORT", "8080")
)

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

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

func main() {
	// Enable CORS for all routes
	http.HandleFunc("/api/health", corsMiddleware(handleTest))
	http.HandleFunc("/api/reset", corsMiddleware(handleReset))
	http.HandleFunc("/api/upload", corsMiddleware(handleUpload))
	http.HandleFunc("/api/search", corsMiddleware(handleSearch))

	// Serve frontend static files
	fs := http.FileServer(http.Dir("frontend/dist"))
	http.Handle("/", fs)

	log.Printf("VectorGo server starting on :%s...", Port)
	log.Printf("Ollama URL: %s", OllamaURL)
	log.Printf("ChromaDB URL: %s", ChromaURL)
	log.Printf("Embedding Model: %s", DefaultModel)
	log.Printf("Collection: %s", Collection)

	if err := http.ListenAndServe(":"+Port, nil); err != nil {
		log.Fatal(err)
	}
}

func corsMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next(w, r)
	}
}

func handleTest(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status":  "ok",
		"service": "VectorGo",
		"version": "1.0.0",
	})
}

func handleReset(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost && r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	log.Printf("Resetting collection: %s", Collection)

	url := fmt.Sprintf("%s%s/%s", ChromaURL, ChromaAPIBase, Collection)
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
	json.NewEncoder(w).Encode(map[string]string{"status": "reset successful", "collection": Collection})
}

func handleUpload(w http.ResponseWriter, r *http.Request) {
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

	log.Printf("Received file: %s (size: %d bytes)", header.Filename, header.Size)

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

	log.Printf("Processing with chunk size: %d, stride: %d", chunkSize, chunkStride)

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
		http.Error(w, fmt.Sprintf("failed to save file: %v", err), http.StatusInternalServerError)
		return
	}

	// Process PDF
	err = processPDF(tmpFile.Name(), header.Filename, chunkSize, chunkStride)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to process PDF: %v", err), http.StatusInternalServerError)
		return
	}

	log.Printf("Successfully processed: %s", header.Filename)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":      "completed",
		"filename":    header.Filename,
		"chunkSize":   chunkSize,
		"chunkStride": chunkStride,
	})
}

func handleSearch(w http.ResponseWriter, r *http.Request) {
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

	embedding, err := getEmbedding(query)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to get embedding: %v", err), http.StatusInternalServerError)
		return
	}

	results, err := queryChroma(embedding)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to query chroma: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(results)
}

func processPDF(path, filename string, chunkSize, chunkStride int) error {
	content, err := readPDF(path)
	if err != nil {
		return fmt.Errorf("failed to read PDF: %v", err)
	}

	log.Printf("Extracted %d characters from PDF", len(content))

	chunks := chunkText(content, chunkSize, chunkStride)
	log.Printf("Split PDF into %d chunks (size: %d words, stride: %d words)", len(chunks), chunkSize, chunkStride)

	for i, chunk := range chunks {
		log.Printf("Processing chunk %d/%d (length: %d chars)", i+1, len(chunks), len(chunk))

		embedding, err := getEmbedding(chunk)
		if err != nil {
			log.Printf("WARNING: failed to get embedding for chunk %d: %v", i+1, err)
			continue
		}

		err = addToChroma(chunk, embedding, filename, i+1)
		if err != nil {
			log.Printf("WARNING: failed to add chunk %d to chroma: %v", i+1, err)
			continue
		}

		log.Printf("Successfully stored chunk %d/%d", i+1, len(chunks))
	}

	log.Printf("Completed processing all %d chunks", len(chunks))
	return nil
}

func readPDF(path string) (string, error) {
	r, err := pdf.Open(path)
	if err != nil {
		return "", err
	}
	var buf bytes.Buffer
	b, err := r.GetPlainText()
	if err != nil {
		return "", err
	}
	_, err = io.Copy(&buf, b)
	if err != nil {
		return "", err
	}
	return buf.String(), nil
}

func chunkText(text string, size int, stride int) []string {
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

func getEmbedding(text string) ([]float32, error) {
	reqBody, _ := json.Marshal(EmbeddingRequest{
		Model:  DefaultModel,
		Prompt: text,
	})

	resp, err := http.Post(OllamaURL+"/api/embeddings", "application/json", bytes.NewBuffer(reqBody))
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

func addToChroma(text string, embedding []float32, filename string, chunkNum int) error {
	colID, err := getOrCreateCollection(Collection)
	if err != nil {
		return fmt.Errorf("getOrCreateCollection failed: %w", err)
	}

	id := uuid.New().String()
	reqBody, _ := json.Marshal(ChromaAddRequest{
		Documents: []string{text},
		Metadatas: []interface{}{map[string]interface{}{
			"source":    "pdf",
			"filename":  filename,
			"chunk_num": chunkNum,
		}},
		Ids:        []string{id},
		Embeddings: [][]float32{embedding},
	})

	url := fmt.Sprintf("%s%s/%s/add", ChromaURL, ChromaAPIBase, colID)
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

func queryChroma(embedding []float32) (*ChromaQueryResponse, error) {
	colID, err := getOrCreateCollection(Collection)
	if err != nil {
		return nil, err
	}

	reqBody, _ := json.Marshal(ChromaQueryRequest{
		QueryEmbeddings: [][]float32{embedding},
		NResults:        5,
	})

	url := fmt.Sprintf("%s%s/%s/query", ChromaURL, ChromaAPIBase, colID)
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

func getOrCreateCollection(name string) (string, error) {
	// 1. Try to get
	getURL := fmt.Sprintf("%s%s/%s", ChromaURL, ChromaAPIBase, name)
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
	createURL := fmt.Sprintf("%s%s", ChromaURL, ChromaAPIBase)
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

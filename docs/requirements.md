# VectorGo - Requirements Document

## Overview

VectorGo is a PDF processing and semantic search application that enables users to upload PDF documents, automatically chunk and embed the text content, and perform intelligent semantic searches across their document collection.

---

## Functional Requirements

### 1. PDF Upload and Processing

#### 1.1 File Upload
- **FR-1.1.1**: The system shall accept PDF file uploads through a web interface
- **FR-1.1.2**: The system shall validate that uploaded files are in PDF format
- **FR-1.1.3**: The system shall support PDF files up to 32MB in size
- **FR-1.1.4**: The system shall display the filename and size of the selected file before upload

#### 1.2 Text Extraction
- **FR-1.2.1**: The system shall extract plain text content from uploaded PDF files
- **FR-1.2.2**: The system shall log the number of characters extracted from each PDF
- **FR-1.2.3**: The system shall handle multi-page PDF documents

#### 1.3 Text Chunking
- **FR-1.3.1**: The system shall split extracted text into configurable chunks based on word count
- **FR-1.3.2**: The system shall support configurable chunk size (default: 100 words)
- **FR-1.3.3**: The system shall support configurable chunk stride/step size (default: 80 words)
- **FR-1.3.4**: The system shall create overlapping chunks when stride is less than chunk size
- **FR-1.3.5**: The system shall log the total number of chunks created for each document
- **FR-1.3.6**: The system shall accept chunk size values between 10 and 1000 words
- **FR-1.3.7**: The system shall accept chunk stride values between 1 and 1000 words

### 2. Embedding Generation

#### 2.1 Vector Embeddings
- **FR-2.1.1**: The system shall generate vector embeddings for each text chunk using Ollama
- **FR-2.1.2**: The system shall use the embeddinggemma:300m model by default
- **FR-2.1.3**: The system shall support configurable embedding models via environment variables
- **FR-2.1.4**: The system shall log progress for each chunk being embedded
- **FR-2.1.5**: The system shall handle embedding generation failures gracefully and continue processing remaining chunks

### 3. Vector Storage

#### 3.1 ChromaDB Integration
- **FR-3.1.1**: The system shall store generated embeddings in ChromaDB
- **FR-3.1.2**: The system shall create a collection if it doesn't exist
- **FR-3.1.3**: The system shall store document text alongside embeddings
- **FR-3.1.4**: The system shall store metadata including:
  - Source type (pdf)
  - Original filename
  - Chunk number
- **FR-3.1.5**: The system shall assign unique UUIDs to each stored chunk
- **FR-3.1.6**: The system shall support configurable collection names via environment variables

### 4. Semantic Search

#### 4.1 Query Processing
- **FR-4.1.1**: The system shall accept natural language search queries
- **FR-4.1.2**: The system shall generate embeddings for search queries using the same model as documents
- **FR-4.1.3**: The system shall perform vector similarity search in ChromaDB
- **FR-4.1.4**: The system shall return the top 5 most similar chunks by default
- **FR-4.1.5**: The system shall include relevance scores (distance metrics) in search results

#### 4.2 Search Results
- **FR-4.2.1**: The system shall display search results with:
  - Document text snippets
  - Source filename
  - Chunk number
  - Relevance score
- **FR-4.2.2**: The system shall rank results by similarity (lowest distance first)
- **FR-4.2.3**: The system shall handle queries with no results gracefully

### 5. Collection Management

#### 5.1 Reset Functionality
- **FR-5.1.1**: The system shall provide an API endpoint to clear all documents from the collection
- **FR-5.1.2**: The system shall log collection reset operations
- **FR-5.1.3**: The system shall handle cases where the collection doesn't exist

### 6. User Interface

#### 6.1 Upload Interface
- **FR-6.1.1**: The UI shall provide a file input for PDF selection
- **FR-6.1.2**: The UI shall provide input fields for chunk size and stride configuration
- **FR-6.1.3**: The UI shall display default values for chunk parameters
- **FR-6.1.4**: The UI shall show upload progress and status
- **FR-6.1.5**: The UI shall display success/error messages after upload
- **FR-6.1.6**: The UI shall reset the form after successful upload

#### 6.2 Search Interface
- **FR-6.2.1**: The UI shall provide a search input field
- **FR-6.2.2**: The UI shall support search submission via Enter key or button click
- **FR-6.2.3**: The UI shall display search results in a structured format
- **FR-6.2.4**: The UI shall show loading indicators during search
- **FR-6.2.5**: The UI shall display "no results" messages when appropriate

#### 6.3 Information Display
- **FR-6.3.1**: The UI shall display a "How It Works" section explaining the 3-step process
- **FR-6.3.2**: The UI shall show VectorGo branding and logo
- **FR-6.3.3**: The UI shall use a modern, gradient-based design aesthetic

---

## Non-Functional Requirements

### 1. Performance

- **NFR-1.1**: The system shall process PDF uploads within a reasonable time based on file size
- **NFR-1.2**: The system shall generate embeddings asynchronously to avoid blocking
- **NFR-1.3**: Search queries shall return results within 2 seconds under normal load
- **NFR-1.4**: The system shall handle concurrent uploads from multiple users

### 2. Reliability

- **NFR-2.1**: The system shall continue processing remaining chunks if individual chunk processing fails
- **NFR-2.2**: The system shall log all errors for debugging purposes
- **NFR-2.3**: The system shall validate all user inputs before processing
- **NFR-2.4**: The system shall handle network failures to Ollama and ChromaDB gracefully

### 3. Scalability

- **NFR-3.1**: The system shall support multiple document collections
- **NFR-3.2**: ChromaDB storage shall persist across container restarts
- **NFR-3.3**: The system shall support horizontal scaling of the backend service

### 4. Usability

- **NFR-4.1**: The UI shall be responsive and work on desktop and mobile devices
- **NFR-4.2**: Error messages shall be clear and actionable
- **NFR-4.3**: The UI shall provide visual feedback for all user actions
- **NFR-4.4**: The system shall use consistent terminology throughout the interface

### 5. Security

- **NFR-5.1**: The system shall validate file types before processing
- **NFR-5.2**: The system shall enforce file size limits to prevent abuse
- **NFR-5.3**: The system shall sanitize user inputs to prevent injection attacks
- **NFR-5.4**: The system shall use CORS headers to control API access

### 6. Maintainability

- **NFR-6.1**: The system shall use environment variables for all configuration
- **NFR-6.2**: The code shall follow Go and Svelte best practices
- **NFR-6.3**: The system shall provide comprehensive logging for debugging
- **NFR-6.4**: The system shall be containerized for easy deployment

---

## Technical Requirements

### 1. Technology Stack

- **Backend**: Go (Golang) 1.25.5+
- **Frontend**: Svelte 5+ with TypeScript
- **Styling**: Tailwind CSS 4+
- **Embeddings**: Ollama with embeddinggemma:300m model
- **Vector Database**: ChromaDB
- **Containerization**: Docker and Docker Compose

### 2. API Endpoints

#### 2.1 Health Check
- **Endpoint**: `GET /`
- **Response**: JSON with service status and version

#### 2.2 PDF Upload
- **Endpoint**: `POST /api/upload`
- **Content-Type**: `multipart/form-data`
- **Parameters**:
  - `file` (required): PDF file
  - `chunkSize` (optional): Integer, default 100
  - `chunkStride` (optional): Integer, default 80
- **Response**: JSON with processing status and metadata

#### 2.3 Search
- **Endpoint**: `GET /api/search?q=<query>`
- **Parameters**:
  - `q` (required): Search query string
- **Response**: JSON with matching documents, metadata, and distances

#### 2.4 Reset Collection
- **Endpoint**: `POST /api/reset`
- **Response**: JSON with reset status

### 3. Environment Configuration

Required environment variables:
- `OLLAMA_URL`: Ollama service URL (default: http://localhost:11434)
- `CHROMA_URL`: ChromaDB service URL (default: http://localhost:8000)
- `EMBEDDING_MODEL`: Embedding model name (default: embeddinggemma:300m)
- `COLLECTION_NAME`: ChromaDB collection name (default: documents)
- `PORT`: Server port (default: 8080)

### 4. Data Models

#### 4.1 Chunk Metadata
```json
{
  "source": "pdf",
  "filename": "document.pdf",
  "chunk_num": 1
}
```

#### 4.2 Search Response
```json
{
  "ids": [["uuid1", "uuid2"]],
  "documents": [["text1", "text2"]],
  "metadatas": [[{metadata1}, {metadata2}]],
  "distances": [[0.123, 0.456]]
}
```

---

## Deployment Requirements

### 1. Development Environment

- Docker Compose with:
  - Ollama service
  - ChromaDB service
  - Node.js builder container (for frontend development)

### 2. Production Environment

- Docker container with:
  - Compiled Go backend binary
  - Built frontend static files
- External services:
  - Ollama (with embeddinggemma:300m model)
  - ChromaDB (with persistent storage)

### 3. Build Process

- Frontend: Vite build to static files
- Backend: CGO_ENABLED=0 GOOS=linux Go build
- Docker: Multi-stage build or pre-built binaries

---

## Future Enhancements

### Potential Features (Not in Current Scope)

1. **Multi-format Support**: Support for DOCX, TXT, and other document formats
2. **Batch Upload**: Upload and process multiple PDFs simultaneously
3. **Advanced Search**: Filters, date ranges, and metadata-based search
4. **User Accounts**: Multi-user support with document ownership
5. **Export Results**: Export search results to CSV or JSON
6. **Analytics**: Dashboard showing document statistics and search patterns
7. **Custom Models**: Support for different embedding models per collection
8. **API Authentication**: Token-based API access control
9. **Document Preview**: Display PDF pages alongside search results
10. **Highlighting**: Highlight matching text within document snippets

---

## Constraints and Assumptions

### Constraints
- PDF files must be text-based (not scanned images)
- Maximum file size: 32MB
- Ollama and ChromaDB must be accessible via network
- Embedding model must be pre-pulled in Ollama

### Assumptions
- Users have basic understanding of chunking concepts
- Network connectivity is reliable
- Ollama service has sufficient resources for embedding generation
- ChromaDB has sufficient storage for vector data
- Users will primarily search in English (model-dependent)

---

## Success Criteria

1. Users can successfully upload PDF files
2. Text is accurately extracted and chunked
3. Embeddings are generated and stored in ChromaDB
4. Search returns relevant results based on semantic similarity
5. UI is intuitive and responsive
6. System handles errors gracefully
7. Documentation is complete and accurate
8. Deployment process is straightforward and repeatable

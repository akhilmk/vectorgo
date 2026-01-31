# VectorGo

VectorGo is a modern PDF processing and semantic search application that leverages vector embeddings to enable intelligent document search and retrieval.

## 🚀 Overview

VectorGo allows you to upload PDF documents, automatically chunk and embed the text using Ollama, store the vectors in ChromaDB, and perform semantic search across your document collection. The application features a clean, responsive interface and a robust backend designed for reliability and ease of deployment.

### Key Features

- **📄 PDF Processing**: Upload and process PDF documents automatically
- **✂️ Configurable Chunking**: Customize chunk size and stride for optimal search results
- **🤖 AI-Powered Embeddings**: Generate embeddings using Ollama's embedding models
- **🔍 Semantic Search**: Find relevant content using natural language queries
- **💾 Vector Storage**: Persistent storage with ChromaDB
- **⚡ Modern Tech Stack**: Built with Go, Svelte, Ollama, ChromaDB, and containerized with Docker

---

## 🛠️ Tech Stack

- **Backend**: Go (Golang)
- **Frontend**: Svelte + Tailwind CSS
- **Embeddings**: Ollama (embeddinggemma:300m)
- **Vector Database**: ChromaDB
- **Infrastructure**: Docker & Docker Compose

---

## 🚦 Quick Start

### Prerequisites
- Docker & Docker Compose
- Make

### Local Development Setup

1. **Clone the repository**:
   ```bash
   git clone <repository-url>
   cd vectorgo
   ```

2. **Start the development environment**:
   ```bash
   make dev-up
   ```
   This starts Ollama, ChromaDB, and a Node.js builder container.

3. **Pull the embedding model** (first time only):
   ```bash
   docker exec -it vectorgo-ollama ollama pull embeddinggemma:300m
   ```

4. **Build the application**:
   ```bash
   make docker
   ```

5. **Run the application**:
   ```bash
   make run
   ```

The app will be available at [http://localhost:8080](http://localhost:8080).

---

## 📖 How It Works

1. **Upload**: Select a PDF file and configure chunking parameters (chunk size and stride)
2. **Process**: The backend extracts text, splits it into chunks, generates embeddings via Ollama, and stores them in ChromaDB
3. **Search**: Enter natural language queries to find semantically similar content across all uploaded documents

### Chunking Parameters

- **Chunk Size**: Number of words per chunk (default: 100)
- **Chunk Stride**: Step size between chunks (default: 80)
  - Overlap = Chunk Size - Chunk Stride
  - Example: Size 100, Stride 80 = 20 words overlap

---

## 📖 Documentation

For more detailed information, please refer to our documentation:

- **[Developer Documentation](docs/developer-doc.md)**: Complete list of all available commands and development setup instructions

---

## 🔧 Configuration

Key environment variables (see `docker/.env.dev`):

- `OLLAMA_URL`: Ollama service URL (default: http://ollama:11434)
- `CHROMA_URL`: ChromaDB service URL (default: http://chromadb:8000)
- `EMBEDDING_MODEL`: Ollama embedding model (default: embeddinggemma:300m)
- `COLLECTION_NAME`: ChromaDB collection name (default: documents)

---

## 📄 License

This project is licensed under the MIT License - see the LICENSE file for details.

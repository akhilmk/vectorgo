# VectorGo Development Commands

This project uses a `Makefile` to simplify development tasks. Environment variables are automatically loaded from `docker/.env.dev`.

## Development Environment Commands

- `make dev-up`: Starts the development environment (Ollama + ChromaDB + Node.js builder) in the background.
- `make dev-down`: Stops the dev environment and removes the containers and volumes (data is reset).
- `make dev-logs`: Follows the output logs of the dev environment containers.

## Build Commands

- `make frontend-install`: **[First Time]** Installs the required NPM dependencies for the frontend.
- `make frontend-audit-fix`: If audit error occurs during `make frontend-install`, run this command to fix it.
- `make build-frontend`: **[Develop Time]** Build only frontend and copy to bin folder, copy UI changes to running container (using volume path mount).
- `make build-backend`: **[Develop Time]** Build only the Go backend binary.
- `make build-all`: Build both frontend and backend.
- `make docker`: **[App Docker Image]** Removes old containers, images, and build files, then `build-all` (frontend and backend), and finally **creates a new local Docker image**.

## App Run Commands

- `make run`: Starts the VectorGo application container locally. It automatically handles container removal if one is already running.
- `make logs`: Follows the application container logs.
- `make app-shell`: Opens an interactive shell inside the running application container for debugging.

## Stop and Clean Commands

- `make clean`: Deletes local build artifacts (`bin/` and `frontend/dist`).
- `make docker-stop`: Stops the application container.
- `make docker-clean`: Removes the application container and deletes the local Docker image.
- `make help`: Displays a summary of all available commands.

## Testing Commands

- `make go-test`: Runs all Go backend unit tests with verbose output.

---

## API Endpoints

VectorGo provides the following REST API endpoints:

### Health Check
- **GET** `/` - Returns service status and version information

### PDF Upload
- **POST** `/api/upload`
  - **Content-Type**: `multipart/form-data`
  - **Parameters**:
    - `file` (required): PDF file to upload
    - `chunkSize` (optional): Number of words per chunk (default: 100)
    - `chunkStride` (optional): Step size between chunks (default: 80)
  - **Response**: JSON with processing status and metadata

### Search
- **GET** `/api/search?q=<query>`
  - **Parameters**:
    - `q` (required): Search query string
  - **Response**: JSON with matching documents, metadata, and relevance scores

### Reset Collection
- **POST** `/api/reset` - Deletes all documents from the ChromaDB collection

---

## Development Workflow

### Initial Setup

1. **Start development environment**:
   ```bash
   make dev-up
   ```

2. **Pull the Ollama embedding model** (first time only):
   ```bash
   docker exec -it vectorgo-ollama ollama pull embeddinggemma:300m
   ```

3. **Install frontend dependencies** (first time only):
   ```bash
   make frontend-install
   ```

4. **Build the Docker image**:
   ```bash
   make docker
   ```

5. **Run the application**:
   ```bash
   make run
   ```

6. **Access the application**:
   - Open your browser to [http://localhost:8080](http://localhost:8080)

### During Development

**Frontend Changes**:
```bash
make build-frontend
```
This rebuilds the frontend and copies it to the running container.

**Backend Changes**:
```bash
make docker
make run
```
This rebuilds the entire application and restarts the container.

**View Logs**:
```bash
make logs              # Application logs
make dev-logs          # Dev environment logs
```

### Cleanup

**Stop application**:
```bash
make docker-stop
```

**Clean build artifacts**:
```bash
make clean
```

**Stop dev environment**:
```bash
make dev-down
```

---

## Troubleshooting

### Ollama Model Not Found
If you get embedding errors, ensure the model is pulled:
```bash
docker exec -it vectorgo-ollama ollama pull embeddinggemma:300m
```

### ChromaDB Connection Issues
Verify ChromaDB is running:
```bash
docker ps | grep vectorgo-chromadb
```

### Network Issues
Ensure the `vectorgo-dev` network exists:
```bash
docker network ls | grep vectorgo-dev
```

If not, restart the dev environment:
```bash
make dev-down
make dev-up
```

---

## Configuration

Environment variables can be modified in `docker/.env.dev`:

- `OLLAMA_URL`: Ollama service URL
- `CHROMA_URL`: ChromaDB service URL
- `EMBEDDING_MODEL`: Ollama embedding model name
- `COLLECTION_NAME`: ChromaDB collection name
- `PORT`: Application server port

---

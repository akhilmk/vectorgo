// REST API client for VectorGo

const API_BASE_URL = "/api";

export interface ProcessingResult {
    status: string;
    filename: string;
    chunkSize: number;
    chunkStride: number;
}

export interface SearchResult {
    ids: string[][];
    documents: string[][];
    metadatas: any[][];
    distances: number[][];
}

export interface StatsResult {
    total_chunks: number;
    total_files: number;
    files: string[];
    file_chunk_counts: { [key: string]: number };
}

class ApiError extends Error {
    constructor(public status: number, message: string) {
        super(message);
        this.name = "ApiError";
    }
}

async function handleResponse<T>(response: Response): Promise<T> {
    if (!response.ok) {
        const text = await response.text();
        throw new ApiError(response.status, text || response.statusText);
    }
    return response.json();
}


const TOKEN_KEY = "vectorgo_token";

function getAuthHeader(): HeadersInit {
    const token = localStorage.getItem(TOKEN_KEY);
    return token ? { "Authorization": `Bearer ${token}` } : {};
}

export const api = {
    isLoggedIn(): boolean {
        return !!localStorage.getItem(TOKEN_KEY);
    },

    async login(username: string, password: string): Promise<void> {
        const response = await fetch(`${API_BASE_URL}/login`, {
            method: "POST",
            headers: { "Content-Type": "application/json" },
            body: JSON.stringify({ username, password }),
        });

        const data = await handleResponse<{ token: string }>(response);
        localStorage.setItem(TOKEN_KEY, data.token);
    },

    logout() {
        localStorage.removeItem(TOKEN_KEY);
    },

    async uploadPDF(formData: FormData, onProgress?: (msg: string) => void, signal?: AbortSignal): Promise<ProcessingResult> {
        const file = formData.get("file") as File;
        const fileName = file?.name || "unknown";
        const fileSize = file?.size || 0;

        console.log(`[UPLOAD START] File: ${fileName} | Size: ${fileSize} bytes (${(fileSize / (1024 * 1024)).toFixed(2)} MB) | Timestamp: ${new Date().toISOString()}`);

        const response = await fetch(`${API_BASE_URL}/upload`, {
            method: "POST",
            headers: { ...getAuthHeader() },
            body: formData,
            signal: signal,
        });

        if (!response.body) {
            console.error(`[UPLOAD ERROR] File: ${fileName} | Error: Response body is empty`);
            throw new Error("Response body is empty");
        }

        const reader = response.body.getReader();
        const decoder = new TextDecoder();
        let buffer = "";
        let finalResult: ProcessingResult | null = null;

        try {
            while (true) {
                const { done, value } = await reader.read();
                if (done) break;

                buffer += decoder.decode(value, { stream: true });
                const lines = buffer.split("\n");
                buffer = lines.pop() || ""; // Keep incomplete line

                for (const line of lines) {
                    if (!line.trim()) continue;
                    try {
                        const data = JSON.parse(line);
                        if (data.error) {
                            console.error(`[UPLOAD ERROR] File: ${fileName} | Error: ${data.error} | Timestamp: ${new Date().toISOString()}`);
                            throw new Error(data.error);
                        }
                        if (data.status === "completed") {
                            console.log(`[UPLOAD COMPLETE] File: ${fileName} | Timestamp: ${new Date().toISOString()}`);
                            finalResult = data as ProcessingResult;
                        } else if (data.status && onProgress) {
                            console.log(`[UPLOAD PROGRESS] File: ${fileName} | Status: ${data.status} | Timestamp: ${new Date().toISOString()}`);
                            onProgress(data.status);
                        }
                    } catch (e) {
                        if (e instanceof Error && e.message !== "Unexpected end of JSON input") {
                            console.warn(`[UPLOAD WARNING] File: ${fileName} | Failed to parse JSON line: ${line} | Error: ${e.message}`);
                        }
                        if (line.includes('"error":')) throw e;
                    }
                }
            }
        } catch (error) {
            if (error instanceof Error && error.name === 'AbortError') {
                console.log(`[UPLOAD CANCELLED] File: ${fileName} | Timestamp: ${new Date().toISOString()}`);
                throw new Error('Upload cancelled');
            }
            throw error;
        }

        if (finalResult) {
            return finalResult;
        }

        if (response.ok) {
            console.error(`[UPLOAD ERROR] File: ${fileName} | Error: Upload process ended without completion status`);
            throw new Error("Upload process ended without completion status");
        }

        return handleResponse<ProcessingResult>(response);
    },

    async searchVectors(query: string): Promise<SearchResult> {
        const response = await fetch(`${API_BASE_URL}/search?q=${encodeURIComponent(query)}`, {
            headers: getAuthHeader()
        });
        return handleResponse<SearchResult>(response);
    },

    async resetCollection(): Promise<{ status: string }> {
        const response = await fetch(`${API_BASE_URL}/reset`, {
            method: "POST",
            headers: getAuthHeader(),
        });
        return handleResponse<{ status: string }>(response);
    },

    async getStats(): Promise<StatsResult> {
        const response = await fetch(`${API_BASE_URL}/stats`, {
            headers: getAuthHeader()
        });
        return handleResponse<StatsResult>(response);
    },

    async deleteFile(filename: string): Promise<{ status: string; filename: string }> {
        const response = await fetch(`${API_BASE_URL}/files/${encodeURIComponent(filename)}`, {
            method: "DELETE",
            headers: getAuthHeader()
        });
        return handleResponse<{ status: string; filename: string }>(response);
    }
};

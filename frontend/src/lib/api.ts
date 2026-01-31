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

export const api = {
    async uploadPDF(formData: FormData): Promise<ProcessingResult> {
        const response = await fetch(`${API_BASE_URL}/upload`, {
            method: "POST",
            body: formData,
        });
        return handleResponse<ProcessingResult>(response);
    },

    async searchVectors(query: string): Promise<SearchResult> {
        const response = await fetch(`${API_BASE_URL}/search?q=${encodeURIComponent(query)}`);
        return handleResponse<SearchResult>(response);
    },

    async resetCollection(): Promise<{ status: string }> {
        const response = await fetch(`${API_BASE_URL}/reset`, {
            method: "POST",
        });
        return handleResponse<{ status: string }>(response);
    }
};

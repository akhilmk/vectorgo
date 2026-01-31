import { writable } from 'svelte/store';

export interface UploadState {
    uploading: boolean;
    fileName: string;
    progress: string;
    abortController: AbortController | null;
}

const initialState: UploadState = {
    uploading: false,
    fileName: '',
    progress: '',
    abortController: null
};

function createUploadStore() {
    const { subscribe, set, update } = writable<UploadState>(initialState);

    return {
        subscribe,
        startUpload: (fileName: string, controller: AbortController) => {
            update(state => ({
                ...state,
                uploading: true,
                fileName,
                progress: 'Starting upload...',
                abortController: controller
            }));
        },
        updateProgress: (message: string) => {
            update(state => ({
                ...state,
                progress: message
            }));
        },
        completeUpload: () => {
            set(initialState);
        },
        cancelUpload: () => {
            update(state => {
                if (state.abortController) {
                    state.abortController.abort();
                }
                return initialState;
            });
        },
        reset: () => {
            set(initialState);
        }
    };
}

export const uploadStore = createUploadStore();

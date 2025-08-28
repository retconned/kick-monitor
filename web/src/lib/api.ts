class ApiError extends Error {
    status: number;
    constructor(message: string, status: number) {
        super(message);
        this.name = "ApiError";
        this.status = status;
    }
}

const BASE_URL = "http://localhost:8080";

export async function apiFetch<T>(
    endpoint: string,
    options: RequestInit = {},
): Promise<T> {
    // Get the token from localStorage
    const token = localStorage.getItem("token");

    const headers: HeadersInit = {
        "Content-Type": "application/json",
        ...(options.headers as Record<string, string> | undefined),
    };

    // If a token exists, add it to the Authorization header
    if (token) {
        headers["Authorization"] = `Bearer ${token}`;
    }

    const response = await fetch(`${BASE_URL}${endpoint}`, {
        ...options,
        headers,
        // REMOVE credentials: 'include', as it's for cookies
    });

    // Global 401 handler for expired/invalid tokens
    if (response.status === 401) {
        localStorage.removeItem("token");
        window.location.href = "/auth";
        throw new ApiError(
            "Your session has expired. Please log in again.",
            401,
        );
    }

    if (!response.ok) {
        const errorData = await response.json().catch(() => ({
            message: response.statusText,
        }));
        throw new ApiError(
            errorData.message || "An unknown API error occurred.",
            response.status,
        );
    }

    const contentType = response.headers.get("content-type");
    if (contentType && contentType.includes("application/json")) {
        return response.json() as Promise<T>;
    }
    return Promise.resolve(null as T);
}

export const HTTP_STATUS = {
	UNAUTHORIZED: 401,
	FORBIDDEN: 403,
	SERVER_ERROR: 500,
}

export const ERROR_MESSAGES = {
	AUTH_REQUIRED: "Authentication required. Please log in.",
	NO_PERMISSION: "You do not have permission to access this resource",
	SERVER_ERROR: "Server error occurred. Please try again later.",
	NO_RESPONSE:
		"No response received from server. Please check your connection.",
}

// Minimum source version that supports column selection.
export const MIN_COLUMN_SELECTION_SOURCE_VERSION = "v0.4.0"

export const DISPLAYED_JOBS_COUNT = 5

export const OLAKE_LATEST_VERSION_URL = "https://olake.io/docs/release/overview"

export const LOCALSTORAGE_TOKEN_KEY = "token"

export const LOCALSTORAGE_USERNAME_KEY = "username"

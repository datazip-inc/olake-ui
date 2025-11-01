import axios, {
	AxiosError,
	InternalAxiosRequestConfig,
	AxiosResponse,
} from "axios"
import { API_CONFIG } from "./config"
import {
	ERROR_MESSAGES,
	HTTP_STATUS,
	LOCALSTORAGE_TOKEN_KEY,
} from "../utils/constants"
import { notificationService } from "./services/notificationService"
/**
 * Extend Axios types to support our custom notification flag
 */
declare module "axios" {
	export interface AxiosRequestConfig {
		showNotification?: boolean // Controls whether the interceptor shows a toast (default: false)
	}
}

/**
 * Creates and configures an axios instance with default settings
 */
const api = axios.create({
	baseURL: API_CONFIG.BASE_URL,
	headers: {
		"Content-Type": "application/json",
		Accept: "application/json",
	},
	timeout: 10000,
	withCredentials: true,
})

/**
 * Request interceptor to add authentication token to requests
 */
api.interceptors.request.use(
	(config: InternalAxiosRequestConfig) => {
		const token = localStorage.getItem(LOCALSTORAGE_TOKEN_KEY)
		if (token && config.headers) {
			config.headers.Authorization = `Bearer ${token}`
		}
		return config
	},
	(error: AxiosError) => {
		return Promise.reject(error)
	},
)

/**
 * Response interceptor to handle common error cases
 */
api.interceptors.response.use(
	(response: AxiosResponse) => {
		const config = response.config
		const payload = response.data

		// Show toast only if explicitly enabled for this request
		if (config.showNotification === true) {
			notificationService.success(payload.message)
		}

		// Return only the actual data to the caller (unwrap the envelope)
		response.data = payload.data

		return response
	},
	(error: AxiosError) => {
		const payload = error.response?.data as any

		// Skip showing errors for canceled requests
		if (axios.isCancel(error) || error.code === "ERR_CANCELED") {
			return Promise.reject(error)
		}

		// Always show error toasts
		if (payload.message) {
			notificationService.error(payload.message)
		}

		// Handle specific HTTP status codes
		if (error.response) {
			const { status } = error.response

			switch (status) {
				case HTTP_STATUS.UNAUTHORIZED:
					localStorage.removeItem(LOCALSTORAGE_TOKEN_KEY)
					window.location.href = "/login"
					break
				case HTTP_STATUS.FORBIDDEN:
					console.error(ERROR_MESSAGES.NO_PERMISSION)
					break
				case HTTP_STATUS.SERVER_ERROR:
					console.error(ERROR_MESSAGES.SERVER_ERROR)
					break
			}
		} else if (error.request) {
			console.error(ERROR_MESSAGES.NO_RESPONSE)
		} else {
			console.error("Error setting up request:", error.message)
		}

		return Promise.reject(error)
	},
)

export default api

import axios, {
	AxiosError,
	InternalAxiosRequestConfig,
	AxiosResponse,
} from "axios"
import { API_CONFIG } from "./config"
import { ERROR_MESSAGES, HTTP_STATUS } from "../utils/constants"

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
		const token = localStorage.getItem("token")
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
		return response
	},
	(error: AxiosError) => {
		if (error.response) {
			const { status } = error.response

			switch (status) {
				case HTTP_STATUS.UNAUTHORIZED:
					localStorage.removeItem("token")
					console.error(ERROR_MESSAGES.AUTH_REQUIRED)
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

import axios from "axios"

// Create axios instance with default config
const api = axios.create({
	baseURL: import.meta.env.VITE_API_BASE_URL || "http://localhost:3000/api",
	headers: {
		"Content-Type": "application/json",
	},
	timeout: 10000, // 10 seconds
})

// Request interceptor
api.interceptors.request.use(
	config => {
		// Get token from localStorage
		const token = localStorage.getItem("token")

		// If token exists, add to headers
		if (token) {
			config.headers.Authorization = `Bearer ${token}`
		}

		return config
	},
	error => {
		return Promise.reject(error)
	},
)

// Response interceptor
api.interceptors.response.use(
	response => {
		return response
	},
	error => {
		// Handle common errors
		if (error.response) {
			// Server responded with a status code outside of 2xx range
			const { status } = error.response

			if (status === 401) {
				// Unauthorized - clear token and redirect to login
				localStorage.removeItem("token")
				// You can add redirect logic here if needed
			}

			if (status === 403) {
				// Forbidden - user doesn't have permission
				console.error("You do not have permission to access this resource")
			}

			if (status === 500) {
				// Server error
				console.error("Server error occurred. Please try again later.")
			}
		} else if (error.request) {
			// Request was made but no response received
			console.error(
				"No response received from server. Please check your connection.",
			)
		} else {
			// Something else happened while setting up the request
			console.error("Error setting up request:", error.message)
		}

		return Promise.reject(error)
	},
)

export default api

import axios from "axios"

const api = axios.create({
	baseURL: "http://4.240.65.100:8080",
	headers: {
		"Content-Type": "application/json",
		Accept: "application/json",
	},
	timeout: 10000,
	withCredentials: true,
})

api.interceptors.request.use(
	config => {
		const token = localStorage.getItem("token")
		if (token) {
			config.headers.Authorization = `Bearer ${token}`
		}
		return config
	},
	error => {
		return Promise.reject(error)
	},
)

api.interceptors.response.use(
	response => {
		return response
	},
	error => {
		if (error.response) {
			const { status } = error.response
			if (status === 401) {
				// Unauthorized - clear token and redirect to login
				localStorage.removeItem("token")
				console.error("Authentication required. Please log in.")
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

import api from "../axios"
import { APIResponse } from "../../types"

interface LoginResponse {
	username: string
}

export const authService = {
	login: async (username: string, password: string) => {
		try {
			const response = await api.post<APIResponse<LoginResponse>>(
				"/login",
				{
					username,
					password,
				},
				{
					headers: {
						"Content-Type": "application/json",
					},
				},
			)

			// Save username in localStorage to indicate logged in status
			if (response.data.success) {
				localStorage.setItem("username", response.data.data.username)
				localStorage.setItem("token", "authenticated")
				return response.data.data
			}

			throw new Error(response.data.message || "Login failed")
		} catch (error) {
			console.error("Login error:", error)
			throw error
		}
	},

	logout: () => {
		localStorage.removeItem("token")
		localStorage.removeItem("username")
	},

	isLoggedIn: () => {
		return !!localStorage.getItem("token")
	},

	getUsername: () => {
		return localStorage.getItem("username")
	},
}

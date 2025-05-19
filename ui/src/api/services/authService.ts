/**
 * AuthService handles authentication-related API calls and localStorage management.
 */
import api from "../axios"
import { APIResponse, LoginArgs } from "../../types"
import {
	LOCALSTORAGE_TOKEN_KEY,
	LOCALSTORAGE_USERNAME_KEY,
} from "../../utils/constants"

interface LoginResponse {
	username: string
	password: string
}

export const authService = {
	login: async ({ username, password }: LoginArgs) => {
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

			if (response.data.success) {
				localStorage.setItem(
					LOCALSTORAGE_USERNAME_KEY,
					response.data.data.username,
				)
				localStorage.setItem(LOCALSTORAGE_TOKEN_KEY, "authenticated")
				return response.data.data
			}

			throw new Error(response.data.message || "Login failed")
		} catch (error: any) {
			console.error("Login error:", error)
			throw new Error(error?.message || "Login failed")
		}
	},

	logout: () => {
		localStorage.removeItem(LOCALSTORAGE_TOKEN_KEY)
		localStorage.removeItem(LOCALSTORAGE_USERNAME_KEY)
	},

	isLoggedIn: () => {
		return !!localStorage.getItem(LOCALSTORAGE_TOKEN_KEY)
	},
}

/**
 * AuthService handles authentication-related API calls and localStorage management.
 */
import api from "../axios"
import { APIResponse, LoginArgs, LoginResponse } from "@types/index"
import {
	LOCALSTORAGE_TOKEN_KEY,
	LOCALSTORAGE_USERNAME_KEY,
} from "@utils/constants"

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

			throw new Error(response.data.message)
		} catch (error: any) {
			// Handle 400 status code specifically
			if (error.response?.status === 400) {
				throw new Error(error.response.data.message || "Invalid credentials")
			}
			// Handle other errors
			throw new Error(
				error.response?.data?.message || error.message || "Login failed",
			)
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

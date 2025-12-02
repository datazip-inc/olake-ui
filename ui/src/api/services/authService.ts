/**
 * AuthService handles authentication-related API calls and localStorage management.
 */
import api from "../axios"
import { LoginArgs, LoginResponse } from "../../types"
import {
	LOCALSTORAGE_TOKEN_KEY,
	LOCALSTORAGE_USERNAME_KEY,
} from "../../utils/constants"

export const authService = {
	login: async ({ username, password }: LoginArgs) => {
		try {
			const response = await api.post<LoginResponse>(
				"/login",
				{
					username,
					password,
				},
				{
					headers: {
						"Content-Type": "application/json",
					},
					disableErrorNotification: true,
				},
			)

			localStorage.setItem(LOCALSTORAGE_USERNAME_KEY, response.data.username)
			localStorage.setItem(LOCALSTORAGE_TOKEN_KEY, "authenticated")
			return response.data
		} catch (error: any) {
			throw error
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

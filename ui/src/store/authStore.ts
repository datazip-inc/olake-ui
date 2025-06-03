import { StateCreator } from "zustand"
import { authService } from "../api/services/authService"

export interface AuthSlice {
	isAuthenticated: boolean
	isAuthLoading: boolean
	initAuth: () => Promise<void>
	login: (username: string, password: string) => Promise<void>
	logout: () => void
}

export const createAuthSlice: StateCreator<
	AuthSlice & { jobsError: string | null }, // combines the slice with the state
	[],// middleware type
	[],//actions type
	AuthSlice // slice type
> = (set) => ({
	isAuthenticated: false,
	isAuthLoading: false,
	jobsError: null,
	initAuth: async () => {
		set({ isAuthLoading: true })
		try {
			if (!authService.isLoggedIn()) {
				set({ isAuthenticated: false, isAuthLoading: false })
				return
			}
			set({ isAuthenticated: true, isAuthLoading: false })
		} catch (error) {
			set({
				isAuthLoading: false,
				isAuthenticated: false,
				jobsError:
					error instanceof Error ? error.message : "Failed to initialize auth",
			})
		}
	},
	login: async (username: string, password: string) => {
		set({ isAuthLoading: true })
		try {
			await authService.login({ username, password })
			set({ isAuthenticated: true, isAuthLoading: false })
		} catch (error) {
			set({ isAuthLoading: false })
			throw error
		}
	},

	logout: () => {
		authService.logout()
		set({ isAuthenticated: false })
	},
})

import { StateCreator } from "zustand"
import { persist, createJSONStorage } from "zustand/middleware"

import { authService } from "../api/services/authService"

export interface AuthSlice {
	isAuthenticated: boolean
	isAuthLoading: boolean
	jobsError: string | null
	initAuth: () => Promise<void>
	login: (username: string, password: string) => Promise<void>
	logout: () => void
}

type AuthState = AuthSlice

export const createAuthSlice: StateCreator<
	AuthState,
	[],
	[["zustand/persist", Pick<AuthState, "isAuthenticated">]],
	AuthState
> = persist(
	set => ({
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
				// Identify user for analytics when returning with an existing session
			} catch (error) {
				set({
					isAuthLoading: false,
					isAuthenticated: false,
					jobsError:
						error instanceof Error
							? error.message
							: "Failed to initialize auth",
				})
			}
		},
		login: async (username: string, password: string) => {
			try {
				await authService.login({ username, password })
				set({ isAuthenticated: true, isAuthLoading: false })
				// Identify user for analytics after new login
				// Run in background without blocking to provide faster login experience
			} catch (error) {
				set({ isAuthLoading: false })
				throw error
			}
		},
		logout: () => {
			authService.logout()
			set({ isAuthenticated: false })
		},
	}),
	{
		name: "auth-storage",
		storage: createJSONStorage(() => localStorage),
		partialize: state => ({ isAuthenticated: state.isAuthenticated }),
	},
)

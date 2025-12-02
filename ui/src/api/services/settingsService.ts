import { api } from ".."
import {
	SystemSettings,
	UpdateSystemSettingsRequest,
} from "../../types/settingsTypes"
import { API_CONFIG } from "../config"

export const settingsService = {
	getSystemSettings: async (): Promise<SystemSettings> => {
		try {
			const response = await api.get<SystemSettings>(
				API_CONFIG.ENDPOINTS.SETTINGS(API_CONFIG.PROJECT_ID),
			)

			return response.data
		} catch (error) {
			console.error("Error fetching system settings from API:", error)
			throw error
		}
	},

	updateSystemSettings: async (systemSettings: UpdateSystemSettingsRequest) => {
		try {
			await api.put<SystemSettings>(
				API_CONFIG.ENDPOINTS.SETTINGS(API_CONFIG.PROJECT_ID),
				systemSettings,
				{ showNotification: true },
			)
		} catch (error) {
			console.error("Error updating system settings from API:", error)
			throw error
		}
	},
}

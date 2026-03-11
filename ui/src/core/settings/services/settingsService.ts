import { api } from "@/core/api"
import { SystemSettings, UpdateSystemSettingsRequest } from "../types"
import { API_CONFIG } from "@/config/apiConfig"

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
				{ ...systemSettings, project_id: API_CONFIG.PROJECT_ID },
				{ showNotification: true },
			)
		} catch (error) {
			console.error("Error updating system settings from API:", error)
			throw error
		}
	},
}

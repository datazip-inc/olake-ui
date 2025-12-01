import { StateCreator } from "zustand"
import { SystemSettings } from "../types/settingsTypes"
import { settingsService } from "../api/services/settingsService"

export interface SettingsSlice {
	systemSettings: SystemSettings | null
	isLoadingSystemSettings: boolean
	systemSettingsError: string | null
	isUpdatingSystemSettings: boolean

	fetchSystemSettings: () => Promise<void>
	updateWebhookAlertUrl: (webhookAlertUrl: string) => Promise<void>
}

export const createSettingsSlice: StateCreator<SettingsSlice> = (set, get) => ({
	systemSettings: null,
	isLoadingSystemSettings: false,
	systemSettingsError: null,
	isUpdatingSystemSettings: false,
	fetchSystemSettings: async () => {
		set({ isLoadingSystemSettings: true, systemSettingsError: null })
		try {
			const systemSettings = await settingsService.getSystemSettings()
			set({
				systemSettings,
				isLoadingSystemSettings: false,
			})
		} catch (error) {
			set({
				systemSettingsError:
					error instanceof Error
						? error.message
						: "Failed to fetch system settings",
				isLoadingSystemSettings: false,
			})
		}
	},
	updateWebhookAlertUrl: async (webhookAlertUrl: string) => {
		const systemSettings = get().systemSettings
		if (!systemSettings) {
			return
		}

		set({ isUpdatingSystemSettings: true, systemSettingsError: null })

		try {
			await settingsService.updateSystemSettings({
				...systemSettings,
				webhook_alert_url: webhookAlertUrl,
			})
			// fetch updated system settings
			await get().fetchSystemSettings()
		} catch (error) {
			set({
				systemSettingsError:
					error instanceof Error
						? error.message
						: "Failed to update system settings",
			})
			throw error
		} finally {
			set({ isUpdatingSystemSettings: false })
		}
	},
})

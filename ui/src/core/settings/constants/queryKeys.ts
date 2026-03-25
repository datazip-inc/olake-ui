import { API_CONFIG } from "@/config/apiConfig"

export const settingsKeys = {
	all: () => ["projects", API_CONFIG.PROJECT_ID, "settings"] as const,
	systemSettings: () => [...settingsKeys.all(), "system"] as const,
}

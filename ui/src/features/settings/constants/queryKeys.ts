import { API_CONFIG } from "@/api/config"

export const settingsKeys = {
	all: () => ["projects", API_CONFIG.PROJECT_ID, "settings"] as const,
	systemSettings: () => [...settingsKeys.all(), "system"] as const,
}

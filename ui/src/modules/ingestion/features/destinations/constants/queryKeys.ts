import { API_CONFIG } from "@/config/apiConfig"

export const destinationKeys = {
	all: () => ["projects", API_CONFIG.PROJECT_ID, "destinations"] as const,
	lists: () => [...destinationKeys.all(), "list"] as const,
	details: () => [...destinationKeys.all(), "details"] as const,
	detail: (id: string) => [...destinationKeys.details(), id] as const,
	versions: (type: string) =>
		[...destinationKeys.all(), "versions", type] as const,
	// Separate root — not nested under destinationKeys.all() so destination mutations never invalidate it
	spec: (
		type: string,
		version: string,
		sourceType: string = "",
		sourceVersion: string = "",
	) =>
		["spec", "destinations", type, version, sourceType, sourceVersion] as const,
}

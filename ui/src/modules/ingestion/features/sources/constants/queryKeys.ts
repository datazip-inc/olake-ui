import { API_CONFIG } from "@/config/apiConfig"

export const sourceKeys = {
	all: () => ["projects", API_CONFIG.PROJECT_ID, "sources"] as const,
	lists: () => [...sourceKeys.all(), "list"] as const,
	details: () => [...sourceKeys.all(), "details"] as const,
	detail: (id: string) => [...sourceKeys.details(), id] as const,
	versions: (type: string) => [...sourceKeys.all(), "versions", type] as const,
	// Separate root — not nested under sourceKeys.all() so source mutations never invalidate it
	spec: (type: string, version: string) =>
		["spec", "sources", type, version] as const,
}

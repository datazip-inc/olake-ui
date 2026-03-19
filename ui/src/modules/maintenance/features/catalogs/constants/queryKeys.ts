import { API_CONFIG } from "@/config/apiConfig"

export const catalogKeys = {
	all: () => ["projects", API_CONFIG.PROJECT_ID, "catalogs"] as const,

	list: () => [...catalogKeys.all(), "catalogs"] as const,

	details: () => [...catalogKeys.all(), "details"] as const,
	detail: (id: string) => [...catalogKeys.details(), id] as const,

	versions: (type: string) => [...catalogKeys.all(), "versions", type] as const,

	// Separate root — not nested under catalogKeys.all() so catalog mutations never invalidate it
	spec: (type: string, version: string) =>
		["spec", "catalogs", type, version] as const,
}

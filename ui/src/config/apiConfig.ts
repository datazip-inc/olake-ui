export const API_CONFIG = {
	// BASE_URL determines where API requests go.
	// This uses the same protocol, host, and port that the frontend was loaded from.
	// Example: if the app is served from http://localhost:8000, requests go to http://localhost:8000/api/...
	BASE_URL:
		import.meta.env.VITE_IS_DEV === "true"
			? `${window.location.protocol}//${window.location.host.split(":")[0]}:8000`
			: `${window.location.protocol}//${window.location.host}`,
	PROJECT_ID: "123",
	ENDPOINTS: {
		PROJECT: (projectId: string) => `/api/v1/project/${projectId}`,
		DESTINATIONS: (projectId: string) =>
			`/api/v1/project/${projectId}/destinations`,
		SOURCES: (projectId: string) => `/api/v1/project/${projectId}/sources`,
		JOBS: (projectId: string) => `/api/v1/project/${projectId}/jobs`,
		SETTINGS: (projectId: string) => `/api/v1/project/${projectId}/settings`,
		// Optimization / Fusion APIs
		FUSION_CATALOGS: `/api/v1/fusion/catalogs`,
		FUSION_CATALOG: (catalogName?: string) =>
			catalogName
				? `/api/v1/fusion/catalog/${encodeURIComponent(catalogName)}`
				: `/api/v1/fusion/catalog`,
		FUSION_TABLE: (catalog: string, database: string, tableName?: string) =>
			tableName
				? `/api/v1/fusion/tables/${encodeURIComponent(catalog)}/${encodeURIComponent(database)}/${encodeURIComponent(tableName)}`
				: `/api/v1/fusion/tables/${encodeURIComponent(catalog)}/${encodeURIComponent(database)}`,
		FUSION_PROCESS: (processId: string) =>
			`/api/v1/fusion/logs/process/${encodeURIComponent(processId)}`,
		PLATFORM: `/api/v1/platform`,
	},
}

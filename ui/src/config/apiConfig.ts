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
		ETL: {
			PROJECT: (projectId: string) => `/api/v1/project/${projectId}`,
			DESTINATIONS: (projectId: string) =>
				`/api/v1/project/${projectId}/destinations`,
			SOURCES: (projectId: string) => `/api/v1/project/${projectId}/sources`,
			JOBS: (projectId: string) => `/api/v1/project/${projectId}/jobs`,
			SETTINGS: (projectId: string) => `/api/v1/project/${projectId}/settings`,
		},
		OPT: {
			CATALOGS: (catalogName?: string) =>
				catalogName
					? `/api/opt/v1/catalogs/${encodeURIComponent(catalogName)}`
					: `/api/opt/v1/catalogs`,
			CATALOG: (catalogName?: string) =>
				catalogName
					? `/api/opt/v1/catalog/${encodeURIComponent(catalogName)}`
					: `/api/opt/v1/catalog`,
			TABLE_CONFIG: (catalog: string, database: string, tableName: string) =>
				`/api/opt/v1/${encodeURIComponent(catalog)}/${encodeURIComponent(database)}/${encodeURIComponent(tableName)}`,
			TABLE: (catalog: string, database: string, tableName: string) =>
				`/api/opt/v1/tables/catalogs/${encodeURIComponent(catalog)}/dbs/${encodeURIComponent(database)}/tables/${encodeURIComponent(tableName)}`,
			TABLES: (catalog: string, database: string) =>
				`/api/opt/v1/${encodeURIComponent(catalog)}/${encodeURIComponent(database)}/tables`,
			PROCESS: (processId: string) =>
				`/api/opt/v1/logs/process/${encodeURIComponent(processId)}`,
		},
		PLATFORM: `/api/v1/platform`,
	},
}

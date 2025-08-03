export const API_CONFIG = {
	BASE_URL: `${window.location.protocol}//${window.location.host}`,
	PROJECT_ID: "123",
	ENDPOINTS: {
		PROJECT: (projectId: string) => `/api/v1/project/${projectId}`,
		DESTINATIONS: (projectId: string) =>
			`/api/v1/project/${projectId}/destinations`,
		SOURCES: (projectId: string) => `/api/v1/project/${projectId}/sources`,
		JOBS: (projectId: string) => `/api/v1/project/${projectId}/jobs`,
	},
}

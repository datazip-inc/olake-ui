export const API_CONFIG = {
	// BASE_URL determines where API requests go.
	// This uses the same protocol, host, and port that the frontend was loaded from.
	// Example: if the app is served from http://localhost:8000, requests go to http://localhost:8000/api/...
	BASE_URL: import.meta.env.VITE_IS_DEV === "true"
		? `${window.location.protocol}//${window.location.host.split(":")[0]}:8000`
		: `${window.location.protocol}//${window.location.host}`,
	PROJECT_ID: "123",
	ENDPOINTS: {
		PROJECT: (projectId: string) => `/api/v1/project/${projectId}`,
		DESTINATIONS: (projectId: string) =>
			`/api/v1/project/${projectId}/destinations`,
		SOURCES: (projectId: string) => `/api/v1/project/${projectId}/sources`,
		JOBS: (projectId: string) => `/api/v1/project/${projectId}/jobs`,
	},
};

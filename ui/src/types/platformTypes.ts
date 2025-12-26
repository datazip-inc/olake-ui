// Release Type Enum
export enum ReleaseType {
	FEATURES = "features",
	OLAKE_UI_WORKER = "olake_ui_worker",
	OLAKE_HELM = "olake_helm",
	OLAKE = "olake",
}

// API Response Types (matching backend response)
export interface ReleaseMetadataResponse {
	version: string
	description: string
	tags: string[]
	date: string
	link: string
}

export interface ReleaseTypeData {
	current_version?: string
	releases: ReleaseMetadataResponse[]
}

export interface ReleasesResponse {
	[ReleaseType.OLAKE_UI_WORKER]?: ReleaseTypeData
	[ReleaseType.OLAKE_HELM]?: ReleaseTypeData
	[ReleaseType.OLAKE]?: ReleaseTypeData
	[ReleaseType.FEATURES]?: ReleaseTypeData
}

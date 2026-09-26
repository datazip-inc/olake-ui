export interface APIResponse<T> {
	success: boolean
	message: string
	data: T
}

export interface SpecResponse {
	spec?: {
		jsonschema: object
		uischema: string
		// Only returned for available_query_engines requests.
		query_engines?: QueryEngineCatalog
	}
	message?: string
}

export interface QueryEngineCatalog {
	engines: QueryEngineSpec[]
	writable_update_types: string[]
}

export interface QueryEngineSpec {
	engine: string
	label: string
	supports: string[]
}

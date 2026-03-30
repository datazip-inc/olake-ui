export interface APIResponse<T> {
	success: boolean
	message: string
	data: T
}

export interface SpecResponse {
	spec?: {
		jsonschema: object
		uischema: string
	}
	message?: string
}

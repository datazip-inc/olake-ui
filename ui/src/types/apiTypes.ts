export interface APIResponse<T> {
	success: boolean
	message: string
	data: T
}
export interface SourceTestResponse {
	message: string
	status: "FAILED" | "SUCCEEDED"
}
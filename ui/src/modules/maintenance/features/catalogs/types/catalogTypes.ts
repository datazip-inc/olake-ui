// Backend API Types
export interface CatalogPayload {
	name: string
	version: string
	config: CatalogFormData
}

export interface FusionCatalog {
	catalogName: string
	catalogType: string
	catalogProperties: Record<string, string>
}

export type GetCatalogsResponse = {
	code: number
	message: string
	result: FusionCatalog[]
}

export interface CatalogTestRequest {
	version: string
	config: string
}

export interface GetCatalogDatabasesResponse {
	result: string[]
}

// Frontend Domain Types
export type CatalogFormData = Record<string, unknown>

export interface CatalogModalProps {
	open: boolean
	onClose: () => void
	onSuccess?: () => void
	catalogName?: string
}

export interface Catalog {
	id: string
	name: string
	type: string
	createdOn: string
	olakeCreated: boolean
}

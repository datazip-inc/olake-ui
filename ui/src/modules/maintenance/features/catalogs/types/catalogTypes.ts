import type { TestConnectionStatus } from "@/common/types"

// Backend API Types
export interface CatalogPayload {
	name: string
	version: string
	config: CatalogFormData
}

export interface FusionCatalog {
	name: string
	type: string
	databases: string[]
}

export type GetCatalogsResponse = {
	catalogs: FusionCatalog[]
}

export interface CatalogTestRequest {
	version: string
	config: string
}

export interface CatalogLogEntry {
	level: string
	time: string
	message: string
}

export interface CatalogTestResponse {
	connection_result: {
		message: string
		status: TestConnectionStatus
	}
	logs: CatalogLogEntry[]
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
	databases: string[]
}

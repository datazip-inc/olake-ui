import { SourceConnector, DestinationConnector, CatalogType } from "../enums"

export interface SourceFormConfig {
	name: string
	connector: SourceConnector
	version?: string
	fields: Record<string, any>
}

export interface DestinationFormConfig {
	name: string
	connector: DestinationConnector
	version?: string
	catalogType?: CatalogType
	fields: Record<string, any>
}

export interface JobFormConfig {
	sourceName: string
	destinationName: string
	sourceConnector: SourceConnector
	destinationConnector: DestinationConnector
	streamName: string
	jobName: string
	frequency?: string
}

export interface SourceFormConfig {
	name: string
	connector: string
	version?: string
	fields: Record<string, any>
}

export interface DestinationFormConfig {
	name: string
	connector: string
	version?: string
	catalogType?: "glue" | "jdbc" | "hive" | "rest"
	fields: Record<string, any>
}

export interface JobFormConfig {
	sourceName: string
	destinationName: string
	streamName: string
	jobName: string
	frequency?: string
}

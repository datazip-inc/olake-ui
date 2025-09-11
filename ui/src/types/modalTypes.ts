import { StreamsDataStructure } from "./streamTypes"

export interface DeleteModalProps {
	fromSource: boolean
}

export interface DestinationDatabaseModalProps {
	destinationType: string
	destinationDatabase: string | null
	allStreams: StreamsDataStructure | null
	onSave: (format: string, databaseName: string) => void
	originalDatabase: string
}

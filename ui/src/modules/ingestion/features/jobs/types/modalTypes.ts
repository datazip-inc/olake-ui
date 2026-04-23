import { StreamsDataStructure } from "@/modules/ingestion/common/types"

import { IngestionMode } from "../enums"

export interface DestinationDatabaseModalProps {
	destinationType: string
	destinationDatabase: string | null
	allStreams: StreamsDataStructure | null
	onSave: (format: string, databaseName: string) => void
	originalDatabase: string
	initialStreams: StreamsDataStructure | null
}

export interface ResetStreamsModalProps {
	onConfirm: () => void
}

export interface StreamDifferenceModalProps {
	streamDifference: StreamsDataStructure
	onConfirm: () => void
}

export interface StreamEditDisabledModalProps {
	from: "jobSettings" | "jobEdit"
}

export interface IngestionModeChangeModalProps {
	onConfirm: (ingestionMode: IngestionMode) => void
	ingestionMode: IngestionMode
}

export interface BulkConfigureStreamsModalProps {
	open: boolean
	onClose: () => void
	streamsData: StreamsDataStructure | null
	sourceType?: string
	destinationType?: string
}

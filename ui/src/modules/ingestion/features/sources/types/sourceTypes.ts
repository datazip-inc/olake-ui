import { Entity } from "@/modules/ingestion/common/types"

export interface SourceTableProps {
	sources: Entity[]
	loading: boolean
	onEdit: (id: string) => void
	onDelete: (source: Entity) => void
}

export interface Source {
	id: string | number
	name: string
	type: string
	version: string
	config?: any
}

export interface SourceJob {
	destination_type: string
	last_run_time: string
	last_run_state: string
	id: number
	name: string
	activate: boolean
	destination_name: string
}

export interface SourceData {
	id?: number
	name: string
	type: string
	config: Record<string, any>
	version?: string
}

export interface DiscoverSourceStreamsParams {
	name: string
	type: string
	version: string
	config: string
	job_name: string
	job_id?: number
	max_discover_threads?: number | null
}

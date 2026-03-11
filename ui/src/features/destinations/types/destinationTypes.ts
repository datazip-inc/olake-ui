import { Entity, EntityJob } from "@/common/types"

export interface DestinationConfig {
	[key: string]: any
	catalog?: string
	catalog_type?: string
	writer?: {
		catalog?: string
		catalog_type?: string
	}
}

export interface Destination {
	id: string | number
	name: string
	type: string
	version: string
	config: string | DestinationConfig
}

export interface ExtendedDestination extends Destination {
	config: DestinationConfig
}
export interface DestinationTableProps {
	destinations: Entity[]
	loading: boolean
	onEdit: (id: string) => void
	onDelete: (destination: Entity) => void
}

export type SelectOption = { value: string; label: React.ReactNode | string }

export interface DestinationJob extends EntityJob {
	source_type: string
	last_run_time: string
	last_run_state: string
	source_name: string
}

export interface DestinationData {
	id?: number
	name: string
	type: string
	config: Record<string, any>
	version?: string
}

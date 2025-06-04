import { CatalogType } from "./commonTypes"
import { Entity, EntityJob } from "./entityTypes"

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

export interface CreateDestinationProps {
	fromJobFlow?: boolean
	hitBack?: boolean
	onComplete?: () => void
	stepNumber?: number
	stepTitle?: string
	initialConfig?: {
		name: string
		type: string
		config?: DestinationConfig
	}
	initialFormData?: DestinationConfig
	initialName?: string
	initialConnector?: string
	initialCatalog?: CatalogType | null
	onDestinationNameChange?: (name: string) => void
	onConnectorChange?: (connector: string) => void
	onFormDataChange?: (formData: DestinationConfig) => void
	onVersionChange?: (version: string) => void
	onCatalogTypeChange?: (catalog: CatalogType | null) => void
}

export interface DestinationTableProps {
	destinations: Entity[]
	loading: boolean
	onEdit: (id: string) => void
	onDelete: (destination: Entity) => void
}

export type SelectOption = { value: string; label: React.ReactNode | string }

export interface DestinationJob extends Omit<EntityJob, "last_runtime"> {
	source_type: string
	last_run_time: string
	last_run_state: string
	source_name: string
}

export interface DestinationData {
	id?: string
	name: string
	type: string
	config: Record<string, any>
	version?: string
}

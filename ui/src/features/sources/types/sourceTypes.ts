import { Entity, EntityBase } from "@/common/types"

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

export interface CreateSourceProps {
	onComplete?: () => void
	initialConfig?: EntityBase
	initialFormData?: any
	initialName?: string
	initialConnector?: string
	initialVersion?: string
	initialExistingSourceId?: number | null
	onSourceNameChange?: (name: string) => void
	onConnectorChange?: (connector: string) => void
	onFormDataChange?: (formData: any) => void
	onVersionChange?: (version: string) => void
	onExistingSourceIdChange?: (id: number | null) => void
	docsMinimized?: boolean
	onDocsMinimizedChange?: React.Dispatch<React.SetStateAction<boolean>>
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

export interface SourceEditProps {
	initialData?: any
	onNameChange?: (name: string) => void
	onConnectorChange?: (type: string) => void
	onVersionChange?: (version: string) => void
	onFormDataChange?: (config: Record<string, any>) => void
	docsMinimized?: boolean
	onDocsMinimizedChange?: React.Dispatch<React.SetStateAction<boolean>>
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

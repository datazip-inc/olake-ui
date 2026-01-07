import { Entity, EntityBase } from "./entityTypes"

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
	fromJobFlow?: boolean
	onComplete?: () => void
	stepNumber?: number
	stepTitle?: string
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
	fromJobFlow?: boolean
	stepNumber?: number
	stepTitle?: string
	initialData?: any
	onNameChange?: (name: string) => void
	onConnectorChange?: (type: string) => void
	onVersionChange?: (version: string) => void
	onFormDataChange?: (config: Record<string, any>) => void
	docsMinimized?: boolean
	onDocsMinimizedChange?: React.Dispatch<React.SetStateAction<boolean>>
}

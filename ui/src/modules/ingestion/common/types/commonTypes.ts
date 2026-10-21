import {
	EntityJob,
	EntityType,
} from "@/modules/ingestion/common/types/entityTypes"

export type UnknownObject = {
	[key: string]: unknown | UnknownObject
}

export type SetupType = "new" | "existing"

export interface ConnectorOption {
	value: string
	label: React.ReactNode
}

export interface EndpointTitleProps {
	title?: string
}
export interface SetupTypeSelectorProps {
	value: SetupType
	onChange: (value: SetupType) => void
	newLabel?: string
}

export interface TabsFilterProps {
	tabs: { key: string; label: string }[]
	activeTab: string
	onChange: (key: string) => void
}

export interface OpenDocsButtonProps {
	docUrl: string
	className?: string
}
export interface DocumentationPanelProps {
	docUrl: string
	// Heading anchor on the docs page to scroll to
	anchor?: string
	isMinimized?: boolean
	onToggle?: () => void
	showResizer?: boolean
	initialWidth?: number
}
export interface StepIndicatorProps {
	step: string
	index: number
	currentStep: string
	onStepClick?: (step: string) => void
	isEditMode?: boolean
	disabled?: boolean
}

export interface StepProgressProps {
	currentStep: string
	onStepClick?: (step: string) => void
	isEditMode?: boolean
	disabled?: boolean
}

export interface LayoutProps {
	children: React.ReactNode
}

export interface JobConnectionProps {
	sourceType: string
	destinationType: string
	jobName: string
	remainingJobs?: number
	jobs: EntityJob[]
}

export interface AssociatedJobsProps {
	jobs?: EntityJob[] | null
	entityType: EntityType
	// connector of the source/destination the row belongs to
	connectorType: string
}

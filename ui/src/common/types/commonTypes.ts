import type { IconProps } from "@phosphor-icons/react"
import { EntityJob } from "./entityTypes"

export type UnknownObject = {
	[key: string]: unknown | UnknownObject
}
export interface NavItem {
	path: string
	label: string
	icon: React.ComponentType<IconProps>
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

export interface DocumentationPanelProps {
	docUrl: string
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

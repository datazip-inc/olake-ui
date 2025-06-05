import type { IconProps } from "@phosphor-icons/react"
import type { CATALOG_TYPES } from "../utils/constants"

export type UnknownObject = {
	[key: string]: unknown | UnknownObject
}
export interface NavItem {
	path: string
	label: string
	icon: React.ComponentType<IconProps>
}
export type CatalogType = (typeof CATALOG_TYPES)[keyof typeof CATALOG_TYPES]

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
	existingLabel?: string
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

export type FilterButtonProps = {
	filter: string
	selectedFilters: string[]
	setSelectedFilters: (filters: string[]) => void
}

export interface StepIndicatorProps {
	step: string
	index: number
	currentStep: string
}

export interface StepProgressProps {
	currentStep: string
}

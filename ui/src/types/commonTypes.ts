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

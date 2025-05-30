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

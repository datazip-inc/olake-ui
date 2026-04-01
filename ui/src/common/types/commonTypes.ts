import type { IconProps } from "@phosphor-icons/react"

export interface NavItem {
	path: string
	label: string
	icon: React.ComponentType<IconProps>
}

export type TestConnectionStatus = "FAILED" | "SUCCEEDED"

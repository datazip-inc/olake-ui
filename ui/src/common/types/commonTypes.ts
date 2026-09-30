import type { IconProps } from "@phosphor-icons/react"

import { LogEntry } from "./errorTypes"

export interface NavItem {
	path: string
	label: string
	icon: React.ComponentType<IconProps>
}

export type TestConnectionStatus = "FAILED" | "SUCCEEDED"

// One CDC setup check reported by the source driver during test connection
export interface PrerequisiteCheck {
	name: string
	required: boolean
	passed: boolean
	current_value: string
	recommended_value: string
	description: string
}

export interface TestConnectionResponse {
	connection_result: {
		message: string
		status: TestConnectionStatus
		prerequisites?: PrerequisiteCheck[]
	}
	logs: LogEntry[]
}

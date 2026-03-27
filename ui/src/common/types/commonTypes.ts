import type { IconProps } from "@phosphor-icons/react"

import { LogEntry } from "./errorTypes"

export interface NavItem {
	path: string
	label: string
	icon: React.ComponentType<IconProps>
}

export type TestConnectionStatus = "FAILED" | "SUCCEEDED"

export interface TestResponse {
	connection_result: {
		message: string
		status: TestConnectionStatus
	}
	logs: LogEntry[]
}

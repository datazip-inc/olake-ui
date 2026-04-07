import { ReactNode } from "react"

export interface Props {
	children: ReactNode
	fallback?: ReactNode
}

export interface State {
	hasError: boolean
	error: Error | null
}

export interface TestConnectionError {
	message: string
	logs: LogEntry[]
}

export interface LogEntry {
	level: string
	time: string
	message: string
}

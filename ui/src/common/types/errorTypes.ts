import { ReactNode } from "react"
import { LogEntry } from "@/modules/ingestion/common/types/entityTypes"

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

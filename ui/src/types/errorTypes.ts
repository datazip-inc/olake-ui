import { ReactNode } from "react"

export interface Props {
	children: ReactNode
	fallback?: ReactNode
}

export interface State {
	hasError: boolean
	error: Error | null
}

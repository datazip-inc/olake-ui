import type React from "react"

export type ActionRouteProps = {
	feature?: string
	children: React.ReactNode
}

export const ActionRoute: React.FC<ActionRouteProps> = ({ children }) => (
	<>{children}</>
)

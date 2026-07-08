import type React from "react"

export type ActionPanelProps = {
	feature?: string
	children: React.ReactNode
}

export const ActionPanel: React.FC<ActionPanelProps> = ({ children }) => (
	<>{children}</>
)

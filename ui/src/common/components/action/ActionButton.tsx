import type React from "react"

export type ActionButtonProps = {
	feature?: string
	children: React.ReactNode
}

export const ActionButton: React.FC<ActionButtonProps> = ({ children }) => (
	<>{children}</>
)

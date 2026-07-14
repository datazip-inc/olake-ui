import type React from "react"

import { useActionAccess } from "@/common/components/action/useActionAccess"

export type ActionProps = {
	access?: string
	fallback?: React.ReactNode
	children: React.ReactNode
}

export const Action: React.FC<ActionProps> = ({
	access,
	fallback = null,
	children,
}) => {
	const { canAccess } = useActionAccess()

	if (access && !canAccess(access)) {
		return <>{fallback}</>
	}

	return <>{children}</>
}

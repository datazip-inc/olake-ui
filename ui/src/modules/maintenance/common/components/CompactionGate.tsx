import { Navigate, Outlet } from "react-router-dom"

import { useOptimizationStatus } from "@/core/platform/hooks/useOptimizationStatus"

const CompactionGate: React.FC = () => {
	const { data, isLoading } = useOptimizationStatus()

	// While loading, render nothing to avoid a flash redirect
	if (isLoading) return null

	if (data?.enabled) {
		return <Outlet />
	}

	return (
		<Navigate
			to="/jobs"
			replace
		/>
	)
}

export default CompactionGate

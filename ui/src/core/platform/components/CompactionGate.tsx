import { Navigate, Outlet } from "react-router-dom"

import { useCompactionStatus } from "../hooks/useCompactionStatus"

const CompactionGate: React.FC = () => {
	const { data, isLoading } = useCompactionStatus()

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

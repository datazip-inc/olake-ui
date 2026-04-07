import { useIsFetching } from "@tanstack/react-query"
import { Spin } from "antd"
import { useState } from "react"
import { useNavigate, useParams } from "react-router-dom"

import { RunLogSidebar, RunLogsPanel } from "../components"
import { DRIVER_SOURCE_KEY } from "../constants"
import { tableKeys } from "../constants/queryKeys"

const RunLogs: React.FC = () => {
	const navigate = useNavigate()
	const { catalog, database, tableName, runId } = useParams<{
		catalog: string
		database: string
		tableName: string
		runId: string
	}>()

	const decodedTableName = decodeURIComponent(tableName ?? "")
	const decodedCatalog = decodeURIComponent(catalog ?? "")
	const decodedDatabase = decodeURIComponent(database ?? "")
	const decodedRunId = decodeURIComponent(runId ?? "")
	const isLoading =
		useIsFetching({
			queryKey: tableKeys.processLogs(decodedRunId),
		}) > 0

	const [selectedSourceKey, setSelectedSourceKey] = useState(DRIVER_SOURCE_KEY)
	const runsPath = `/maintenance/tables/${encodeURIComponent(decodedCatalog)}/${encodeURIComponent(decodedDatabase)}/${encodeURIComponent(decodedTableName)}/runs`

	return (
		<div className="relative flex h-full min-h-0">
			{isLoading && (
				<div className="absolute inset-0 z-10 flex items-center justify-center bg-white">
					<Spin size="large" />
				</div>
			)}
			<RunLogSidebar
				tableName={decodedTableName}
				runId={decodedRunId}
				selectedSourceKey={selectedSourceKey}
				onSelectSource={setSelectedSourceKey}
				onBack={() => navigate(runsPath)}
			/>
			<RunLogsPanel
				runId={decodedRunId}
				selectedSourceKey={selectedSourceKey}
			/>
		</div>
	)
}

export default RunLogs

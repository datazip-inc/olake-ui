import { useState } from "react"
import { useNavigate, useParams } from "react-router-dom"

import { RunLogSidebar, RunLogsPanel } from "../components"
import { DRIVER_SOURCE_KEY } from "../constants"

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

	const [selectedSourceKey, setSelectedSourceKey] = useState(DRIVER_SOURCE_KEY)
	const runsPath = `/maintenance/tables/${encodeURIComponent(decodedCatalog)}/${encodeURIComponent(decodedDatabase)}/${encodeURIComponent(decodedTableName)}/runs`

	return (
		<div className="flex h-full min-h-0">
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

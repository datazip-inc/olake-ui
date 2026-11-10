import { useIsFetching } from "@tanstack/react-query"
import { Spin } from "antd"
import { useState } from "react"
import { useNavigate, useParams, useSearchParams } from "react-router"

import { RunLogSidebar, RunLogsPanel } from "../components"
import {
	DRIVER_SOURCE_KEY,
	RUN_LOGS_TYPE_QUERY_PARAM,
	RUN_TYPE,
	tableKeys,
} from "../constants"

const RunLogs: React.FC = () => {
	const navigate = useNavigate()
	const [searchParams] = useSearchParams()
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
	// Run history passes the run kind along; the logs API itself does not say which kind it is.
	// TODO: read the run kind from Fusion's /logs/process/{id} once it returns processType, instead of ?runType=.
	const isExpiryRun =
		searchParams.get(RUN_LOGS_TYPE_QUERY_PARAM) === RUN_TYPE.EXPIRY
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
				isExpiryRun={isExpiryRun}
				selectedSourceKey={selectedSourceKey}
				onSelectSource={setSelectedSourceKey}
				onBack={() => navigate(runsPath)}
			/>
			<RunLogsPanel
				runId={decodedRunId}
				isExpiryRun={isExpiryRun}
				selectedSourceKey={selectedSourceKey}
			/>
		</div>
	)
}

export default RunLogs

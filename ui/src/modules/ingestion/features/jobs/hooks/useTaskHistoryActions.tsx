import { EyeIcon } from "@phosphor-icons/react"
import { Button } from "antd"
import { useNavigate } from "react-router-dom"

import { JobTask } from "../types"

export function useTaskHistoryActions(jobId: string | undefined) {
	const navigate = useNavigate()

	const renderTaskAction = (record: JobTask) => (
		<Button
			type="default"
			icon={<EyeIcon size={16} />}
			onClick={() => {
				if (!jobId) return
				navigate(
					`/jobs/${jobId}/history/1/logs?file=${encodeURIComponent(record.file_path)}`,
				)
			}}
		>
			View logs
		</Button>
	)

	return {
		renderTaskAction,
		modal: null,
	}
}

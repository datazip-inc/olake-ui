import { Tag } from "@/common/components"
import { AssociatedJobsProps } from "@/modules/ingestion/common/types"

import JobConnection from "./JobConnection"

// "Associated jobs" cell of the sources and destinations tables. Only active
// jobs are listed; entities whose jobs are all paused get a red label.
const AssociatedJobs: React.FC<AssociatedJobsProps> = ({
	jobs,
	entityType,
	connectorType,
}) => {
	const allJobs = jobs ?? []
	if (allJobs.length === 0) {
		return <Tag className="text-xs">No associated jobs</Tag>
	}

	const activeJobs = allJobs.filter(job => job.activate)
	if (activeJobs.length === 0) {
		return (
			<Tag
				color="error"
				className="text-xs"
			>
				No active jobs associated
			</Tag>
		)
	}

	const [firstJob] = activeJobs
	return (
		<JobConnection
			sourceType={
				entityType === "source" ? connectorType : firstJob.source_type || ""
			}
			destinationType={
				entityType === "destination"
					? connectorType
					: firstJob.destination_type || ""
			}
			jobName={firstJob.name}
			remainingJobs={activeJobs.length - 1}
			jobs={activeJobs}
		/>
	)
}

export default AssociatedJobs

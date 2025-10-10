import { Tooltip } from "antd"
import { getConnectorImage } from "@utils/utils"
import { JobConnectionProps } from "@app-types/index"

const JobConnection: React.FC<JobConnectionProps> = ({
	sourceType,
	destinationType,
	jobName,
	remainingJobs = 0,
	jobs,
}) => {
	const remainingJobNames = jobs
		.slice(1)
		.map(job => job.name)
		.join(", ")

	return (
		<div className="flex-end flex w-fit flex-col items-end gap-3">
			<div className="mb-1 flex items-center">
				<div className="flex items-center gap-3">
					<div className="flex items-center">
						<img
							src={getConnectorImage(sourceType)}
							className="size-8"
							alt={`${sourceType} connector`}
						/>
						<div className="ml-2 text-text-muted">-------</div>
						<div className="w-36 truncate rounded-md border border-[#D9D9D9] bg-black bg-opacity-[2%] px-2 py-1 text-center text-black">
							{jobName.length > 15 ? (
								<Tooltip title={jobName}>{jobName}</Tooltip>
							) : (
								jobName
							)}
						</div>
						<div className="mr-2 text-text-muted">-------</div>
						<img
							src={getConnectorImage(destinationType)}
							className="size-8"
							alt={`${destinationType} connector`}
						/>
					</div>
					{remainingJobs > 0 && (
						<Tooltip title={remainingJobNames}>
							<div className="cursor-pointer items-end text-sm font-bold text-primary">
								+{remainingJobs} more jobs
							</div>
						</Tooltip>
					)}
				</div>
			</div>
		</div>
	)
}

export default JobConnection

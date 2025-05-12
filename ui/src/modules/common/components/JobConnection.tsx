import { Tooltip } from "antd"
import { getConnectorImage } from "../../../utils/utils"

interface JobConnectionProps {
	sourceType: string
	destinationType: string
	jobName: string
	remainingJobs?: number
}

const JobConnection: React.FC<JobConnectionProps> = ({
	sourceType,
	destinationType,
	jobName,
	remainingJobs = 0,
}) => {
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
						<div className="ml-2 text-[#A3A3A3]">-------</div>
						<div className="w-36 truncate rounded-[6px] border border-[#D9D9D9] bg-black bg-opacity-[2%] px-2 py-1 text-center text-black">
							{jobName.length > 15 ? (
								<Tooltip title={jobName}>{jobName}</Tooltip>
							) : (
								jobName
							)}
						</div>
						<div className="mr-2 text-[#A3A3A3]">-------</div>
						<img
							src={getConnectorImage(destinationType)}
							className="size-8"
							alt={`${destinationType} connector`}
						/>
					</div>
					{remainingJobs > 0 && (
						<div className="cursor-pointer items-end text-sm font-bold text-[#203FDD]">
							+{remainingJobs} more jobs
						</div>
					)}
				</div>
			</div>
		</div>
	)
}

export default JobConnection

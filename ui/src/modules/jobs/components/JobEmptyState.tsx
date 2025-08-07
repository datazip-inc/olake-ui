import { Button } from "antd"
import { GitCommit, PlayCircle } from "@phosphor-icons/react"

import { JobTutorialYTLink } from "../../../utils/constants"
import FirstJob from "../../../assets/FirstJob.svg"
import JobsTutorial from "../../../assets/JobsTutorial.svg"

const JobEmptyState = ({
	handleCreateJob,
}: {
	handleCreateJob: () => void
}) => {
	return (
		<div className="flex flex-col items-center justify-center py-16">
			<img
				src={FirstJob}
				alt="Empty state"
				className="mb-8 h-64 w-96"
			/>
			<div className="text-brand-blue mb-2">Welcome User !</div>
			<h2 className="mb-2 text-2xl font-bold">Ready to run your first Job</h2>
			<p className="text-text-primary mb-8">
				Get started and experience the speed of O<b>Lake</b> by running jobs
			</p>
			<Button
				type="primary"
				className="bg-brand-blue mb-12 text-sm"
				onClick={handleCreateJob}
			>
				<GitCommit />
				Create your first Job
			</Button>
			<div className="w-[412px] rounded-xl border-[1px] border-[#D9D9D9] bg-white p-6 shadow-sm">
				<div className="flex items-center gap-4">
					<a
						href={JobTutorialYTLink}
						target="_blank"
						rel="noopener noreferrer"
						className="cursor-pointer"
					>
						<img
							src={JobsTutorial}
							alt="Job Tutorial"
							className="rounded-lg transition-opacity hover:opacity-80"
						/>
					</a>
					<div className="flex-1">
						<div className="mb-1 flex items-center gap-1 text-xs">
							<PlayCircle color="#9f9f9f" />
							<span className="text-text-placeholder">OLake/ Tutorial</span>
						</div>
						<div className="text-xs">
							Checkout this tutorial, to know more about running jobs
						</div>
					</div>
				</div>
			</div>
		</div>
	)
}

export default JobEmptyState

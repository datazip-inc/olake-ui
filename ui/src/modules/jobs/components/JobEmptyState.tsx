import { GitCommit, PlayCircle } from "@phosphor-icons/react"
import FirstJob from "../../../assets/FirstJob.svg"
import JobsTutorial from "../../../assets/JobsTutorial.svg"
import { Button } from "antd"

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
			<div className="mb-2 text-[#193AE6]">Welcome User !</div>
			<h2 className="mb-2 text-2xl font-bold">Ready to run your first Job</h2>
			<p className="mb-8 text-[#0A0A0A]">
				Get started and experience the speed of O<b>Lake</b> by running jobs
			</p>
			<Button
				type="primary"
				className="mb-12 bg-[#193AE6] text-sm"
				onClick={handleCreateJob}
			>
				<GitCommit />
				Create your first Job
			</Button>
			<div className="w-[412px] rounded-xl border-[1px] border-[#D9D9D9] bg-white p-6 shadow-sm">
				<div className="flex items-center gap-4">
					<img
						src={JobsTutorial}
						alt="Job Tutorial"
						className="rounded-lg"
					/>
					<div className="flex-1">
						<div className="mb-1 flex items-center gap-1 text-xs">
							<PlayCircle color="#9f9f9f" />
							<span className="text-[#9F9F9F]">OLake/ Tutorial</span>
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

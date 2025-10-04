import { Button } from "antd"
import { PlayCircleIcon, PlusIcon } from "@phosphor-icons/react"

import { DestinationTutorialYTLink } from "../../../utils/constants"
import FirstDestination from "../../../assets/FirstDestination.svg"
import DestinationTutorial from "../../../assets/DestinationTutorial.svg"

const DestinationEmptyState = ({
	handleCreateDestination,
}: {
	handleCreateDestination: () => void
}) => {
	return (
		<div className="flex flex-col items-center justify-center py-16">
			<img
				src={FirstDestination}
				alt="Empty state"
				className="mb-8 h-64 w-96"
			/>
			<div className="mb-2 text-brand-blue">Welcome User !</div>
			<h2 className="mb-2 text-2xl font-bold">
				Ready to create your first destination
			</h2>
			<p className="mb-8 text-text-primary">
				Get started and experience the speed of OLake by running jobs
			</p>
			<Button
				type="primary"
				className="border-1 mb-12 border-[1px] border-[#D9D9D9] bg-white px-6 py-4 text-black"
				onClick={handleCreateDestination}
			>
				<PlusIcon />
				New Destination
			</Button>
			<div className="w-[412px] rounded-xl border-[1px] border-[#D9D9D9] bg-white p-4 shadow-sm">
				<div className="flex items-center gap-4">
					<a
						href={DestinationTutorialYTLink}
						target="_blank"
						rel="noopener noreferrer"
						className="cursor-pointer"
					>
						<img
							src={DestinationTutorial}
							alt="Job Tutorial"
							className="rounded-lg transition-opacity hover:opacity-80"
						/>
					</a>
					<div className="flex-1">
						<div className="mb-1 flex items-center gap-1 text-xs">
							<PlayCircleIcon color="#9f9f9f" />
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

export default DestinationEmptyState

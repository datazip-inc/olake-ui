import { PlayCircle, Plus, GitCommit } from "@phosphor-icons/react"
import { Button } from "antd"
import { EmptyStateType, EmptyStateConfig } from "../../../utils/constants"

interface EmptyStateProps {
	type: EmptyStateType
	onButtonClick: () => void
}

const EmptyState = ({ type, onButtonClick }: EmptyStateProps) => {
	const config = EmptyStateConfig[type]
	return (
		<div className="flex flex-col items-center justify-center py-16">
			<img
				src={config.image}
				alt={`${type} empty state`}
				className="mb-8 h-64 w-96"
			/>

			<div className={`mb-2 text-[#193AE6]`}>Welcome User !</div>

			<h2 className="mb-2 text-2xl font-bold">
				Ready to run your first {config.title}
			</h2>

			<p className={`mb-8 text-[#0A0A0A]`}>
				Get started and experience the speed of OLake by setting up
				{config.description}
			</p>

			<Button
				type="primary"
				className={`mb-12 px-6 py-4 ${
					type === EmptyStateType.JOB
						? "border-[#193AE6] bg-[#193AE6] text-white"
						: "border-[#D9D9D9] bg-white text-black"
				}`}
				onClick={onButtonClick}
			>
				{type === EmptyStateType.JOB ? <GitCommit /> : <Plus />}
				{config.buttonText}
			</Button>
			{config.tutorial && (
				<div className="w-[412px] rounded-xl border-[1px] border-[#D9D9D9] bg-white p-4 shadow-sm">
					<div className="flex items-center gap-4">
						<a
							href={config.tutorial.link}
							target="_blank"
							rel="noopener noreferrer"
							className="cursor-pointer"
						>
							<img
								src={config.tutorial.image}
								alt={`${type} tutorial`}
								className="rounded-lg transition-opacity hover:opacity-80"
							/>
						</a>
						<div className="flex-1">
							<div className="mb-1 flex items-center gap-1 text-xs">
								<PlayCircle color="#9f9f9f" />
								<span className="text-[#9F9F9F]">OLake/ Tutorial</span>
							</div>
							<div className="text-xs">
								Checkout this tutorial, to know more about {type.toLowerCase()}
							</div>
						</div>
					</div>
				</div>
			)}
		</div>
	)
}

export default EmptyState

import { PlayCircle, Plus, GitCommit } from "@phosphor-icons/react"
import { Button } from "antd"
import { EmptyStateType, EmptyStateConfig } from "../../../utils/constants"

interface EmptyStateProps {
	type: EmptyStateType
	onButtonClick: () => void
}

interface TutorialCardProps {
	tutorial: {
		link: string
		image: string
	}
	type: string
}

const TutorialCard = ({ tutorial, type }: TutorialCardProps) => (
	<div className="w-[412px] rounded-xl border border-neutral-disabled bg-white p-4 shadow-sm">
		<div className="flex items-center gap-4">
			<a
				href={tutorial.link}
				target="_blank"
				rel="noopener noreferrer"
				className="cursor-pointer"
			>
				<img
					src={tutorial.image}
					alt={`${type} tutorial`}
					className="rounded-lg transition-opacity hover:opacity-80"
				/>
			</a>
			<div className="flex-1">
				<div className="mb-1 flex items-center gap-1 text-xs">
					<PlayCircle className="text-text-placeholder" />
					<span className="text-text-placeholder">OLake / Tutorial</span>
				</div>
				<div className="text-xs">
					Checkout this tutorial to know more about {type.toLowerCase()}
				</div>
			</div>
		</div>
	</div>
)

const getButtonClass = (type: EmptyStateType) =>
	type === EmptyStateType.JOB
		? "border-brand-blue bg-brand-blue text-white"
		: "border-neutral-disabled bg-white text-black"

const getEmptyStateMessage = (type: EmptyStateType, config: any) => {
	switch (type) {
		case EmptyStateType.SOURCE:
			return "Ready to create your first source"
		case EmptyStateType.DESTINATION:
			return "Ready to create your first destination"
		case EmptyStateType.JOB:
			return "Ready to run your first Job"
		default:
			return `Ready to run your first ${config.title}`
	}
}

const EmptyState = ({ type, onButtonClick }: EmptyStateProps) => {
	const config = EmptyStateConfig[type]
	if (!config) return null

	return (
		<div className="flex flex-col items-center justify-center py-16">
			<img
				src={config.image}
				alt={`${type} empty state illustration`}
				className="mb-8 h-64 w-96"
			/>
			<div className="mb-2 font-medium text-brand-blue">Welcome User !</div>
			<h2 className="mb-2 text-2xl font-bold">
				{getEmptyStateMessage(type, config)}
			</h2>
			<p className="mb-8 text-text-primary">
				Get started and experience the speed of{" "}
				<span className="font-bold">OLake</span> by setting up
				{config.description}
			</p>
			<Button
				type="primary"
				className={`mb-12 px-6 py-4 ${getButtonClass(type)}`}
				onClick={onButtonClick}
			>
				{type === EmptyStateType.JOB ? <GitCommit /> : <Plus />}
				{config.buttonText}
			</Button>
			{config.tutorial && (
				<TutorialCard
					tutorial={config.tutorial}
					type={type}
				/>
			)}
		</div>
	)
}

export default EmptyState

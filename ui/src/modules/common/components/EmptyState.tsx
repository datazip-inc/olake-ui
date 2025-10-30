import { Button } from "antd"
import { PlayCircleIcon, PlusIcon, GitCommitIcon } from "@phosphor-icons/react"

import { EmptyStateProps } from "../../../types"

import FirstDestination from "../../../assets/FirstDestination.svg"
import DestinationTutorial from "../../../assets/DestinationTutorial.svg"
import FirstSource from "../../../assets/FirstSource.svg"
import SourcesTutorial from "../../../assets/SourcesTutorial.svg"
import FirstJob from "../../../assets/FirstJob.svg"
import JobsTutorial from "../../../assets/JobsTutorial.svg"

interface PageConfig {
	welcomeTextColor: string
	heading: string
	description: string
	descriptionColor: string
	buttonText: string
	buttonIcon: "Plus" | "GitCommit"
	buttonClassName: string
	image: string
	tutorialImage: string
	tutorialLink: string
}

const PAGES_CONFIG: Record<"job" | "source" | "destination", PageConfig> = {
	destination: {
		welcomeTextColor: "text-brand-blue",
		heading: "Ready to create your first destination",
		description:
			"Get started and experience the speed of OLake by running jobs",
		descriptionColor: "text-text-primary",
		buttonText: "New Destination",
		buttonIcon: "Plus",
		buttonClassName:
			"border-1 mb-12 border-[1px] border-[#D9D9D9] bg-white px-6 py-4 text-black",
		image: FirstDestination,
		tutorialImage: DestinationTutorial,
		tutorialLink: "https://youtu.be/Ub1pcLg0WsM?si=V2tEtXvx54wDoa8Y",
	},

	source: {
		welcomeTextColor: "text-blue-600",
		heading: "Ready to create your first source",
		description:
			"Get started and experience the speed of OLake by running jobs",
		descriptionColor: "text-gray-600",
		buttonText: "New Source",
		buttonIcon: "Plus",
		buttonClassName:
			"border-1 mb-12 border-[1px] border-[#D9D9D9] bg-white px-6 py-4 text-black",
		image: FirstSource,
		tutorialImage: SourcesTutorial,
		tutorialLink: "https://youtu.be/ndCHGlK5NCM?si=jvPy-aMrpEXCQA-8",
	},

	job: {
		welcomeTextColor: "text-brand-blue",
		heading: "Ready to run your first Job",
		description:
			"Get started and experience the speed of O<b>Lake</b> by running jobs",
		descriptionColor: "text-text-primary",
		buttonText: "Create your first Job",
		buttonIcon: "GitCommit",
		buttonClassName: "mb-12 bg-brand-blue text-sm",
		image: FirstJob,
		tutorialImage: JobsTutorial,
		tutorialLink: "https://youtu.be/_qRulFv-BVM?si=NPTw9V0hWQ3-9wOP",
	},
}

// Icon mapping
const ICON_MAP: Record<"Plus" | "GitCommit", React.ComponentType<any>> = {
	Plus: PlusIcon,
	GitCommit: GitCommitIcon,
}

const EmptyState: React.FC<EmptyStateProps> = ({ page, onButtonClick }) => {
	const config = PAGES_CONFIG[page]
	const IconComponent = ICON_MAP[config.buttonIcon]

	return (
		<div className="flex flex-col items-center justify-center py-16">
			<img
				src={config.image}
				alt="Empty state"
				className="mb-8 h-64 w-96"
			/>

			<div className={`mb-2 ${config.welcomeTextColor}`}>Welcome User!</div>

			<h2 className="mb-2 text-2xl font-bold">{config.heading}</h2>

			<div
				className={`mb-8 ${config.descriptionColor}`}
				dangerouslySetInnerHTML={{ __html: config.description }}
			/>

			<Button
				type="primary"
				className={config.buttonClassName}
				onClick={onButtonClick}
			>
				<IconComponent />
				{config.buttonText}
			</Button>

			<div className="w-[412px] rounded-xl border-[1px] border-[#D9D9D9] bg-white p-4 shadow-sm">
				<div className="flex items-center gap-4">
					<a
						href={config.tutorialLink}
						target="_blank"
						rel="noopener noreferrer"
						className="cursor-pointer"
					>
						<img
							src={config.tutorialImage}
							alt={`${page} Tutorial`}
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

export default EmptyState

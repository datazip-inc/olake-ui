import { Button } from "antd"
import { PlayCircle, Plus, GitCommit } from "@phosphor-icons/react"

import { EmptyStateProps } from "../../../types"

import FirstDestination from "../../../assets/FirstDestination.svg"
import DestinationTutorial from "../../../assets/DestinationTutorial.svg"
import FirstSource from "../../../assets/FirstSource.svg"
import SourcesTutorial from "../../../assets/SourcesTutorial.svg"
import FirstJob from "../../../assets/FirstJob.svg"
import JobsTutorial from "../../../assets/JobsTutorial.svg"

const ASSET_MAP: Record<string, string> = {
	"/src/assets/FirstDestination.svg": FirstDestination,
	"/src/assets/DestinationTutorial.svg": DestinationTutorial,
	"/src/assets/FirstSource.svg": FirstSource,
	"/src/assets/SourcesTutorial.svg": SourcesTutorial,
	"/src/assets/FirstJob.svg": FirstJob,
	"/src/assets/JobsTutorial.svg": JobsTutorial,
}

const ICON_MAP: Record<string, any> = {
	Plus: Plus,
	GitCommit: GitCommit,
}

const EmptyState: React.FC<EmptyStateProps> = ({ config, onButtonClick }) => {
	const IconComponent = ICON_MAP[config.button.icon]

	return (
		<div className="flex flex-col items-center justify-center py-16">
			<img
				src={ASSET_MAP[config.image]}
				alt="Empty state"
				className="mb-8 h-64 w-96"
			/>

			<div className={`mb-2 ${config.welcomeTextColor}`}>
				{config.welcomeText}
			</div>

			<h2 className="mb-2 text-2xl font-bold">{config.heading}</h2>

			<div
				className={`mb-8 ${config.descriptionColor}`}
				dangerouslySetInnerHTML={{ __html: config.description }}
			/>

			<Button
				type="primary"
				className={config.button.className}
				onClick={onButtonClick}
			>
				<IconComponent />
				{config.button.text}
			</Button>

			<div className="w-[412px] rounded-xl border-[1px] border-[#D9D9D9] bg-white p-4 shadow-sm">
				<div className="flex items-center gap-4">
					<a
						href={config.tutorial.link}
						target="_blank"
						rel="noopener noreferrer"
						className="cursor-pointer"
					>
						<img
							src={ASSET_MAP[config.tutorial.image]}
							alt={config.tutorial.altText}
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

export default EmptyState

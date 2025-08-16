// ui/src/components/common/EmptyState.tsx
import { PlayCircle } from "@phosphor-icons/react"
import { Button } from "antd"
import { ReactNode } from "react"

interface EmptyStateProps {
	// Main content
	image: string
	imageAlt?: string
	welcomeText: string
	welcomeTextColor?: string
	title: string
	description: string
	descriptionColor?: string

	// Action button
	buttonText: string
	buttonIcon: ReactNode
	onButtonClick: () => void
	buttonClassName?: string

	// Tutorial section (optional)
	tutorialImage?: string
	tutorialImageAlt?: string
	tutorialLink?: string
	tutorialTitle?: string
	tutorialDescription?: string
	showTutorial?: boolean
}

// This is a common reusable component for Empty state for Jobs Sources and Destination

const CommonEmptyState = ({
	image,
	imageAlt = "Empty state",
	welcomeText,
	welcomeTextColor = "text-blue-600",
	title,
	description,
	descriptionColor = "text-gray-600",
	buttonText,
	buttonIcon,
	onButtonClick,
	buttonClassName = "border-1 mb-12 border-[1px] border-[#D9D9D9] bg-white px-6 py-4 text-black",
	tutorialImage,
	tutorialImageAlt = "Tutorial",
	tutorialLink,
	tutorialTitle = "OLake/ Tutorial",
	tutorialDescription = "Checkout this tutorial, to know more about running jobs",
	showTutorial = true,
}: EmptyStateProps) => {
	return (
		<div className="flex flex-col items-center justify-center py-16">
			{/* Main illustration */}
			<img
				src={image}
				alt={imageAlt}
				className="mb-8 h-64 w-96"
			/>

			{/* Welcome text */}
			<div className={`mb-2 ${welcomeTextColor}`}>{welcomeText}</div>

			{/* Title */}
			<h2 className="mb-2 text-2xl font-bold">{title}</h2>

			{/* Description */}
			<p className={`mb-8 ${descriptionColor}`}>{description}</p>

			{/* Action button */}
			<Button
				type="primary"
				className={buttonClassName}
				onClick={onButtonClick}
			>
				{buttonIcon}
				{buttonText}
			</Button>

			{/* Tutorial section */}
			{showTutorial && tutorialImage && tutorialLink && (
				<div className="w-[412px] rounded-xl border-[1px] border-[#D9D9D9] bg-white p-4 shadow-sm">
					<div className="flex items-center gap-4">
						<a
							href={tutorialLink}
							target="_blank"
							rel="noopener noreferrer"
							className="cursor-pointer"
						>
							<img
								src={tutorialImage}
								alt={tutorialImageAlt}
								className="rounded-lg transition-opacity hover:opacity-80"
							/>
						</a>
						<div className="flex-1">
							<div className="mb-1 flex items-center gap-1 text-xs">
								<PlayCircle color="#9f9f9f" />
								<span className="text-[#9F9F9F]">{tutorialTitle}</span>
							</div>
							<div className="text-xs">{tutorialDescription}</div>
						</div>
					</div>
				</div>
			)}
		</div>
	)
}

export default CommonEmptyState

// Usage example for Sources:
/*
<EmptyState
	image={FirstSource}
	welcomeText="Welcome User !"
	welcomeTextColor="text-blue-600"
	title="Ready to create your first source"
	description="Get started and experience the speed of OLake by running jobs"
	buttonText="New Source"
	buttonIcon={<Plus />}
	onButtonClick={handleCreateSource}
	tutorialImage={SourcesTutorial}
	tutorialLink={SourceTutorialYTLink}
/>
*/

// Usage example for Jobs:
/*
<EmptyState
	image={FirstJob}
	welcomeText="Welcome User !"
	welcomeTextColor="text-[#193AE6]"
	title="Ready to run your first Job"
	description="Get started and experience the speed of OLake by running jobs"
	descriptionColor="text-[#0A0A0A]"
	buttonText="Create your first Job"
	buttonIcon={<GitCommit />}
	onButtonClick={handleCreateJob}
	buttonClassName="mb-12 bg-[#193AE6] text-sm"
	tutorialImage={JobsTutorial}
	tutorialLink={JobTutorialYTLink}
/>
*/

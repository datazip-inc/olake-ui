import { EmptyStateConfig } from "../types"

export const EMPTY_STATE_CONFIGS: Record<string, EmptyStateConfig> = {
	destination: {
		image: "/src/assets/FirstDestination.svg",
		welcomeText: "Welcome User !",
		welcomeTextColor: "text-brand-blue",
		heading: "Ready to create your first destination",
		description:
			"Get started and experience the speed of OLake by running jobs",
		descriptionColor: "text-text-primary",
		button: {
			text: "New Destination",
			icon: "Plus",
			className:
				"border-1 mb-12 border-[1px] border-[#D9D9D9] bg-white px-6 py-4 text-black",
		},
		tutorial: {
			link: "https://youtu.be/Ub1pcLg0WsM?si=V2tEtXvx54wDoa8Y",
			image: "/src/assets/DestinationTutorial.svg",
			altText: "Destination Tutorial",
		},
	},

	source: {
		image: "/src/assets/FirstSource.svg",
		welcomeText: "Welcome User !",
		welcomeTextColor: "text-blue-600",
		heading: "Ready to create your first source",
		description:
			"Get started and experience the speed of OLake by running jobs",
		descriptionColor: "text-gray-600",
		button: {
			text: "New Source",
			icon: "Plus",
			className:
				"border-1 mb-12 border-[1px] border-[#D9D9D9] bg-white px-6 py-4 text-black",
		},
		tutorial: {
			link: "https://youtu.be/ndCHGlK5NCM?si=jvPy-aMrpEXCQA-8",
			image: "/src/assets/SourcesTutorial.svg",
			altText: "Source Tutorial",
		},
	},

	job: {
		image: "/src/assets/FirstJob.svg",
		welcomeText: "Welcome User !",
		welcomeTextColor: "text-brand-blue",
		heading: "Ready to run your first Job",
		description:
			"Get started and experience the speed of O<b>Lake</b> by running jobs",
		descriptionColor: "text-text-primary",
		button: {
			text: "Create your first Job",
			icon: "GitCommit",
			className: "mb-12 bg-brand-blue text-sm",
		},
		tutorial: {
			link: "https://youtu.be/_qRulFv-BVM?si=NPTw9V0hWQ3-9wOP",
			image: "/src/assets/JobsTutorial.svg",
			altText: "Job Tutorial",
		},
	},
}

import { BellIcon, XIcon } from "@phosphor-icons/react"
import { Modal, Tooltip } from "antd"
import { useAppStore } from "../../../store"
import { useState } from "react"
import clsx from "clsx"
import ReactMarkdown from "react-markdown"

const CATEGORIES = ["Features", "OLake UI and Worker", "OLake Helm", "OLake"]

const TAG_STYLES: Record<string, { bg: string; text: string }> = {
	"new-release": { bg: "bg-blue-50", text: "text-blue-600" },
	performance: { bg: "bg-green-50", text: "text-green-600" },
	stable: { bg: "bg-green-50", text: "text-green-600" },
	"bug-fix": { bg: "bg-red-50", text: "text-red-600" },
	feature: { bg: "bg-purple-50", text: "text-purple-600" },
	security: { bg: "bg-orange-50", text: "text-orange-600" },
	breaking: { bg: "bg-yellow-50", text: "text-yellow-700" },
	default: { bg: "bg-gray-50", text: "text-gray-600" },
}

const getTagStyle = (tag: string) => {
	return TAG_STYLES[tag] ?? TAG_STYLES.default
}

// Type definition for update items from Docker
interface UpdateItem {
	title?: string
	version?: string
	description: string // Markdown formatted release notes
	releaseDate: string
	tags: string[] // Tags like "New Release", "Critical", "Security Fix", etc.
	releaseNotesLink?: string
}

// Mock data - will be replaced with API call
const MOCK_UPDATES: Record<string, UpdateItem[]> = {
	"OLake UI and Worker": [
		{
			version: "v0.2.5",
			description: `- perf: add pagination in logs api`,
			releaseDate: "Released on 12-Dec-23",
			tags: ["new-release", "performance", "feature", "security", "breaking"],
			releaseNotesLink:
				"https://github.com/datazip-inc/olake-ui/releases/tag/v0.2.5",
		},
		{
			version: "v0.2.4",
			description: `- fix: prevent job state from being overwritten during job updates
- fix: increase beego request body limits`,
			releaseDate: "Released on 08-Dec-23",
			tags: ["bug-fix", "security"],
			releaseNotesLink:
				"https://github.com/datazip-inc/olake-ui/releases/tag/v0.2.5",
		},
		{
			version: "v0.2.3",
			description: `- chore: guard clear-destination by source version and enforce unique names for jobs/sources/destinations
- feat: add system settings page and webhook alerts section
- fix: disable editing on saved webhook alert url`,
			releaseDate: "Released on 05-Dec-23",
			tags: ["bug-fix", "stable", "feature"],
			releaseNotesLink:
				"https://github.com/datazip-inc/olake-ui/releases/tag/v0.2.5",
		},
		{
			version: "v0.2.2",
			description: `- fix: add s3 backward compatibility in ui
- fix: cursor_field issue and other ui issues
- fix: telemetry fixes
- fix: remove redundant api calls from streams`,
			releaseDate: "Released on 03-Dec-23",
			tags: ["new-release", "performance", "feature", "security", "breaking"],
			releaseNotesLink:
				"https://github.com/datazip-inc/olake-ui/releases/tag/v0.2.5",
		},
		{
			version: "v0.2.1",
			description: `- test: remove iceberg db from integration test config
- fix: temporal payload blob size issue`,
			releaseDate: "Released on 01-Dec-23",
			tags: [],
			releaseNotesLink:
				"https://github.com/datazip-inc/olake-ui/releases/tag/v0.2.5",
		},
	],
	"OLake Helm": [
		{
			version: "v1.2.0",
			description: `- feat: add support for custom ingress configurations
- feat: enable horizontal pod autoscaling
- fix: update resource limits for better performance`,
			releaseDate: "Released on 10-Dec-23",
			tags: ["new-release", "feature"],
			releaseNotesLink:
				"https://github.com/datazip-inc/olake-helm/releases/tag/v1.2.0",
		},
		{
			version: "v1.1.5",
			description: `- fix: correct service account permissions
				- fix: update default resource requests`,
			releaseDate: "Released on 05-Dec-23",
			tags: ["bug-fix", "stable"],
			releaseNotesLink:
				"https://github.com/datazip-inc/olake-helm/releases/tag/v1.2.0",
		},
		{
			version: "v1.1.4",
			description: `- feat: add support for external secrets
					- fix: improve chart documentation`,
			releaseDate: "Released on 28-Nov-23",
			tags: ["feature", "bug-fix"],
			releaseNotesLink:
				"https://github.com/datazip-inc/olake-helm/releases/tag/v1.2.0",
		},
	],
	OLake: [
		{
			version: "v2.0.0",
			description: `- feat: major architecture improvements
- feat: new connector framework
- breaking: updated API endpoints`,
			releaseDate: "Released on 15-Dec-23",
			tags: ["new-release", "breaking", "feature"],
			releaseNotesLink:
				"https://github.com/datazip-inc/olake/releases/tag/v2.0.0",
		},
		{
			version: "v1.9.2",
			description: `- fix: memory leak in data processing pipeline
				- fix: improve error handling in connectors
				- perf: optimize query performance`,
			releaseDate: "Released on 08-Dec-23",
			tags: ["bug-fix", "performance", "stable"],
			releaseNotesLink:
				"https://github.com/datazip-inc/olake/releases/tag/v2.0.0",
		},
		{
			version: "v1.9.1",
			description: `- security: update dependencies to patch vulnerabilities
					- fix: correct data type handling in transformations`,
			releaseDate: "Released on 01-Dec-23",
			tags: ["security", "bug-fix"],
			releaseNotesLink:
				"https://github.com/datazip-inc/olake/releases/tag/v2.0.0",
		},
	],
	// Features: [
	// 	{
	// 		title: "Advanced Data Transformation Pipeline",
	// 		description:
	// 			"Introducing a powerful new transformation engine that allows you to apply complex data transformations in real-time. Build custom transformation logic with our visual editor or write custom SQL/Python code.",
	// 		releaseDate: "Released on 15-Dec-23",
	// 		tags: ["new-release", "feature"],
	// 	},
	// 	{
	// 		title: "Real-time Monitoring Dashboard",
	// 		description:
	// 			"Monitor your data pipelines in real-time with our new comprehensive dashboard. Get instant insights into data flow, performance metrics, and error rates with customizable widgets and alerts.",
	// 		releaseDate: "Released on 10-Dec-23",
	// 		tags: ["feature", "performance"],
	// 	},
	// 	{
	// 		title: "Enhanced Security Controls",
	// 		description:
	// 			"New enterprise-grade security features including role-based access control (RBAC), audit logging, and data encryption at rest. Ensure your data pipelines meet compliance requirements.",
	// 		releaseDate: "Released on 05-Dec-23",
	// 		tags: ["security", "feature"],
	// 	},
	// 	{
	// 		title: "Multi-Cloud Support",
	// 		description:
	// 			"Deploy and manage your data pipelines across AWS, GCP, and Azure from a single interface. Seamlessly move data between cloud providers with our unified connector framework.",
	// 		releaseDate: "Released on 01-Dec-23",
	// 		tags: ["feature"],
	// 	},
	// ],
}

const UpdatesModal = () => {
	const { showUpdatesModal, setShowUpdatesModal } = useAppStore()
	const [selectedCategory, setSelectedCategory] = useState(
		"OLake UI and Worker",
	)
	// TODO: Replace with actual API call to fetch updates from Docker
	const [updates] = useState<Record<string, UpdateItem[]>>(MOCK_UPDATES)

	const currentUpdates = updates[selectedCategory] || []

	return (
		<Modal
			open={showUpdatesModal}
			footer={null}
			closable={false}
			centered
			width={800}
			styles={{
				content: { padding: 0 },
			}}
		>
			<div className="flex h-[600px] flex-col">
				{/* Full Width Header */}
				<div className="flex items-center justify-between border-b border-gray-200 px-6 py-4">
					<div className="flex items-center gap-2">
						<div className="relative">
							<div className="absolute right-0 top-0 h-2 w-2 animate-pulse rounded-full bg-red-500"></div>
							<BellIcon
								size={20}
								color="#203FDD"
							/>
						</div>
						<span className="text-lg font-semibold text-gray-900">Updates</span>
					</div>
					<button
						onClick={() => setShowUpdatesModal(false)}
						className="rounded-md p-1 hover:bg-gray-100"
					>
						<XIcon
							size={20}
							color="#203FDD"
						/>
					</button>
				</div>

				{/* Content Area with Sidebar */}
				<div className="flex flex-1 overflow-hidden">
					{/* Left Sidebar */}
					<div className="w-[255px] border-r border-gray-200">
						{/* Categories */}
						{CATEGORIES.map(category => (
							<button
								key={category}
								onClick={() => setSelectedCategory(category)}
								className={clsx(
									"w-full border-b px-6 py-2.5 text-left text-sm transition-colors",
									selectedCategory === category
										? "bg-gray-50 font-medium text-gray-900"
										: "font-normal text-gray-700 hover:bg-gray-50",
								)}
							>
								{category}
							</button>
						))}
					</div>

					{/* Main Content */}
					<div className="flex flex-1 flex-col">
						{/* Category Title */}
						<div className="border-b border-gray-200 bg-white px-6 py-4">
							<h2 className="text-lg font-semibold text-gray-900">
								{selectedCategory}
							</h2>
						</div>
						{/* Updates List */}
						<div className="flex-1 overflow-y-auto px-6 py-4">
							{currentUpdates.length === 0 ? (
								<div className="flex h-full items-center justify-center text-sm text-gray-500">
									No updates available
								</div>
							) : (
								<div className="space-y-3">
									{currentUpdates.map((update, index) => (
										<div
											key={index}
											className="rounded-lg border border-gray-200 bg-white p-4 pb-0 shadow-sm hover:border-gray-300 hover:shadow-sm"
										>
											<div className="flex items-start justify-between">
												<div className="mb-3 flex-1">
													<div className="flex justify-between">
														<a
															href={update.releaseNotesLink}
															target="_blank"
															rel="noopener noreferrer"
															className="max-w-[60%] text-base font-semibold text-primary-500 underline hover:underline"
														>
															{update.version || update.title}
														</a>
														<div>
															{update.tags.length > 0 && (
																<div className="ml-3 flex gap-2">
																	{update.tags
																		.slice(0, 2)
																		.map((tag, tagIdx) => {
																			const style = getTagStyle(tag)
																			return (
																				<span
																					key={tagIdx}
																					className={`rounded px-2.5 py-1 text-xs font-medium capitalize ${style.bg} ${style.text}`}
																				>
																					{tag.replace(/-/g, " ")}
																				</span>
																			)
																		})}
																	{update.tags.length > 2 && (
																		<Tooltip
																			title={
																				<div className="flex flex-wrap gap-1">
																					{update.tags
																						.slice(2)
																						.map((tag, idx) => (
																							<span
																								key={idx}
																								className="text-xs capitalize"
																							>
																								{tag.replace(/-/g, " ")}
																								{idx < update.tags.length - 3 &&
																									","}
																							</span>
																						))}
																				</div>
																			}
																		>
																			<span className="cursor-help rounded bg-gray-100 px-2.5 py-1 text-xs font-medium text-gray-600">
																				+{update.tags.length - 2}
																			</span>
																		</Tooltip>
																	)}
																</div>
															)}
														</div>
													</div>
													<div className="mt-1 text-sm leading-relaxed [&>h2]:mb-3 [&>h2]:border-b [&>h2]:border-gray-200 [&>h2]:pb-2 [&>h2]:text-base [&>h2]:font-semibold [&>h2]:text-gray-900 [&>p]:mt-3 [&>ul>li]:pl-0 [&>ul>li]:before:mr-2 [&>ul>li]:before:content-['•'] [&>ul]:ml-0 [&>ul]:list-none [&>ul]:space-y-1.5">
														<ReactMarkdown
															components={{
																a: ({ href, children, ...props }) => {
																	return (
																		<a
																			href={href}
																			target="_blank"
																			rel="noopener noreferrer"
																			className="font-medium text-blue-600 hover:underline"
																			{...props}
																		>
																			{children}
																		</a>
																	)
																},
																strong: ({ ...props }) => (
																	<strong
																		{...props}
																		className="font-semibold text-gray-900"
																	/>
																),
															}}
														>
															{update.description}
														</ReactMarkdown>
													</div>

													<p className="mt-2 text-xs text-[#bcbcbc]">
														{update.releaseDate}
													</p>
												</div>
											</div>
										</div>
									))}
								</div>
							)}
						</div>
					</div>
				</div>
			</div>
		</Modal>
	)
}

export default UpdatesModal

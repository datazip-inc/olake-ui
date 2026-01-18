import {
	BellIcon,
	XIcon,
	CaretUpIcon,
	CaretDownIcon,
} from "@phosphor-icons/react"
import { Modal, Tooltip } from "antd"
import { useAppStore } from "../../../store"
import { useState, useEffect } from "react"
import clsx from "clsx"
import ReactMarkdown from "react-markdown"
import {
	ReleaseMetadataResponse,
	ReleaseType,
} from "../../../types/platformTypes"
import { usePlatformStore } from "../../../store/platformStore"

const CATEGORIES: ReleaseType[] = [
	ReleaseType.FEATURES,
	ReleaseType.OLAKE_UI_WORKER,
	ReleaseType.OLAKE_HELM,
	ReleaseType.OLAKE,
]

const RELEASE_TYPE_TO_LABEL: Record<ReleaseType, string> = {
	[ReleaseType.OLAKE_UI_WORKER]: "OLake UI and Worker",
	[ReleaseType.OLAKE_HELM]: "OLake Helm",
	[ReleaseType.OLAKE]: "OLake",
	[ReleaseType.FEATURES]: "Features",
}

const TAG_STYLES: Record<string, { bg: string; text: string }> = {
	"New Release": { bg: "bg-blue-50", text: "text-blue-600" },
	Performance: { bg: "bg-green-50", text: "text-green-600" },
	Stable: { bg: "bg-green-50", text: "text-green-600" },
	"Bug Fix": { bg: "bg-red-50", text: "text-red-600" },
	Feature: { bg: "bg-purple-50", text: "text-purple-600" },
	Security: { bg: "bg-orange-50", text: "text-orange-600" },
	Breaking: { bg: "bg-yellow-50", text: "text-yellow-700" },
	default: { bg: "bg-gray-50", text: "text-gray-600" },
}

const getTagStyle = (tag: string) => {
	return TAG_STYLES[tag] ?? TAG_STYLES.default
}

// Convert GitHub-style references to markdown links
const formatGithubReferences = (text: string): string => {
	// Replace full PR/issue URLs with [#123](url) format
	text = text.replace(
		/https:\/\/github\.com\/([^/]+)\/([^/]+)\/(pull|issues)\/(\d+)/g,
		"[#$4](https://github.com/$1/$2/$3/$4)",
	)

	// Replace full compare URLs with [v1.2.3...v1.2.4](url) format
	text = text.replace(
		/https:\/\/github\.com\/([^/]+)\/([^/]+)\/compare\/(v\d+\.\d+\.\d+)\.\.\.(v\d+\.\d+\.\d+)/g,
		"[$3...$4](https://github.com/$1/$2/compare/$3...$4)",
	)

	// Replace @username with GitHub profile links (only if not already in a markdown link)
	text = text.replace(
		/@([a-zA-Z0-9](?:[a-zA-Z0-9-]*[a-zA-Z0-9])?)/g,
		"[@$1](https://github.com/$1)",
	)

	return text
}

const UpdatesModal = () => {
	const { showUpdatesModal, setShowUpdatesModal } = useAppStore()
	const { releases, seenCategories, markCategoryAsSeen } = usePlatformStore()
	const [selectedCategory, setSelectedCategory] = useState<ReleaseType>(
		ReleaseType.OLAKE_UI_WORKER,
	)
	const [expandedCards, setExpandedCards] = useState<Set<number>>(new Set())

	const currentUpdates = releases?.[selectedCategory]?.releases || []

	useEffect(() => {
		if (showUpdatesModal) {
			markCategoryAsSeen(selectedCategory)
		}
	}, [showUpdatesModal, selectedCategory])

	const toggleCard = (index: number) => {
		setExpandedCards(prev => {
			const newSet = new Set(prev)
			if (newSet.has(index)) {
				newSet.delete(index)
			} else {
				newSet.add(index)
			}
			return newSet
		})
	}

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
						<BellIcon
							size={20}
							color="#203FDD"
						/>
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
						{CATEGORIES.map(category => {
							const hasNewReleases = releases?.[category]?.releases.some(
								(r: ReleaseMetadataResponse) => r.tags.includes("New Release"),
							)
							const showDot =
								hasNewReleases && !seenCategories.includes(category)

							return (
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
									<div className="flex items-center justify-between">
										{RELEASE_TYPE_TO_LABEL[category]}
										{showDot && (
											<div className="h-2 w-2 rounded-full bg-red-500" />
										)}
									</div>
								</button>
							)
						})}
					</div>

					{/* Main Content */}
					<div className="flex flex-1 flex-col">
						{/* Category Title */}
						<div className="border-b border-gray-200 bg-white px-6 py-4">
							<h2 className="text-lg font-semibold text-gray-900">
								{RELEASE_TYPE_TO_LABEL[selectedCategory]}
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
									{currentUpdates.map(
										(update: ReleaseMetadataResponse, index: number) => {
											const isExpanded = expandedCards.has(index)
											const updateKey = `${update.version || update.title}-${index}`
											return (
												<div
													key={updateKey}
													className="rounded-lg border border-gray-200 bg-white shadow-sm hover:border-gray-300 hover:shadow-sm"
												>
													<button
														onClick={() => toggleCard(index)}
														className="w-full p-4 pb-3 text-left"
													>
														<div className="flex items-start justify-between">
															<div className="flex-1">
																<div className="flex items-center justify-between gap-2">
																	<a
																		href={update.link}
																		target="_blank"
																		rel="noopener noreferrer"
																		onClick={e => e.stopPropagation()}
																		className="max-w-[60%] text-base font-semibold text-primary-500 underline hover:underline"
																	>
																		{update.version || update.title}
																	</a>
																	<div className="flex items-center gap-2">
																		{update.tags.length > 0 && (
																			<div className="flex gap-2">
																				{update.tags
																					.slice(0, 2)
																					.map(
																						(tag: string, tagIdx: number) => {
																							const style = getTagStyle(tag)
																							return (
																								<span
																									key={`${updateKey}-tag-${tagIdx}`}
																									className={`rounded px-2.5 py-1 text-xs font-medium capitalize ${style.bg} ${style.text}`}
																								>
																									{tag}
																								</span>
																							)
																						},
																					)}
																				{update.tags.length > 2 && (
																					<Tooltip
																						title={
																							<div className="flex flex-wrap gap-1">
																								{update.tags
																									.slice(2)
																									.map(
																										(
																											tag: string,
																											idx: number,
																										) => (
																											<span
																												key={`${updateKey}-tooltip-tag-${idx}`}
																												className="text-xs capitalize"
																											>
																												{tag}
																												{idx <
																													update.tags.length -
																														3 && ","}
																											</span>
																										),
																									)}
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
																<div className="mt-2 flex items-center justify-between">
																	<p className="text-xs text-neutral-400">
																		{update.date}
																	</p>
																	{isExpanded ? (
																		<CaretUpIcon
																			size={16}
																			weight="bold"
																			className="text-neutral-400"
																		/>
																	) : (
																		<CaretDownIcon
																			size={16}
																			weight="bold"
																			className="text-neutral-400"
																		/>
																	)}
																</div>
															</div>
														</div>
													</button>
													{isExpanded && (
														<div className="border-t border-gray-100 px-4 pb-4 pt-3">
															<div className="text-sm leading-relaxed [&>h2]:mb-3 [&>h2]:border-b [&>h2]:border-gray-200 [&>h2]:pb-2 [&>h2]:text-base [&>h2]:font-semibold [&>h2]:text-gray-900 [&>p]:mt-3 [&>ul>li]:pl-0 [&>ul>li]:before:mr-2 [&>ul>li]:before:content-['•'] [&>ul]:ml-0 [&>ul]:list-none [&>ul]:space-y-1.5">
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
																	{formatGithubReferences(update.description)}
																</ReactMarkdown>
															</div>
														</div>
													)}
												</div>
											)
										},
									)}
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

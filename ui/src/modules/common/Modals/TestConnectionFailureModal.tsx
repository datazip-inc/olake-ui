import { useState } from "react"
import { useNavigate } from "react-router-dom"
import { message, Modal } from "antd"
import { CopySimpleIcon } from "@phosphor-icons/react"
import clsx from "clsx"

import { useAppStore } from "../../../store"
import ErrorIcon from "../../../assets/ErrorIcon.svg"
import { getLogTextColor, getLogLevelClass } from "../../../utils/utils"

const TestConnectionFailureModal = ({
	fromSources,
}: {
	fromSources: boolean
}) => {
	const {
		showFailureModal,
		setShowFailureModal,
		sourceTestConnectionError,
		destinationTestConnectionError,
	} = useAppStore()
	const [isExpanded, setIsExpanded] = useState(false)
	const navigate = useNavigate()

	const handleCancel = () => {
		setShowFailureModal(false)
		setIsExpanded(false)
	}

	const handleBackToPath = () => {
		setShowFailureModal(false)
		setIsExpanded(false)
		if (fromSources) {
			navigate("/sources")
		} else {
			navigate("/destinations")
		}
	}

	const handleReadMore = () => setIsExpanded(!isExpanded)

	const handleCopyLogs = async () => {
		try {
			await navigator.clipboard.writeText(
				JSON.stringify(
					fromSources
						? sourceTestConnectionError?.logs || []
						: destinationTestConnectionError?.logs || [],
					null,
					4,
				),
			)
			message.success("Logs copied to clipboard!")
		} catch {
			message.error("Failed to copy logs")
		}
	}

	return (
		<Modal
			open={showFailureModal}
			footer={null}
			closable={false}
			centered
			width={isExpanded ? 980 : 680}
			className="transition-all duration-300"
		>
			<div
				className={`flex flex-col items-center justify-start gap-7 overflow-hidden pb-6 transition-all duration-300 ease-in-out ${
					isExpanded ? "w-full pt-6" : "mx-auto max-w-[680px] pt-16"
				}`}
			>
				<div className="relative">
					<div>
						<img
							src={ErrorIcon}
							alt="Error"
						/>
					</div>
				</div>
				<div className="flex w-full flex-col items-center">
					<p className="text-sm text-text-tertiary">Failed</p>
					<h2 className="text-center text-xl font-medium">
						Your test connection has failed
					</h2>
					<div className="mt-4 flex w-full flex-col rounded-md border border-neutral-300 text-sm">
						<div className="flex w-full items-center justify-between border-b border-neutral-300 px-3 py-2">
							<div className="font-bold">Error </div>
							{isExpanded && (
								<CopySimpleIcon
									onClick={handleCopyLogs}
									className="size-[14px] flex-shrink-0 cursor-pointer"
								/>
							)}
						</div>
						<div
							className={`flex flex-col px-3 py-2 text-neutral-500 ${
								isExpanded ? "h-[300px] overflow-auto" : "h-auto"
							}`}
						>
							{!isExpanded ? (
								<div className="max-h-[150px] overflow-auto text-red-500">
									{fromSources
										? sourceTestConnectionError?.message || ""
										: destinationTestConnectionError?.message || ""}
								</div>
							) : (
								<table className="min-w-full">
									<tbody>
										{(fromSources
											? sourceTestConnectionError?.logs || []
											: destinationTestConnectionError?.logs || []
										).map((jobLog, index) => (
											<tr key={index}>
												<td className="w-24 px-4 py-1 text-sm">
													<span
														className={clsx(
															"rounded-xl px-2 py-[5px] text-xs capitalize",
															getLogLevelClass(jobLog.level),
														)}
													>
														{jobLog.level}
													</span>
												</td>
												<td
													className={clsx(
														"whitespace-pre-wrap break-words px-4 py-3 text-sm text-gray-700",
														getLogTextColor(jobLog.level),
													)}
												>
													{jobLog.message}
												</td>
											</tr>
										))}
									</tbody>
								</table>
							)}

							{!isExpanded && (
								<button
									type="button"
									onClick={handleReadMore}
									aria-label="Read more"
									aria-expanded={isExpanded}
									className="mt-2 text-left text-blue-600 hover:underline"
								>
									Read more
								</button>
							)}
						</div>
					</div>
				</div>
				<div className="flex items-center gap-4">
					<button
						onClick={handleBackToPath}
						className="w-fit rounded-md border border-[#d9d9d9] px-4 py-2 text-black"
					>
						{fromSources ? "Back to Sources Page" : "Back to Destinations Page"}
					</button>
					<button
						onClick={handleCancel}
						className="w-fit flex-1 rounded-md border border-danger px-4 py-2 text-danger"
					>
						{fromSources ? "Edit  Source" : "Edit  Destination"}
					</button>
				</div>
			</div>
		</Modal>
	)
}

export default TestConnectionFailureModal

import { useState, useRef, useEffect } from "react"
import { Button, Tooltip } from "antd"
import {
	CornersOut,
	CaretRight,
	Info,
	ArrowSquareOut,
} from "@phosphor-icons/react"
import { DocumentationPanelProps } from "../../../types"

const DocumentationPanel: React.FC<DocumentationPanelProps> = ({
	docUrl,
	isMinimized = false,
	onToggle,
	showResizer = true,
	initialWidth = 40,
}) => {
	const [docPanelWidth] = useState(initialWidth)
	const [isDocPanelCollapsed, setIsDocPanelCollapsed] = useState(isMinimized)
	const [isLoading, setIsLoading] = useState(true)
	const [isReady, setIsReady] = useState(false)

	const iframeRef = useRef<HTMLIFrameElement>(null)
	const panelRef = useRef<HTMLDivElement>(null)

	// Sync collapsed state with isMinimized prop
	useEffect(() => {
		setIsDocPanelCollapsed(isMinimized)
	}, [isMinimized])

	// Reset loading state when docUrl changes
	useEffect(() => {
		setIsLoading(true)
		setIsReady(false)
		if (iframeRef.current) {
			iframeRef.current.src = docUrl
		}
	}, [docUrl])

	// Handle iframe load event
	useEffect(() => {
		const iframe = iframeRef.current
		if (!iframe) return

		const handleLoad = () => {
			// Post message to iframe for theming
			iframe.contentWindow?.postMessage({ theme: "light" }, "https://olake.io")

			// Set loading states with slight delay for animations
			setTimeout(() => {
				setIsLoading(false)
				setTimeout(() => {
					setIsReady(true)
				}, 50)
			}, 100)
		}

		iframe.addEventListener("load", handleLoad)
		return () => iframe.removeEventListener("load", handleLoad)
	}, [docUrl])

	const toggleDocPanel = () => {
		setIsDocPanelCollapsed(!isDocPanelCollapsed)
		onToggle?.()
	}

	const openInNewTab = () => {
		window.open(docUrl, "_blank")
	}

	// Show only the button when panel is collapsed and resizer is hidden
	if (isDocPanelCollapsed && !showResizer) {
		return (
			<div className="fixed bottom-6 right-6 flex gap-2">
				<Button
					type="default"
					className="flex items-center"
					onClick={openInNewTab}
					icon={
						<ArrowSquareOut
							size={16}
							className="mr-2"
						/>
					}
				>
					Open Docs
				</Button>
				<Button
					type="primary"
					className="flex items-center bg-blue-600"
					onClick={toggleDocPanel}
					icon={
						<CornersOut
							size={16}
							className="mr-2"
						/>
					}
				>
					Show Documentation
				</Button>
			</div>
		)
	}

	return (
		<>
			{showResizer && (
				<div
					className="relative z-10"
					style={{ width: isDocPanelCollapsed ? "16px" : "0" }}
				>
					<button
						onClick={toggleDocPanel}
						className="absolute right-0 top-1/2 z-10 -translate-y-1/2 translate-x-1/2 rounded-xl border border-gray-200 bg-white p-2.5 text-gray-900 shadow-[0_6px_16px_0_rgba(0,0,0,0.08)] hover:text-gray-700 focus:outline-none"
					>
						<div
							className={`transition-transform duration-300 ${
								isDocPanelCollapsed ? "rotate-180" : "rotate-0"
							}`}
						>
							<CaretRight size={16} />
						</div>
					</button>
				</div>
			)}

			{/* Documentation panel */}
			<div
				ref={panelRef}
				className="relative overflow-hidden border-l-4 border-gray-200 bg-white transition-all duration-500 ease-in-out"
				style={{ width: isDocPanelCollapsed ? "80px" : `${docPanelWidth}%` }}
			>
				<div
					className={`transition-opacity ${!isReady ? "opacity-0" : "h-full opacity-100"}`}
					style={{ transition: "opacity 0.3s ease" }}
				>
					{!isDocPanelCollapsed && (
						<div className="absolute right-16 top-3.5 z-10">
							<Button
								type="default"
								icon={
									<ArrowSquareOut
										size={20}
										weight="bold"
										className="text-primary"
									/>
								}
								onClick={openInNewTab}
								className="flex items-center gap-2 border border-gray-200 px-2 py-3 hover:border-blue-600 hover:text-blue-600"
							>
								Open Docs
							</Button>
						</div>
					)}
					<iframe
						ref={iframeRef}
						src={docUrl}
						className="h-full w-full border-none"
						title="Documentation"
						sandbox="allow-scripts allow-same-origin allow-popups allow-forms"
						data-theme="light"
						style={{
							visibility:
								isDocPanelCollapsed || isLoading ? "hidden" : "visible",
						}}
					/>
					{isDocPanelCollapsed && (
						<div className="flex h-full w-full items-start justify-center">
							<div className="absolute right-3 top-10 z-10 flex flex-col gap-2">
								<div className="rounded-xl border border-gray-200 bg-neutral-light p-2">
									<Info
										size={25}
										className="cursor-pointer text-primary transition-all duration-300 ease-in-out hover:text-primary/80"
										onClick={toggleDocPanel}
									/>
								</div>
								<div className="rounded-xl border border-gray-200 bg-neutral-light p-2">
									<Tooltip
										title="Open documentation in new tab"
										placement="left"
									>
										<ArrowSquareOut
											size={25}
											className="cursor-pointer text-primary transition-all duration-300 ease-in-out hover:text-primary/80"
											onClick={openInNewTab}
										/>
									</Tooltip>
								</div>
							</div>
						</div>
					)}
				</div>
			</div>

			<style>{`
				.resizing {
					transition: none !important;
				}
			`}</style>
		</>
	)
}

export default DocumentationPanel

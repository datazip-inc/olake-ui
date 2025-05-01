import { useState, useRef, useEffect } from "react"
import { Button } from "antd"
import {
	DotsThreeVertical,
	CornersOut,
	CaretRight,
} from "@phosphor-icons/react"

interface DocumentationPanelProps {
	docUrl: string
	isMinimized?: boolean
	onToggle?: () => void
	showResizer?: boolean
	initialWidth?: number
}

const DocumentationPanel: React.FC<DocumentationPanelProps> = ({
	docUrl,
	isMinimized = false,
	onToggle,
	showResizer = true,
	initialWidth = 30,
}) => {
	const [docPanelWidth, setDocPanelWidth] = useState(initialWidth)
	const [isDocPanelCollapsed, setIsDocPanelCollapsed] = useState(isMinimized)
	const [isLoading, setIsLoading] = useState(true)
	const [isReady, setIsReady] = useState(false)
	const resizerRef = useRef<HTMLDivElement>(null)
	const iframeRef = useRef<HTMLIFrameElement>(null)
	const panelRef = useRef<HTMLDivElement>(null)
	const isDragging = useRef(false)
	const animationFrame = useRef<number>()

	useEffect(() => {
		setIsDocPanelCollapsed(isMinimized)
	}, [isMinimized])

	useEffect(() => {
		setIsLoading(true)
		setIsReady(false)
		if (iframeRef.current) {
			iframeRef.current.src = docUrl
		}
	}, [docUrl])

	useEffect(() => {
		const iframe = iframeRef.current
		if (!iframe) return

		const handleLoad = () => {
			iframe.contentWindow?.postMessage({ theme: "light" }, "https://olake.io")
			setTimeout(() => {
				setIsLoading(false)
				setTimeout(() => {
					setIsReady(true)
				}, 50)
			}, 100)
		}

		iframe.addEventListener("load", handleLoad)
		return () => {
			iframe.removeEventListener("load", handleLoad)
		}
	}, [docUrl])

	const handleResizeStart = (e: React.MouseEvent<HTMLDivElement>) => {
		e.preventDefault()
		e.stopPropagation() // Prevent click event from firing

		const startX = e.clientX
		const panel = panelRef.current
		const startWidth = panel?.getBoundingClientRect().width || 0
		const containerWidth = window.innerWidth

		isDragging.current = true
		panel?.classList.add("resizing")

		const updateWidth = (clientX: number) => {
			if (!panel) return
			const delta = startX - clientX
			const newWidthPx = startWidth + delta
			const newWidthPercent = Math.max(
				15,
				Math.min(75, (newWidthPx / containerWidth) * 100),
			)
			panel.style.width = `${newWidthPercent}%`
		}

		const onMouseMove = (e: MouseEvent) => {
			if (!isDragging.current) return
			if (animationFrame.current) cancelAnimationFrame(animationFrame.current)
			animationFrame.current = requestAnimationFrame(() =>
				updateWidth(e.clientX),
			)
		}

		const onMouseUp = () => {
			isDragging.current = false
			if (animationFrame.current) cancelAnimationFrame(animationFrame.current)

			const widthStr = panel?.style.width.replace("%", "")
			if (widthStr) {
				setDocPanelWidth(parseFloat(widthStr))
			}

			panel?.classList.remove("resizing")
			document.removeEventListener("mousemove", onMouseMove)
			document.removeEventListener("mouseup", onMouseUp)
		}

		document.addEventListener("mousemove", onMouseMove)
		document.addEventListener("mouseup", onMouseUp)
	}

	const toggleDocPanel = () => {
		setIsDocPanelCollapsed(!isDocPanelCollapsed)
		if (onToggle) {
			onToggle()
		}
	}

	if (isDocPanelCollapsed && !showResizer) {
		return (
			<div className="fixed bottom-6 right-6">
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
					style={{
						width: isDocPanelCollapsed ? "16px" : "0",
					}}
				>
					<div
						ref={resizerRef}
						className="group absolute left-0 top-1/2 flex h-20 w-4 -translate-y-1/2 cursor-ew-resize items-center justify-center"
						onMouseDown={handleResizeStart}
						onClick={e => {
							e.stopPropagation()
							toggleDocPanel()
						}}
					>
						<DotsThreeVertical
							size={16}
							className="text-gray-500"
						/>
					</div>

					<button
						onClick={toggleDocPanel}
						className="absolute bottom-10 right-0 z-10 translate-x-1/2 rounded-xl border border-gray-200 bg-white p-2.5 text-[#383838] shadow-[0_6px_16px_0_rgba(0,0,0,0.08)] hover:text-gray-700 focus:outline-none"
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
				className="overflow-hidden border-l-4 border-gray-200 bg-white transition-all duration-500 ease-in-out"
				style={{
					width: isDocPanelCollapsed ? "30px" : `${docPanelWidth}%`,
				}}
			>
				<div
					className={`transition-opacity ${!isReady ? "opacity-0" : "h-full opacity-100"}`}
					style={{ transition: "opacity 0.3s ease" }}
				>
					<iframe
						ref={iframeRef}
						src={docUrl}
						className="h-full w-full border-none"
						title="Documentation"
						sandbox="allow-scripts allow-same-origin allow-popups allow-forms"
						data-theme="light"
						style={{ visibility: isLoading ? "hidden" : "visible" }}
					/>
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

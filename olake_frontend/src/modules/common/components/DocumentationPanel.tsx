import { useState, useRef, useEffect } from "react"
import { Button } from "antd"
import { DotsThreeVertical, CornersOut,ArrowsIn, ArrowsOut } from "@phosphor-icons/react"

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
	const resizerRef = useRef<HTMLDivElement>(null)
	const iframeRef = useRef<HTMLIFrameElement>(null)
	const lastWidth = useRef(initialWidth);
	const [isResizing, setIsResizing] = useState(false);
	const resizeRequestRef = useRef<number | null>(null);
	const [isMaximized, setIsMaximized] = useState(false)

	useEffect(() => {
		setIsDocPanelCollapsed(isMinimized)
	}, [isMinimized])

	useEffect(() => {
		if (iframeRef.current) {
			iframeRef.current.src = docUrl
		}
	}, [docUrl])

	const handleResizeStart = (e: React.MouseEvent<HTMLDivElement>) => {
		e.preventDefault()
		e.stopPropagation() // Prevent click event from firing
		setIsResizing(true)

		const startX = e.clientX
		const startWidth = docPanelWidth

		const handleMouseMove = (e: MouseEvent) => {
			if (resizeRequestRef.current) return // Prevent redundant updates

			resizeRequestRef.current = window.requestAnimationFrame(() => {
				const containerWidth = window.innerWidth
				const newWidth = Math.max(
					15,
					Math.min(
						75,
						startWidth - ((e.clientX - startX) / containerWidth) * 100,
					),
				)

				lastWidth.current = newWidth
				setDocPanelWidth(newWidth)
				resizeRequestRef.current = null
			})
		}

		const handleMouseUp = () => {
			setIsResizing(false)
			document.removeEventListener("mousemove", handleMouseMove)
			document.removeEventListener("mouseup", handleMouseUp)
			if (resizeRequestRef.current) {
				window.cancelAnimationFrame(resizeRequestRef.current)
				resizeRequestRef.current = null
			}
		}

		document.addEventListener("mousemove", handleMouseMove)
		document.addEventListener("mouseup", handleMouseUp)
	}

	const toggleDocPanel = () => {
		setIsDocPanelCollapsed(!isDocPanelCollapsed)
		if (onToggle) {
			onToggle()
		}
	}
	// For minimizing and maximizing the Documentation
	const toggleSize = () => {
		if (isMaximized) {
			setDocPanelWidth(35)
		} else {
			setDocPanelWidth(75)
		}
		setIsMaximized(!isMaximized)
	};

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
			{/* Resizer Handle */}
			{showResizer && (
				<div
					className="relative z-10"
					style={{
						position: "relative",
						width: isDocPanelCollapsed ? "16px" : "0",
					}}
				>
					<div
						ref={resizerRef}
						className={`absolute left-0 top-1/2 flex h-20 w-4 -translate-y-1/2 cursor-ew-resize items-center justify-center transition-transform duration-300 ${
							isResizing ? "scale-125" : "hover:scale-110"
						}`}
						onMouseDown={handleResizeStart}
						onClick={e => {
							e.stopPropagation()
							toggleDocPanel()
						}}
					>
						<DotsThreeVertical
							size={16}
							className="text-gray-500 transition-opacity duration-200 hover:opacity-75"
						/>
					</div> 
				</div>
			)}

			{/* Documentation panel */}
			<div
				className="overflow-hidden border-l-4 border-gray-200 bg-white shadow-lg"
				style={{
					// Changing white to Soft purple 
					borderLeft: "4px solid #C6BEEE",
					width: isDocPanelCollapsed
						? "30px"
						: showResizer
							? `${docPanelWidth}%`
							: "25%",
					transition: isResizing
						? "none"
						: "width 0.5s cubic-bezier(0.22, 1, 0.36, 1), opacity 0.3s ease-in-out, visibility 0.3s ease-in-out",
					opacity: isDocPanelCollapsed ? 0 : 1,
					visibility: isDocPanelCollapsed ? "hidden" : "visible",
				}}
			>
				{/* Button for toggling setting to minimize and maximize */}
				<div className="absolute top-2 right-2">
					<Button
						type="text"
						onClick={toggleSize}
						icon={isMaximized ? <ArrowsIn size={20} /> : <ArrowsOut size={20} />}
						className="hover:bg-gray-200 rounded-full"
					/>
				</div>
				<iframe
					ref={iframeRef}
					src={docUrl}
					className="h-full w-full border-none"
					title="Documentation"
					sandbox="allow-scripts allow-same-origin allow-popups allow-forms"
				/>
			</div>
		</>
	)
}

export default DocumentationPanel

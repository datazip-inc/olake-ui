import { CopySimpleIcon } from "@phosphor-icons/react"
import { Button, Modal } from "antd"

import { ErrorIcon } from "@/assets"
import { copyToClipboard } from "@/common/utils"

const ErrorLogsModal = ({
	open,
	onClose,
	title,
	subtitle = "Please check your connection and try again",
	error,
	onAction,
	actionButtonText = "Try Again",
}: {
	open: boolean
	onClose: () => void
	title: string
	subtitle?: React.ReactNode
	error: string
	onAction?: () => void
	actionButtonText?: string
}) => {
	const handleAction = () => {
		onAction?.()
	}

	const handleCopyLogs = async () => {
		await copyToClipboard(error)
	}

	const handleClose = () => {
		onClose()
	}

	return (
		<Modal
			open={open}
			footer={null}
			closable={false}
			centered
			width={680}
			destroyOnHidden
			styles={{
				content: {
					padding: 0,
					overflow: "hidden",
					borderRadius: 20,
				},
				body: {
					padding: 0,
				},
			}}
		>
			<div className="flex max-h-[calc(100vh-64px)] min-h-[560px] flex-col bg-white">
				<div className="mt-12 flex w-full max-w-[80%] shrink-0 flex-col items-center gap-x-3 gap-y-1 self-center px-4 text-center">
					<img
						src={ErrorIcon}
						alt="Error"
					/>
					<h2 className="text-xl font-medium leading-7 text-olake-text">
						{title}
					</h2>
					{subtitle && (
						<p className="text-sm leading-[22px] text-olake-text-tertiary">
							{subtitle}
						</p>
					)}
				</div>

				<div className="mt-4 flex min-h-0 w-[573px] flex-1 flex-col self-center overflow-hidden rounded-lg bg-olake-surface-muted">
					<div className="flex shrink-0 items-start justify-between px-4 py-4">
						<div className="text-sm font-bold leading-[22px] text-olake-text">
							Error
						</div>
						<button
							type="button"
							onClick={handleCopyLogs}
							className="group flex items-center gap-1 text-xs font-medium leading-5 text-olake-text-secondary"
						>
							<CopySimpleIcon
								className="group-hover:scale-105 group-active:scale-95"
								size={16}
							/>
							<span>Copy Error</span>
						</button>
					</div>

					<div className="min-h-0 flex-1 overflow-auto px-4 pb-4">
						<pre className="whitespace-pre-wrap font-mono text-xs leading-4 text-olake-error">
							{error}
						</pre>
					</div>
				</div>

				<div className="mb-6 mt-6 flex w-[573px] shrink-0 items-center justify-start gap-2 self-center">
					{onAction && (
						<Button
							type="primary"
							onClick={handleAction}
						>
							{actionButtonText}
						</Button>
					)}
					<Button onClick={handleClose}>Close</Button>
				</div>
			</div>
		</Modal>
	)
}

export default ErrorLogsModal

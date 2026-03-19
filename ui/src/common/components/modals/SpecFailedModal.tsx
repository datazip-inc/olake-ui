import { CopySimpleIcon } from "@phosphor-icons/react"
import { Button, Modal } from "antd"

import { ErrorIcon } from "@/assets"
import { copyToClipboard } from "@/common/utils"

const SpecFailedModal = ({
	open,
	onClose,
	fromSource,
	connectionType,
	error,
	onTryAgain,
}: {
	open: boolean
	onClose: () => void
	fromSource?: boolean
	connectionType?: "source" | "destination" | "catalog"
	error: string
	onTryAgain: () => void
}) => {
	const resolvedConnectionType =
		connectionType ?? (fromSource ? "source" : "destination")
	const labelMap = {
		source: "Source",
		destination: "Destination",
		catalog: "Catalog",
	} as const
	const label = labelMap[resolvedConnectionType]

	const handleTryAgain = () => {
		onClose()
		onTryAgain()
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
				<div className="mt-12 flex w-[261px] shrink-0 flex-col items-center gap-3 self-center text-center">
					<img
						src={ErrorIcon}
						alt="Error"
					/>
					<p className="text-sm leading-[22px] text-olake-text-tertiary">
						Failed
					</p>
					<h2 className="text-xl font-medium leading-7 text-olake-text">
						{label} Spec Load Failed
					</h2>
				</div>

				<div className="mt-4 flex min-h-0 w-[573px] flex-1 flex-col self-center overflow-hidden rounded-lg bg-olake-surface-muted">
					<div className="flex h-[73px] shrink-0 items-start justify-between px-4 pb-0 pt-4">
						<div className="text-sm font-bold leading-[22px] text-olake-text">
							Error
						</div>
						<button
							type="button"
							onClick={handleCopyLogs}
							className="text-olake-text-secondary"
						>
							<CopySimpleIcon className="size-[14px] flex-shrink-0" />
						</button>
					</div>

					<div className="min-h-0 flex-1 overflow-auto px-4 pb-4">
						<pre className="whitespace-pre-wrap font-mono text-xs leading-4 text-olake-error">
							{error}
						</pre>
					</div>
				</div>

				<div className="mb-6 mt-6 flex w-[573px] shrink-0 items-center gap-2 self-center">
					<Button onClick={handleClose}>Close</Button>
					<Button
						type="primary"
						onClick={handleTryAgain}
					>
						Try Again
					</Button>
				</div>
			</div>
		</Modal>
	)
}

export default SpecFailedModal

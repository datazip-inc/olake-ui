import { CopySimpleIcon } from "@phosphor-icons/react"
import { Modal } from "antd"

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
			className="transition-all duration-300"
		>
			<div className="mx-auto flex max-w-[680px] flex-col items-center justify-start gap-7 overflow-hidden pb-6 pt-16 transition-all duration-300 ease-in-out">
				<div className="relative">
					<div>
						<img
							src={ErrorIcon}
							alt="Error"
						/>
					</div>
				</div>
				<div className="flex w-full flex-col items-center">
					<p className="text-sm text-olake-text-tertiary">Failed</p>
					<h2 className="text-center text-xl font-medium">
						{label} Spec Load Failed
					</h2>
					<div className="mt-4 flex w-full flex-col rounded-md border border-olake-border text-sm">
						<div className="flex w-full items-center justify-between border-b border-olake-border px-3 py-2">
							<div className="font-bold">Error </div>
							<CopySimpleIcon
								onClick={handleCopyLogs}
								className="size-[14px] flex-shrink-0 cursor-pointer"
							/>
						</div>
						<div className="flex h-auto flex-col px-3 py-2 text-olake-body-secondary">
							<div className="max-h-[150px] overflow-auto text-olake-error">
								{error}
							</div>
						</div>
					</div>
				</div>
				<div className="flex items-center gap-4">
					<button
						onClick={handleClose}
						className="w-fit rounded-md border border-olake-border px-4 py-2 text-olake-text"
					>
						Close
					</button>
					<button
						onClick={handleTryAgain}
						className="w-fit rounded-md border border-olake-primary px-4 py-2 text-olake-primary"
					>
						Try Again
					</button>
				</div>
			</div>
		</Modal>
	)
}

export default SpecFailedModal

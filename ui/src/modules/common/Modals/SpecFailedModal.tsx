import { message, Modal } from "antd"
import { CopySimpleIcon } from "@phosphor-icons/react"

import { useAppStore } from "../../../store"
import ErrorIcon from "../../../assets/ErrorIcon.svg"

const SpecFailedModal = ({
	fromSource,
	error,
	onTryAgain,
}: {
	fromSource: boolean
	error: string
	onTryAgain: () => void
}) => {
	const { showSpecFailedModal, setShowSpecFailedModal } = useAppStore()

	const handleTryAgain = () => {
		setShowSpecFailedModal(false)
		onTryAgain()
	}

	const handleCopyLogs = async () => {
		try {
			await navigator.clipboard.writeText(error)
			message.success("Logs copied to clipboard!")
		} catch {
			message.error("Failed to copy logs")
		}
	}

	const handleClose = () => {
		setShowSpecFailedModal(false)
	}

	return (
		<Modal
			open={showSpecFailedModal}
			footer={null}
			closable={false}
			centered
			width={680}
			className="transition-all duration-300"
		>
			<div
				className={`mx-auto flex max-w-[680px] flex-col items-center justify-start gap-7 overflow-hidden pb-6 pt-16 transition-all duration-300 ease-in-out`}
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
						{fromSource ? "Source" : "Destination"} Spec Load Failed
					</h2>
					<div className="mt-4 flex w-full flex-col rounded-md border border-neutral-300 text-sm">
						<div className="flex w-full items-center justify-between border-b border-neutral-300 px-3 py-2">
							<div className="font-bold">Error </div>
							<CopySimpleIcon
								onClick={handleCopyLogs}
								className="size-[14px] flex-shrink-0 cursor-pointer"
							/>
						</div>
						<div className={`flex h-auto flex-col px-3 py-2 text-neutral-500`}>
							<div className="max-h-[150px] overflow-auto text-red-500">
								{error}
							</div>
						</div>
					</div>
				</div>
				<div className="flex items-center gap-4">
					<button
						onClick={handleClose}
						className="w-fit rounded-md border border-[#d9d9d9] px-4 py-2 text-black"
					>
						Close
					</button>
					<button
						onClick={handleTryAgain}
						className="w-fit rounded-md border border-primary-500 px-4 py-2 text-primary-500"
					>
						Try Again
					</button>
				</div>
			</div>
		</Modal>
	)
}

export default SpecFailedModal

import { Warning } from "@phosphor-icons/react"
import { Button, Modal } from "antd"
import { useAppStore } from "../../../store"
import { ResetStreamsModalProps } from "../../../types"
import { FC } from "react"

const ResetStreamsModal: FC<ResetStreamsModalProps> = ({ onConfirm }) => {
	const { showResetStreamsModal, setShowResetStreamsModal } = useAppStore()

	const handleCancel = () => setShowResetStreamsModal(false)

	const handleConfirm = () => {
		setShowResetStreamsModal(false)
		onConfirm()
	}

	return (
		<Modal
			open={showResetStreamsModal}
			onCancel={handleCancel}
			footer={null}
			closable={false}
			centered
			title={
				<div className="flex items-center gap-2 text-danger">
					<Warning
						className="size-6"
						weight="fill"
					/>
					<span>Go back?</span>
				</div>
			}
		>
			<div className="flex flex-col items-center gap-6">
				<p className="font-medium text-slate-700">
					Going back will{" "}
					<span className="font-semibold">reset all stream configurations</span>{" "}
					and any changes will be lost.
				</p>

				<div className="flex w-full justify-end gap-3">
					<Button onClick={handleCancel}>Cancel</Button>
					<Button
						type="primary"
						danger
						onClick={handleConfirm}
					>
						Go Back
					</Button>
				</div>
			</div>
		</Modal>
	)
}

export default ResetStreamsModal

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
					<span>Your changes will not be saved</span>
				</div>
			}
		>
			<div className="flex flex-col items-center gap-6">
				<div className="flex w-full flex-col">
					<p className="font-medium text-slate-700">
						Leaving this page will loose all your progress & changes
					</p>
					<p className="font-medium text-slate-700">
						Are you sure want to leave?
					</p>
				</div>

				<div className="flex w-full justify-end gap-3">
					<Button
						type="primary"
						danger
						onClick={handleConfirm}
					>
						Yes, Leave
					</Button>
					<Button onClick={handleCancel}>No</Button>
				</div>
			</div>
		</Modal>
	)
}

export default ResetStreamsModal

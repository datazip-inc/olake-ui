import { FC } from "react"
import { WarningIcon } from "@phosphor-icons/react"
import { Button, Modal } from "antd"

import { useAppStore } from "../../../store"
import { ResetStreamsModalProps } from "../../../types"

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
					<WarningIcon
						className="size-9"
						weight="fill"
					/>
				</div>
			}
		>
			<div className="flex flex-col items-center gap-6">
				<div className="flex w-full flex-col">
					<p className="text-xl font-medium text-slate-700">
						Your changes will not be saved
					</p>
					<p className="text-sm text-slate-700">
						Leaving this page will loose all your progress & changes.
					</p>
					<p className="mt-6">Are you sure you want to leave?</p>
				</div>

				<div className="flex w-full justify-end gap-3">
					<Button
						type="primary"
						className="!bg-[#f5222d]"
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

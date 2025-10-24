import { useNavigate } from "react-router-dom"
import { WarningIcon } from "@phosphor-icons/react"
import { Button, message, Modal } from "antd"
import { useAppStore } from "../../../store"

const ClearDataModal = () => {
	const { showClearDataModal, setShowClearDataModal } = useAppStore()
	const navigate = useNavigate()

	return (
		<Modal
			open={showClearDataModal}
			footer={null}
			closable={false}
			centered
		>
			<div className="flex w-full flex-col items-center justify-center gap-8">
				<WarningIcon
					className="size-16 text-danger"
					weight="fill"
				/>

				<div className="text-center text-xl font-medium text-gray-950">
					Clear data will delete all data in your job.
				</div>

				<div className="flex w-full justify-end gap-4">
					<Button
						type="primary"
						danger
						onClick={() => {
							setShowClearDataModal(false)
							message.success("Clearing data")
							navigate("/jobs")
						}}
					>
						Clear Data
					</Button>
					<Button
						type="default"
						onClick={() => setShowClearDataModal(false)}
					>
						Cancel
					</Button>
				</div>
			</div>
		</Modal>
	)
}

export default ClearDataModal

import { Button, message, Modal } from "antd"
import { useAppStore } from "../../../store"
import { Warning } from "@phosphor-icons/react"
import { useNavigate } from "react-router-dom"

const ClearDestinationAndSyncModal = () => {
	const {
		showClearDestinationAndSyncModal,
		setShowClearDestinationAndSyncModal,
	} = useAppStore()
	const navigate = useNavigate()

	return (
		<Modal
			open={showClearDestinationAndSyncModal}
			footer={null}
			closable={false}
			centered
		>
			<div className="flex w-full flex-col items-center justify-center gap-8">
				<Warning
					className="size-16 text-primary"
					weight="fill"
				/>

				<div className="text-center text-xl font-medium text-gray-950">
					Clear destination and sync deletes all the data in your destination
					and sync your job
				</div>

				<div className="flex w-full justify-end gap-4">
					<Button
						type="primary"
						className="bg-primary text-white"
						onClick={() => {
							setShowClearDestinationAndSyncModal(false)
							message.success("Destination cleared and sync initiated")
							navigate("/jobs")
						}}
					>
						Clear destination and sync
					</Button>
					<Button
						type="default"
						onClick={() => setShowClearDestinationAndSyncModal(false)}
					>
						Cancel
					</Button>
				</div>
			</div>
		</Modal>
	)
}

export default ClearDestinationAndSyncModal

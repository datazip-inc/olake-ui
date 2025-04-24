import { Button, message, Modal } from "antd"
import { useAppStore } from "../../../store"
import { Warning } from "@phosphor-icons/react"
import { useNavigate } from "react-router-dom"

const DeleteJobModal = ({
	fromJobSettings = false,
}: {
	fromJobSettings?: boolean
}) => {
	const {
		showDeleteJobModal,
		setShowDeleteJobModal,
		deleteJob,
		selectedJobId,
	} = useAppStore()
	const navigate = useNavigate()

	return (
		<Modal
			open={showDeleteJobModal}
			footer={null}
			closable={false}
			centered
		>
			<div className="flex w-full flex-col items-center justify-center gap-8">
				<Warning
					className="size-16 text-[#F5222D]"
					weight="fill"
				/>

				<div className="text-center text-xl font-medium text-[#2B2B2B]">
					Are you sure you want to delete this job?
				</div>

				<div className="flex w-full justify-end gap-4">
					<Button
						type="primary"
						danger
						onClick={() => {
							console.log("clicked")
							setShowDeleteJobModal(false)
							if (selectedJobId) {
								deleteJob(selectedJobId).catch(error => {
									message.error("Failed to delete job")
									console.error(error)
								})
							}
							if (fromJobSettings) {
								navigate("/jobs")
							}
						}}
					>
						Delete
					</Button>
					<Button
						type="default"
						onClick={() => setShowDeleteJobModal(false)}
					>
						Cancel
					</Button>
				</div>
			</div>
		</Modal>
	)
}

export default DeleteJobModal

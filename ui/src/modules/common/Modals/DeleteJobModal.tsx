import { useNavigate } from "react-router-dom"
import { Button, message, Modal } from "antd"
import { WarningIcon } from "@phosphor-icons/react"

import { useAppStore } from "../../../store"

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
				<WarningIcon
					className="size-16 text-danger"
					weight="fill"
				/>

				<div className="text-center text-xl font-medium text-gray-950">
					Are you sure you want to delete this job?
				</div>

				<div className="flex w-full justify-end gap-4">
					<Button
						type="primary"
						danger
						onClick={() => {
							setShowDeleteJobModal(false)
							if (selectedJobId) {
								deleteJob(selectedJobId).catch(error => {
									message.error("Failed to delete job")
									console.error(error)
								})
							}
							if (fromJobSettings) {
								setTimeout(() => {
									navigate("/jobs")
								}, 500)
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

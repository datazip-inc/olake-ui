import { WarningIcon } from "@phosphor-icons/react"
import { Button, Modal } from "antd"
import { useNavigate } from "react-router-dom"

import { useJobStore } from "@/modules/ingestion/features/jobs/stores"

import { useDeleteJob } from "../../hooks"

const DeleteJobModal = ({
	fromJobSettings = false,
}: {
	fromJobSettings?: boolean
}) => {
	const { showDeleteJobModal, setShowDeleteJobModal, selectedJobId } =
		useJobStore()
	const { mutateAsync: deleteJob } = useDeleteJob()
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
						onClick={async () => {
							setShowDeleteJobModal(false)
							if (selectedJobId) {
								deleteJob(parseInt(selectedJobId, 10))
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

import { WarningIcon } from "@phosphor-icons/react"
import { Button, message, Modal } from "antd"
import { useNavigate } from "react-router-dom"

import { useJobStore } from "@/modules/ingestion/features/jobs/stores"

import { useClearDestination } from "../../hooks"

const ClearDestinationModal = () => {
	const {
		showClearDestinationModal,
		setShowClearDestinationModal,
		selectedJobId,
	} = useJobStore()
	const navigate = useNavigate()
	const { mutate: clearDestination, isPending } = useClearDestination()

	const handleClearDestination = () => {
		if (!selectedJobId) {
			message.error("No job selected")
			return
		}
		clearDestination(selectedJobId, {
			onSuccess: () => {
				setShowClearDestinationModal(false)
				navigate(`/jobs/${selectedJobId}/history`)
			},
			onError: (error: Error) => {
				message.error(`Failed to clear destination: ${error.message}`)
				setShowClearDestinationModal(false)
			},
		})
	}

	return (
		<Modal
			open={showClearDestinationModal}
			footer={null}
			closable={false}
			centered
		>
			<div className="flex w-full flex-col items-center justify-center gap-8">
				<WarningIcon
					className="size-16 text-danger"
					weight="fill"
				/>

				<div className="text-center text-lg font-normal text-gray-950">
					This will erase all data that was synced by this job in the
					destination. This action{" "}
					<span className="font-bold">cannot be undone</span>. Are you sure you
					want to proceed?
				</div>

				<div className="flex w-full justify-end gap-4">
					<Button
						type="primary"
						danger
						onClick={handleClearDestination}
						loading={isPending}
					>
						Confirm
					</Button>
					<Button
						type="default"
						onClick={() => setShowClearDestinationModal(false)}
					>
						Cancel
					</Button>
				</div>
			</div>
		</Modal>
	)
}

export default ClearDestinationModal

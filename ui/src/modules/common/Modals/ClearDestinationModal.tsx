import { useNavigate } from "react-router-dom"
import { WarningIcon } from "@phosphor-icons/react"
import { Button, message, Modal } from "antd"
import { useAppStore } from "../../../store"
import { jobService } from "../../../api"
import { useState } from "react"

const ClearDestinationModal = () => {
	const {
		showClearDestinationModal,
		setShowClearDestinationModal,
		selectedJobId,
		fetchJobs,
	} = useAppStore()
	const navigate = useNavigate()

	const [isLoading, setIsLoading] = useState(false)

	const handleClearDestination = async () => {
		if (!selectedJobId) {
			message.error("No job selected")
			return
		}
		setIsLoading(true)
		try {
			const response = await jobService.clearDestination(selectedJobId)
			// wait for 1 second before refreshing jobs to avoid fetching old state
			await new Promise(resolve => setTimeout(resolve, 1000))
			await fetchJobs()
			message.success(response.data.message)
			navigate(`/jobs/${selectedJobId}/history`)
		} catch (error) {
			message.destroy()
			message.error(error as string)
			console.error("Failed to clear destination", error)
		} finally {
			setShowClearDestinationModal(false)
			setIsLoading(false)
		}
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
						loading={isLoading}
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

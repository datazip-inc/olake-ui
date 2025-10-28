import { InfoIcon } from "@phosphor-icons/react"
import { Button, Modal } from "antd"
import { useAppStore } from "../../../store"
import { useNavigate } from "react-router-dom"
import { StreamEditDisabledModalProps } from "../../../types/modalTypes"

const StreamEditDisabledModal = ({ from }: StreamEditDisabledModalProps) => {
	const navigate = useNavigate()
	const { showStreamEditDisabledModal, setShowStreamEditDisabledModal } =
		useAppStore()

	const handleCloseModal = () => {
		navigate("/jobs")
		setShowStreamEditDisabledModal(false)
	}

	return (
		<>
			<Modal
				title={
					<div className="flex justify-center">
						<InfoIcon
							weight="fill"
							className="size-12 text-blue-600"
						/>
					</div>
				}
				open={showStreamEditDisabledModal}
				closable={false}
				footer={
					<div className="mt-6 flex justify-center">
						<Button
							key="close"
							type="primary"
							onClick={handleCloseModal}
							className="bg-primary hover:bg-primary-600"
						>
							Go back to Jobs
						</Button>
					</div>
				}
				centered
				width="30%"
			>
				<div className="mt-4 text-center">
					<h3 className="text-xl font-medium"> Editing Disabled</h3>
					<p className="mt-4 text-sm text-gray-700">
						{from === "jobSettings" ? "Job Settings Edit" : "Stream editing"} is
						disabled while the destination is being cleared. It will be
						available once the process finishes.
					</p>
				</div>
			</Modal>
		</>
	)
}

export default StreamEditDisabledModal

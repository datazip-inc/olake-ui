import { Modal } from "antd"
import ErrorIcon from "../../../assets/ErrorIcon.svg"
import { useAppStore } from "../../../store"
import { Info } from "@phosphor-icons/react"
import { useNavigate, useLocation } from "react-router-dom"

const TestConnectionFailureModal = () => {
	const {
		showFailureModal,
		setShowFailureModal,
		sourceTestConnectionError,
		destinationTestConnectionError,
	} = useAppStore()
	const navigate = useNavigate()
	const location = useLocation()

	const isFromSources = location.pathname.includes("/sources")

	const handleCancel = () => {
		setShowFailureModal(false)
	}

	const handleBackToPath = () => {
		setShowFailureModal(false)
		if (isFromSources) {
			navigate("/sources")
		} else {
			navigate("/destinations")
		}
	}

	return (
		<Modal
			open={showFailureModal}
			footer={null}
			closable={false}
			centered
			width={400}
		>
			<div className="flex flex-col items-center justify-center gap-7 py-6">
				<div className="relative">
					<div>
						<img
							src={ErrorIcon}
							alt="Error"
						/>
					</div>
				</div>
				<div className="flex flex-col items-center">
					<p className="text-sm text-[#8A8A8A]">Failed</p>
					<h2 className="text-center text-lg font-medium">
						Your test connection has failed
					</h2>
					<div className="mt-2 flex w-[360px] items-center gap-1 overflow-scroll rounded-xl bg-[#f8f8f8] p-3 text-xs">
						<Info
							weight="bold"
							className="size-4 text-[#f5222d]"
						/>
						{sourceTestConnectionError || destinationTestConnectionError
							? sourceTestConnectionError || destinationTestConnectionError
							: "Error : Please check your parameters"}
					</div>
				</div>
				<div className="flex items-center gap-4">
					<button
						onClick={handleBackToPath}
						className="w-fit rounded-md border border-[#d9d9d9] px-4 py-2 text-black"
					>
						{isFromSources ? "Back to Sources" : "Back to Destinations"}
					</button>
					<button
						onClick={handleCancel}
						className="w-fit flex-1 rounded-md border border-[#f5222d] px-4 py-2 text-[#f5222d]"
					>
						Cancel
					</button>
				</div>
			</div>
		</Modal>
	)
}

export default TestConnectionFailureModal

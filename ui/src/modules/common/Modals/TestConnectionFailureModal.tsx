import { useNavigate } from "react-router-dom"
import { Modal } from "antd"
import { InfoIcon } from "@phosphor-icons/react"

import { useAppStore } from "../../../store"
import ErrorIcon from "../../../assets/ErrorIcon.svg"

const TestConnectionFailureModal = ({
	fromSources,
}: {
	fromSources: boolean
}) => {
	const {
		showFailureModal,
		setShowFailureModal,
		sourceTestConnectionError,
		destinationTestConnectionError,
	} = useAppStore()
	const navigate = useNavigate()

	const handleCancel = () => {
		setShowFailureModal(false)
	}

	const handleBackToPath = () => {
		setShowFailureModal(false)
		if (fromSources) {
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
			width={420}
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
					<p className="text-sm text-text-tertiary">Failed</p>
					<h2 className="text-center text-lg font-medium">
						Your test connection has failed
					</h2>
					<div className="mt-2 flex w-[360px] items-center gap-1 overflow-scroll rounded-xl bg-gray-50 p-3 text-xs">
						<InfoIcon
							weight="bold"
							className="size-4 flex-shrink-0 text-danger"
						/>
						<span className="break-words">
							{fromSources
								? sourceTestConnectionError
								: destinationTestConnectionError}
						</span>
					</div>
				</div>
				<div className="flex items-center gap-4">
					<button
						onClick={handleBackToPath}
						className="w-fit rounded-md border border-[#d9d9d9] px-4 py-2 text-black"
					>
						{fromSources ? "Back to Sources Page" : "Back to Destinations Page"}
					</button>
					<button
						onClick={handleCancel}
						className="w-fit flex-1 rounded-md border border-danger px-4 py-2 text-danger"
					>
						{fromSources ? "Edit  Source" : "Edit  Destination"}
					</button>
				</div>
			</div>
		</Modal>
	)
}

export default TestConnectionFailureModal

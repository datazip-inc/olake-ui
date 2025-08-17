import { Modal } from "antd"
import { useAppStore } from "../../../store"
import TestConnectionGif from "../../../assets/TestConnectionGif.gif"

const TestConnectionModal = () => {
	const { showTestingModal } = useAppStore()

	return (
		<Modal
			open={showTestingModal}
			footer={null}
			closable={false}
			centered
			width={400}
		>
			<div className="flex flex-col items-center justify-center gap-1 py-8">
				<img
					src={TestConnectionGif}
					className="max-w-[70%]"
				/>
				<div className="flex flex-col items-center">
					<p className="text-text-tertiary">Please wait...</p>
					<div className="text-xl font-medium text-gray-950">
						Testing your connection
					</div>
				</div>
			</div>
		</Modal>
	)
}

export default TestConnectionModal

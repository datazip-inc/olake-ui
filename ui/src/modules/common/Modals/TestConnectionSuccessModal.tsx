import { Modal } from "antd"
import DestinationSuccess from "@assets/DestinationSuccess.svg"
import { useAppStore } from "@store/index"

const TestConnectionSuccessModal = () => {
	const { showSuccessModal } = useAppStore()
	return (
		<Modal
			open={showSuccessModal}
			footer={null}
			closable={false}
			centered
			width={400}
		>
			<div className="flex flex-col items-center justify-center gap-7 py-6">
				<img src={DestinationSuccess} />{" "}
				<div className="flex flex-col items-center">
					<p className="text-xs text-text-tertiary">Successful</p>
					<h2 className="text-lg font-medium">
						Your test connection is successful
					</h2>
				</div>
			</div>
		</Modal>
	)
}

export default TestConnectionSuccessModal

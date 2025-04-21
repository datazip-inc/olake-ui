import { Modal } from "antd"
import TestConnection from "../../../assets/TestConnection.svg"
import { useAppStore } from "../../../store"

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
			<div className="flex flex-col items-center justify-center gap-7 py-8">
				<img src={TestConnection} />
				<div className="flex flex-col items-center">
					<p className="text-[#8A8A8A]">Please wait...</p>
					<div className="text-xl font-medium text-[#2B2B2B]">
						Testing your connection
					</div>
				</div>
			</div>
		</Modal>
	)
}

export default TestConnectionModal

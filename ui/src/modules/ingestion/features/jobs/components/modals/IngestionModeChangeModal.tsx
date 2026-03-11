import { Button, Modal } from "antd"
import { useJobStore } from "@/modules/ingestion/features/jobs/stores"
import { IngestionModeChangeModalProps } from "@/modules/ingestion/features/jobs/types"

const IngestionModeChangeModal = ({
	onConfirm,
	ingestionMode,
}: IngestionModeChangeModalProps) => {
	const { showIngestionModeChangeModal, setShowIngestionModeChangeModal } =
		useJobStore()

	return (
		<Modal
			open={showIngestionModeChangeModal}
			footer={null}
			closable={false}
			width={400}
			centered
		>
			<div className="flex w-full flex-col justify-center">
				<div className="text-xl font-medium text-black">
					Switch to {ingestionMode} for all tables ?
				</div>

				<div className="mt-2 text-black">
					<div className="text-sm">
						All tables will be switched to {ingestionMode} mode,
					</div>
					<div className="text-sm">
						You can change mode for individual tables
					</div>
				</div>

				<div className="mt-7 flex w-full gap-4">
					<Button
						type="primary"
						onClick={() => {
							setShowIngestionModeChangeModal(false)
							onConfirm(ingestionMode)
						}}
					>
						Confirm
					</Button>
					<Button
						type="default"
						onClick={() => setShowIngestionModeChangeModal(false)}
					>
						Close
					</Button>
				</div>
			</div>
		</Modal>
	)
}

export default IngestionModeChangeModal

import { WarningIcon } from "@phosphor-icons/react"
import { Modal } from "antd"

interface IndexBuildWarningModalProps {
	open: boolean
	confirmLoading?: boolean
	onConfirm: () => void
	onCancel: () => void
}

// Shown when a job that has no destination index yet picks up a stream whose
// delete format needs one.
const IndexBuildWarningModal = ({
	open,
	confirmLoading = false,
	onConfirm,
	onCancel,
}: IndexBuildWarningModalProps) => (
	<Modal
		open={open}
		width={520}
		okText="Confirm"
		cancelText="Cancel"
		onOk={onConfirm}
		onCancel={onCancel}
		confirmLoading={confirmLoading}
		classNames={{
			footer: "!mt-6 border-t border-neutral-border !pt-4",
		}}
		centered
	>
		<div className="flex items-center gap-4 pr-6 pt-1">
			<div className="flex size-11 shrink-0 items-center justify-center rounded-xl bg-warning-light">
				<WarningIcon
					className="size-6 text-warning"
					weight="fill"
				/>
			</div>

			<div>
				<div className="text-lg font-semibold leading-7 text-gray-950">
					Please Confirm
				</div>
				<p className="mt-2 text-sm leading-6 text-neutral-text">
					Indexes might get built for streams using position deletes or deletion
					vectors. If so, the next sync may take some time.
				</p>
			</div>
		</div>
	</Modal>
)

export default IndexBuildWarningModal

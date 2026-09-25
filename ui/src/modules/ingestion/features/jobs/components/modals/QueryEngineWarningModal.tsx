import { WarningIcon } from "@phosphor-icons/react"
import { Modal } from "antd"

interface QueryEngineWarningModalProps {
	open: boolean
	onConfirm: () => void
	onCancel: () => void
}

// Shown before discovery re-runs with a changed target query engine selection.
const QueryEngineWarningModal = ({
	open,
	onConfirm,
	onCancel,
}: QueryEngineWarningModalProps) => (
	<Modal
		open={open}
		width={560}
		okText="Proceed"
		cancelText="Cancel"
		onOk={onConfirm}
		onCancel={onCancel}
		classNames={{ footer: "mt-6 border-t border-neutral-border pt-4" }}
		centered
	>
		<div className="flex gap-4">
			<div className="flex size-12 shrink-0 items-center justify-center rounded-xl bg-warning-light">
				<WarningIcon
					className="size-6 text-warning"
					weight="fill"
				/>
			</div>

			<div>
				<div className="text-lg font-semibold text-gray-950">
					Target query engines changed
				</div>
				<p className="mt-2 text-neutral-text">
					The target query engine selection has changed, which may change the
					resolved delete format for your streams.
				</p>
				<div className="mt-4 flex gap-3 rounded-lg border border-danger/30 bg-danger-light p-4 text-danger-dark">
					<WarningIcon className="mt-0.5 size-5 shrink-0" />
					<p>
						Some delete format changes require{" "}
						<span className="font-semibold">Clear Destination</span> to run on
						all of the selected streams.{" "}
						<a
							href="https://olake.io/docs/understanding/terminologies/olake/#upsert"
							target="_blank"
							rel="noopener noreferrer"
							className="font-semibold text-danger-dark underline underline-offset-2 hover:text-danger-dark hover:underline hover:opacity-80"
						>
							Learn more
						</a>
					</p>
				</div>
			</div>
		</div>
	</Modal>
)

export default QueryEngineWarningModal

import { WarningIcon } from "@phosphor-icons/react"
import { Modal } from "antd"

interface NoCommonUpdateTypeModalProps {
	open: boolean
	onChangeQueryEngines: () => void
	onContinue: () => void
}

// Shown after discovery when no delete format suits both the target query engines
// and the destination, so every stream was switched to append mode.
const NoCommonUpdateTypeModal = ({
	open,
	onChangeQueryEngines,
	onContinue,
}: NoCommonUpdateTypeModalProps) => (
	<Modal
		open={open}
		width={560}
		okText="Change query engines"
		cancelText="Continue in append mode"
		onOk={onChangeQueryEngines}
		onCancel={onContinue}
		closable={false}
		maskClosable={false}
		keyboard={false}
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
					No common upsert type
				</div>
				<p className="mt-2 text-neutral-text">
					Your target query engine selection and catalog support no common
					upsert type, so all streams will be switched to{" "}
					<span className="font-semibold">Append</span> mode.
				</p>
				<p className="mt-2 text-neutral-text">
					Change the query engines to enable upsert, or continue with append
					only.{" "}
					<a
						href="https://olake.io/docs/understanding/terminologies/olake/#upsert"
						target="_blank"
						rel="noopener noreferrer"
						className="font-semibold text-primary underline underline-offset-2 hover:opacity-80"
					>
						Learn more
					</a>
				</p>
			</div>
		</div>
	</Modal>
)

export default NoCommonUpdateTypeModal

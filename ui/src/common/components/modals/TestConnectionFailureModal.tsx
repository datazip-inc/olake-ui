import {
	CopySimpleIcon,
	FolderDashedIcon,
	WarningCircleIcon,
	XCircleIcon,
} from "@phosphor-icons/react"
import { Button, Modal } from "antd"
import clsx from "clsx"

import { TestConnectionError } from "@/common/types"
import {
	copyToClipboard,
	getLogLevelClass,
	getLogTextColor,
} from "@/common/utils"

type ConnectionType = "source" | "destination" | "catalog"

const connectionLabelMap: Record<ConnectionType, string> = {
	source: "source",
	destination: "destination",
	catalog: "catalog",
}

const TestConnectionFailureModal = ({
	open,
	onClose,
	onEdit,
	connectionType = "source",
	testConnectionError,
}: {
	open: boolean
	onClose: () => void
	onEdit?: () => void
	connectionType?: ConnectionType
	testConnectionError: TestConnectionError | null
}) => {
	const entityLabel = connectionLabelMap[connectionType]
	const title =
		connectionType === "catalog"
			? "Failed to add catalog"
			: `Failed to test ${entityLabel} connection`
	const subtitle = "Please check your connection and try again"

	const labelMap = {
		source: "Source",
		destination: "Destination",
		catalog: "Catalog",
	} as const

	const logs = testConnectionError?.logs ?? []
	const fallbackMessage = testConnectionError?.message ?? ""

	const handleEdit = () => {
		onEdit?.()
	}

	const handleCopyLogs = async () => {
		const logsJson = JSON.stringify(logs || [], null, 4)
		await copyToClipboard(logsJson)
	}

	return (
		<Modal
			open={open}
			footer={null}
			closable={false}
			centered
			width={680}
			destroyOnHidden
			styles={{
				content: {
					padding: 0,
					overflow: "hidden",
					borderRadius: 20,
				},
				body: {
					padding: 0,
				},
			}}
		>
			<div className="flex h-[672px] flex-col bg-white">
				<div className="mt-16 flex w-[261px] flex-col items-center gap-3 self-center text-center">
					<div className="grid size-8">
						<FolderDashedIcon
							size={32}
							weight="regular"
							className="col-start-1 row-start-1 text-olake-error"
						/>
						<XCircleIcon
							size={12}
							weight="fill"
							className="col-start-1 row-start-1 place-self-end text-olake-error"
						/>
					</div>
					<p className="whitespace-nowrap text-xl font-medium leading-7 text-olake-text">
						{title}
					</p>
					<p className="whitespace-nowrap text-sm leading-[22px] text-olake-text-secondary">
						{subtitle}
					</p>
				</div>

				<div className="mt-4 h-[373px] w-[573px] self-center overflow-hidden rounded-lg bg-olake-surface-muted">
					<div className="flex h-[73px] items-start justify-between px-4 pb-0 pt-4">
						<div className="flex items-center gap-1">
							<WarningCircleIcon
								size={16}
								weight="fill"
								className="text-olake-error"
							/>
							<span className="text-sm leading-[22px] text-olake-error">
								Error logs
							</span>
						</div>
						<button
							type="button"
							onClick={handleCopyLogs}
							className="flex items-center gap-1 text-xs font-medium leading-5 text-olake-text-secondary"
						>
							<CopySimpleIcon size={16} />
							<span>Copy Error</span>
						</button>
					</div>

					<div className="h-[300px] overflow-auto px-4 pb-4">
						{logs.length > 0 ? (
							<div className="space-y-2 text-xs">
								{logs.map((log, index) => (
									<div
										key={`${log.time}-${index}`}
										className="flex items-start gap-2"
									>
										<span
											className={clsx(
												"rounded-md px-2 py-[2px] font-mono text-[10px] capitalize leading-4",
												getLogLevelClass(log.level),
											)}
										>
											{log.level}
										</span>
										<p
											className={clsx(
												"whitespace-pre-wrap font-mono text-xs leading-4",
												getLogTextColor(log.level),
											)}
										>
											{log.message}
										</p>
									</div>
								))}
							</div>
						) : (
							<pre className="whitespace-pre-wrap font-mono text-xs leading-4 text-olake-text-secondary">
								{fallbackMessage || "No error logs available."}
							</pre>
						)}
					</div>
				</div>

				<div className="mt-8 flex w-[573px] items-center gap-2 self-center">
					<Button
						type="primary"
						onClick={handleEdit}
					>
						Edit {labelMap[connectionType]}
					</Button>
					<Button onClick={onClose}>Cancel</Button>
				</div>
			</div>
		</Modal>
	)
}

export default TestConnectionFailureModal

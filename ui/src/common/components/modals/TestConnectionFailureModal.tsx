import {
	ArrowSquareOutIcon,
	CopySimpleIcon,
	FolderDashedIcon,
	InfoIcon,
	WarningCircleIcon,
	XCircleIcon,
} from "@phosphor-icons/react"
import { Button, Modal } from "antd"
import clsx from "clsx"

import { POSTGRES_CDC_DOCS } from "@/common/constants"
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

// Driver errors for a Postgres CDC config whose replication slot does not exist.
// The fix is a setup step in Postgres, so the modal points to the guide.
const MISSING_SLOT_PATTERNS = [
	/failed to validate cdc configuration for slot (\S+?): no record found/i,
	/replication slot '([^']+)' does not exist/i,
]

const getMissingReplicationSlot = (texts: string[]) => {
	for (const text of texts) {
		for (const pattern of MISSING_SLOT_PATTERNS) {
			const slot = text.match(pattern)?.[1]
			if (slot) return slot
		}
	}
	return null
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
	const missingSlot = getMissingReplicationSlot([
		fallbackMessage,
		...logs.map(log => log.message),
	])

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

				<div className="mt-4 flex h-[373px] w-[573px] flex-col self-center overflow-hidden rounded-lg bg-olake-surface-muted">
					<div className="flex h-[73px] shrink-0 items-start justify-between px-4 pb-0 pt-4">
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

					{missingSlot && (
						<div className="mx-4 mb-3 flex shrink-0 items-start gap-2 rounded-md border border-primary-100 bg-primary-50 px-3 py-2 text-xs leading-5 text-olake-text">
							<InfoIcon
								size={16}
								weight="fill"
								className="mt-0.5 shrink-0 text-olake-primary"
							/>
							<p>
								Replication slot{" "}
								<span className="font-mono font-medium">{`"${missingSlot}"`}</span>{" "}
								was not found in Postgres. Create the replication slot and a
								publication in Postgres, then enter the same names in the source
								config.{" "}
								<a
									href={POSTGRES_CDC_DOCS.REPLICATION_SLOT}
									target="_blank"
									rel="noopener noreferrer"
									className="inline-flex items-center gap-0.5 font-medium text-primary hover:text-primary/80"
								>
									Setup guide
									<ArrowSquareOutIcon size={12} />
								</a>
							</p>
						</div>
					)}

					<div className="min-h-0 flex-1 overflow-auto px-4 pb-4">
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

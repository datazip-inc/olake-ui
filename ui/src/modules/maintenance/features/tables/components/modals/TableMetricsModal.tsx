import { XIcon } from "@phosphor-icons/react"
import { Modal, Spin } from "antd"

import { formatUtcTimestamp } from "@/common/utils"

import { useTableMetrics } from "../../hooks"

type TableMetricsModalProps = {
	open: boolean
	onClose: () => void
	catalog: string
	database: string
	tableName: string
}

const TableMetricsModal: React.FC<TableMetricsModalProps> = ({
	open,
	onClose,
	catalog,
	database,
	tableName,
}) => {
	const { data: metrics, isLoading } = useTableMetrics(
		catalog,
		database,
		tableName,
		open,
	)

	const tableMetrics = metrics?.["table-metrics"]
	const fileCount = tableMetrics?.["file-count"]

	return (
		<Modal
			open={open}
			onCancel={onClose}
			footer={null}
			centered
			width={560}
			destroyOnHidden
			closeIcon={
				<XIcon
					size={24}
					className="text-olake-text-tertiary"
				/>
			}
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
			<div className="bg-white">
				<div className="px-8 pb-5 pt-10">
					<h2 className="text-2xl font-medium leading-[32px] text-olake-text">
						Metrics
					</h2>
				</div>

				<div className="border-t border-olake-border px-8 pb-8 pt-2">
					{isLoading ? (
						<div className="flex h-40 items-center justify-center">
							<Spin />
						</div>
					) : (
						<div className="mt-5">
							<div className="flex min-h-14 items-center justify-between">
								<p className="pr-8 text-base font-normal leading-[24px] text-olake-text">
									File Count
								</p>
								<p className="text-xl font-medium leading-[28px] text-olake-text">
									{fileCount?.total ?? "--"}
								</p>
							</div>

							<div className="flex min-h-14 items-center justify-between border-t border-olake-border">
								<p className="pr-8 text-base font-normal leading-[24px] text-olake-text">
									Average File Size
								</p>
								<p className="text-xl font-medium leading-[28px] text-olake-text">
									{tableMetrics?.["average-file-size"] || "--"}
								</p>
							</div>

							<div className="flex min-h-14 items-center justify-between border-t border-olake-border">
								<p className="pr-8 text-base font-normal leading-[24px] text-olake-text">
									Last Commit Time
								</p>
								<p className="text-xl font-medium leading-[28px] text-olake-text">
									{tableMetrics
										? formatUtcTimestamp(tableMetrics["last-commit-time"])
										: "--"}
								</p>
							</div>

							<div className="flex min-h-14 items-center justify-between border-t border-olake-border">
								<p className="pr-8 text-base font-normal leading-[24px] text-olake-text">
									Data Files
								</p>
								<span className="inline-flex h-8 items-center rounded-md border border-olake-border px-3 text-sm leading-[22px] text-olake-text-secondary">
									{fileCount?.["data-files"] ?? "--"}
								</span>
							</div>

							<div className="flex min-h-14 items-center justify-between border-t border-olake-border">
								<p className="pr-8 text-base font-normal leading-[24px] text-olake-text">
									Delete Files
								</p>
								<span className="inline-flex h-8 items-center rounded-md border border-olake-border px-3 text-sm leading-[22px] text-olake-text-secondary">
									{fileCount?.["delete-files"] ?? "--"}
								</span>
							</div>
						</div>
					)}
				</div>
			</div>
		</Modal>
	)
}

export default TableMetricsModal

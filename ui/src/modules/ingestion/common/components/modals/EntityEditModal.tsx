import { formatDistanceToNow } from "date-fns"
import { Button, Modal, Table } from "antd"
import { InfoIcon, WarningIcon } from "@phosphor-icons/react"

import { EntityEditModalProps } from "@/modules/ingestion/common/types"
import { getConnectorImage } from "@/modules/ingestion/common/utils"

const EntityEditModal = ({
	entityType,
	open,
	jobs,
	onConfirm,
	onCancel,
}: EntityEditModalProps) => {
	const isSource = entityType === "source"

	const columns = [
		{
			title: "Name",
			dataIndex: "name",
			key: "name",
		},
		{
			title: "Status",
			dataIndex: "activate",
			key: "activate",
			render: (activate: boolean) => (
				<span
					className={`rounded px-2 py-1 text-xs ${
						!activate
							? "bg-danger-light text-danger"
							: "bg-primary-200 text-primary-700"
					}`}
				>
					{activate ? "Active" : "Inactive"}
				</span>
			),
		},
		{
			title: "Last runtime",
			dataIndex: "last_run_time",
			key: "last_run_time",
			render: (text: string) => {
				if (text == null || text === "") return <span>-</span>
				const date = new Date(text)
				if (Number.isNaN(date.getTime())) return <span>-</span>
				return <span>{formatDistanceToNow(date, { addSuffix: true })}</span>
			},
		},
		{
			title: isSource ? "Destination" : "Source",
			dataIndex: isSource ? "destination_name" : "source_name",
			key: isSource ? "destination_name" : "source_name",
			render: (name: string, record: any) => (
				<div className="flex items-center">
					<img
						src={getConnectorImage(
							record[isSource ? "destination_type" : "source_type"] || "",
						)}
						alt={record[isSource ? "destination_type" : "source_type"]}
						className="mr-2 size-6"
					/>
					{name || "N/A"}
				</div>
			),
		},
	]

	return (
		<Modal
			title={
				<div className="flex justify-center">
					<WarningIcon
						weight="fill"
						className="size-12 text-primary"
					/>
				</div>
			}
			open={open}
			onCancel={onCancel}
			footer={[
				<Button
					key="confirm"
					type="primary"
					onClick={onConfirm}
					className="bg-blue-600"
				>
					Confirm
				</Button>,
				<Button
					key="cancel"
					onClick={onCancel}
				>
					Cancel
				</Button>,
			]}
			centered
			width="38%"
		>
			<div className="mt-4 text-center">
				<h3 className="text-lg font-medium">Jobs May Be Affected</h3>
				<p className="mt-2 text-xs text-black text-opacity-45">
					Modifying this {entityType} could affect associated jobs. Are you sure
					you want to continue ?
				</p>
				<div className="mt-2 flex w-full items-center justify-center gap-1 text-xs text-red-600">
					<InfoIcon className="size-4" />
					Ongoing jobs may fail if {entityType} is updated
				</div>
			</div>
			<div className="mt-6">
				<Table
					columns={columns}
					dataSource={jobs}
					pagination={false}
					rowKey="id"
					scroll={{ y: 300 }}
				/>
			</div>
		</Modal>
	)
}

export default EntityEditModal

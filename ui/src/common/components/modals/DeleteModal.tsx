import { formatDistanceToNow } from "date-fns"
import { Button, Modal, Table } from "antd"
import { WarningIcon } from "@phosphor-icons/react"

import { Entity } from "@/common/types"
import { getConnectorImage } from "@/common/utils"

//Entity Delete Modal
const DeleteModal = ({
	open,
	onClose,
	entity,
	fromSource,
	onDelete,
}: {
	open: boolean
	onClose: () => void
	entity: Entity | undefined
	fromSource: boolean
	onDelete: () => void
}) => {
	const handleDelete = () => {
		onDelete()
		onClose()
	}

	const loading = false

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
			render: (text: string) => (
				<div className="flex justify-center">
					{text !== undefined
						? formatDistanceToNow(new Date(text), {
								addSuffix: true,
							})
						: "-"}
				</div>
			),
		},
		...(fromSource
			? [
					{
						title: "Destination",
						dataIndex: "destination_name",
						key: "destination_name",
						render: (destination_name: string, record: any) => (
							<div className="flex items-center">
								<img
									src={getConnectorImage(record.destination_type || "")}
									alt={record.destination_type}
									className="mr-2 size-6"
								/>
								{destination_name || "N/A"}
							</div>
						),
					},
				]
			: [
					{
						title: "Source",
						dataIndex: "source_name",
						key: "source_name",
						render: (source_name: string, record: any) => (
							<div className="flex items-center">
								<img
									src={getConnectorImage(record.source_type || "")}
									alt={record.destination_type}
									className="mr-2 size-6"
								/>
								{source_name || "N/A"}
							</div>
						),
					},
				]),
	]

	const dataSource = entity?.jobs

	return (
		<Modal
			open={open}
			footer={null}
			closable={false}
			centered
			width={600}
		>
			<div className="flex flex-col items-center justify-center gap-7 py-8">
				<WarningIcon
					weight="fill"
					className="h-[55px] w-[63px] text-danger"
				/>
				<div className="flex flex-col items-center">
					<div className="text-center text-xl font-medium text-gray-950">
						Deleting {entity?.name} {fromSource ? "source" : "destination"} will
						disable these <br></br>jobs. Are you sure you want to continue?
					</div>
				</div>

				{dataSource && dataSource?.length > 0 && (
					<Table
						dataSource={dataSource}
						columns={columns}
						rowKey="id"
						loading={loading}
						pagination={false}
						className="w-full rounded-md border"
						rowClassName="no-hover"
						scroll={{ y: 300 }}
					/>
				)}
				<div className="flex w-full justify-end space-x-2">
					<Button
						className="px-4 py-4"
						type="primary"
						danger
						onClick={handleDelete}
					>
						Delete
					</Button>
					<Button
						className="px-4 py-4"
						type="default"
						onClick={onClose}
					>
						Cancel
					</Button>
				</div>
			</div>
		</Modal>
	)
}

export default DeleteModal

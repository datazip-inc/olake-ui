import { useState } from "react"
import { Table, Input, Button, Dropdown } from "antd"
import { Destination } from "../../../types"
import { DotsThree, PencilSimpleLine, TrashSimple } from "@phosphor-icons/react"
import { getConnectorImage } from "../../../utils/utils"

interface DestinationTableProps {
	destinations: Destination[]
	loading: boolean
	onEdit: (id: string) => void
	onDelete: (id: string) => void
}

const DestinationTable: React.FC<DestinationTableProps> = ({
	destinations,
	loading,
	onEdit,
	onDelete,
}) => {
	const [searchText, setSearchText] = useState("")

	const { Search } = Input

	const columns = [
		{
			title: () => <span className="font-medium">Actions</span>,
			key: "actions",
			width: 80,
			render: (_: any, record: Destination) => (
				<Dropdown
					menu={{
						items: [
							{
								key: "edit",
								icon: <PencilSimpleLine />,
								label: "Edit",
								onClick: () => onEdit(record.id),
							},
							{
								key: "delete",
								icon: <TrashSimple />,
								label: "Delete",
								danger: true,
								onClick: () => onDelete(record.id),
							},
						],
					}}
					trigger={["click"]}
					overlayStyle={{ minWidth: "170px" }}
				>
					<Button
						type="text"
						icon={<DotsThree className="size-5" />}
					/>
				</Dropdown>
			),
		},
		{
			title: () => <span className="font-medium">Name</span>,
			dataIndex: "name",
			key: "name",
			render: (text: string) => <div className="flex items-center">{text}</div>,
		},
		{
			title: () => <span className="font-medium">Connectors</span>,
			dataIndex: "type",
			key: "type",
			render: (text: string) => (
				<div className="flex items-center">
					<img
						src={getConnectorImage(text)}
						className="mr-2 h-4 w-4"
					/>
					<span>{text}</span>
				</div>
			),
		},
		{
			title: () => <span className="font-medium">Associated jobs</span>,
			key: "associatedJobs",
			render: () => (
				<div>
					<div className="mb-1 flex items-center">
						<div className="mr-2 flex h-6 w-6 items-center justify-center rounded-full bg-green-500 text-white">
							<span>S</span>
						</div>
						<div className="mx-2 w-16 border-t-2 border-dashed border-gray-300"></div>
						<span className="text-blue-500">Table_name_test_1</span>
						<div className="mx-2 w-16 border-t-2 border-dashed border-gray-300"></div>
						<div className="flex h-6 w-6 items-center justify-center rounded-full bg-red-500 text-white">
							<span>D</span>
						</div>
					</div>
					<div className="ml-8 text-sm text-blue-500">+3 more jobs</div>
				</div>
			),
		},
	]

	const filteredDestinations = destinations.filter(
		destination =>
			destination.name.toLowerCase().includes(searchText.toLowerCase()) ||
			destination.type.toLowerCase().includes(searchText.toLowerCase()),
	)

	return (
		<div>
			<div className="mb-4">
				<Search
					placeholder="Search Destinations"
					allowClear
					className="w-1/4"
					value={searchText}
					onChange={e => setSearchText(e.target.value)}
				/>
			</div>

			<Table
				dataSource={filteredDestinations}
				columns={columns}
				rowKey="id"
				loading={loading}
				pagination={{
					pageSize: 10,
					showSizeChanger: false,
				}}
				className="overflow-hidden rounded-lg"
			/>
		</div>
	)
}

export default DestinationTable

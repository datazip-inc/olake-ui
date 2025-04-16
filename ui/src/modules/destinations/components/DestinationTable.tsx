import { useState } from "react"
import { Table, Input, Button, Dropdown } from "antd"
import { Destination, JobBasic } from "../../../types"
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
						className="mr-2 size-6"
					/>
					<span>{text}</span>
				</div>
			),
		},
		{
			title: () => <span className="font-medium">Associated jobs</span>,
			key: "associatedJobs",
			render: (_: any, record: Destination) => {
				if (!record.associatedJobs || record.associatedJobs.length === 0) {
					return <div className="text-gray-500">No associated jobs</div>
				}
				return (
					<div className="flex-end flex w-fit flex-col items-end gap-3">
						<div className="mb-1 flex items-center">
							{record.associatedJobs.map((job: JobBasic) => (
								<>
									<img
										key={job.source}
										src={getConnectorImage(job.source)}
										className="size-8"
									/>
									<div className="ml-2 text-[#A3A3A3]">-------</div>
									<div className="rounded-[6px] bg-[#E6F4FF] px-2 py-1 text-[#0958D9]">
										{job.jobName}
									</div>
									<div className="mr-2 text-[#A3A3A3]">-------</div>
									<img
										key={job.source}
										src={getConnectorImage(job.destination)}
										className="size-8"
									/>
								</>
							))}
						</div>
						<div className="items-end text-sm font-bold text-[#203FDD]">
							+3 more jobs
						</div>
					</div>
				)
			},
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
					className="custom-search-input w-1/4"
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

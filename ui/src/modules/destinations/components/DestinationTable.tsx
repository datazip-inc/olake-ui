import { useState } from "react"
import { Table, Input, Button, Dropdown, Pagination } from "antd"
import { DotsThree, PencilSimpleLine, Trash } from "@phosphor-icons/react"

import { DestinationTableProps, Entity } from "../../../types"
import { getConnectorImage } from "../../../utils/utils"
import JobConnection from "../../common/components/JobConnection"
import DeleteModal from "../../common/Modals/DeleteModal"

const DestinationTable: React.FC<DestinationTableProps> = ({
	destinations,
	loading,
	onEdit,
	onDelete,
}) => {
	const [searchText, setSearchText] = useState("")
	const [currentPage, setCurrentPage] = useState(1)
	const pageSize = 8

	const { Search } = Input

	const columns = [
		{
			title: () => <span className="font-medium">Actions</span>,
			key: "actions",
			width: 80,
			render: (_: any, record: Entity) => (
				<Dropdown
					menu={{
						items: [
							{
								key: "edit",
								icon: <PencilSimpleLine className="size-4" />,
								label: "Edit",
								onClick: () => onEdit(String(record.id)),
							},
							{
								key: "delete",
								icon: <Trash className="size-4" />,
								label: "Delete",
								danger: true,
								onClick: () => onDelete(record),
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
			title: () => <span className="font-medium">Destination</span>,
			dataIndex: "type",
			key: "type",
			render: (text: string) => (
				<div className="flex items-center">
					<img
						src={getConnectorImage(text)}
						className="mr-2 size-6"
					/>
					<span>{text === "iceberg" ? "Apache Iceberg" : "Amazon S3"}</span>
				</div>
			),
		},
		{
			title: () => <span className="font-medium">Associated jobs</span>,
			key: "jobs",
			dataIndex: "jobs",
			render: (_: any, record: Entity) => {
				const jobs = record.jobs as any[]
				if (jobs.length === 0) {
					return <div className="text-gray-500">No associated jobs</div>
				}

				return (
					<JobConnection
						sourceType={jobs[0].source_type || ""}
						destinationType={record.type}
						jobName={jobs[0].name}
						remainingJobs={jobs.length - 1}
						jobs={jobs}
					/>
				)
			},
		},
	]

	const filteredDestinations = destinations.filter(
		destination =>
			destination.name.toLowerCase().includes(searchText.toLowerCase()) ||
			destination.type.toLowerCase().includes(searchText.toLowerCase()),
	)

	const startIndex = (currentPage - 1) * pageSize
	const endIndex = Math.min(startIndex + pageSize, filteredDestinations.length)
	const currentPageData = filteredDestinations.slice(startIndex, endIndex)

	return (
		<>
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
					dataSource={currentPageData}
					columns={columns}
					rowKey="id"
					loading={loading}
					pagination={false}
					className="overflow-hidden rounded-xl"
					rowClassName="no-hover"
				/>
				<DeleteModal fromSource={false} />
			</div>

			<div className="bottom-15 z-100 fixed right-10 flex justify-end bg-white p-2">
				<Pagination
					current={currentPage}
					onChange={setCurrentPage}
					total={filteredDestinations.length}
					pageSize={pageSize}
					showSizeChanger={false}
				/>
			</div>

			<div className="h-20" />
		</>
	)
}

export default DestinationTable

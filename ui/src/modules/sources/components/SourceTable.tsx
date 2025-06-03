import { useState } from "react"
import { Table, Input, Button, Dropdown, Pagination } from "antd"
import { DotsThree, PencilSimpleLine, Trash } from "@phosphor-icons/react"
import { getConnectorImage, getConnectorLabel } from "../../../utils/utils"
import React from "react"
import DeleteModal from "../../common/Modals/DeleteModal"
import { Entity } from "../../../types"
import { PAGE_SIZE } from "../../../utils/constants"
import JobConnection from "../../common/components/JobConnection"

interface SourceTableProps {
	sources: Entity[]
	loading: boolean
	onEdit: (id: string) => void
	onDelete: (source: Entity) => void
}

const renderJobConnection = (record: Entity) => {
	const jobs = record.jobs as any[]
	if (jobs.length === 0) {
		return <div className="text-gray-500">No associated jobs</div>
	}

	return (
		<JobConnection
			sourceType={record.type}
			destinationType={jobs[0].destination_type || ""}
			jobName={jobs[0].name}
			remainingJobs={jobs.length - 1}
		/>
	)
}

const SourceTable: React.FC<SourceTableProps> = ({
	sources,
	loading,
	onEdit,
	onDelete,
}) => {
	const [searchText, setSearchText] = useState("")
	const [currentPage, setCurrentPage] = useState(1)

	const getTableColumns = () => [
		{
			title: () => <span className="font-medium">Actions</span>,
			key: "actions",
			width: 80,
			render: (_: unknown, record: Entity) => (
				<Dropdown
					menu={{
						items: [
							{
								key: "edit",
								icon: <PencilSimpleLine className="size-4" />,
								label: "Edit",
								onClick: () => onEdit(record.id.toString()),
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
			title: () => <span className="font-medium">Source</span>,
			dataIndex: "type",
			key: "type",
			render: (text: string) => (
				<div className="flex items-center">
					<img
						src={getConnectorImage(text)}
						className="mr-2 size-6"
						alt={`${text} connector`}
					/>
					<span>{getConnectorLabel(text)}</span>
				</div>
			),
		},
		{
			title: () => <span className="font-medium">Associated jobs</span>,
			key: "jobs",
			dataIndex: "jobs",
			render: (_: unknown, record: Entity) => renderJobConnection(record),
		},
	]

	const filteredSources = sources.filter(
		source =>
			source.name.toLowerCase().includes(searchText.toLowerCase()) ||
			source.type.toLowerCase().includes(searchText.toLowerCase()),
	)

	const startIndex = (currentPage - 1) * PAGE_SIZE
	const endIndex = Math.min(startIndex + PAGE_SIZE, filteredSources.length)
	const currentPageData = filteredSources.slice(startIndex, endIndex)

	return (
		<>
			<div>
				<div className="mb-4">
					<Input.Search
						placeholder="Search Sources"
						allowClear
						className="custom-search-input w-1/4"
						value={searchText}
						onChange={e => setSearchText(e.target.value)}
					/>
				</div>

				<Table
					dataSource={currentPageData}
					columns={getTableColumns()}
					rowKey="id"
					loading={loading}
					pagination={false}
					className="overflow-hidden rounded-xl"
					rowClassName="no-hover"
				/>
				<DeleteModal fromSource={true} />
			</div>

			<div className="z-100 fixed bottom-[60px] right-[40px] flex justify-end bg-white p-2">
				<Pagination
					current={currentPage}
					onChange={setCurrentPage}
					total={filteredSources.length}
					pageSize={PAGE_SIZE}
					showSizeChanger={false}
				/>
			</div>

			<div className="h-[80px]" />
		</>
	)
}

export default SourceTable

import React, { useState } from "react"
import { Table, Input, Button, Dropdown, Pagination } from "antd"
import {
	DotsThreeIcon,
	PencilSimpleLineIcon,
	TrashIcon,
} from "@phosphor-icons/react"

import { Entity } from "@/common/types"
import { SourceTableProps } from "../types"
import { getConnectorImage } from "@/common/utils"
import { getConnectorLabel } from "../utils"
import { PAGE_SIZE } from "@/common/constants/constants"
import JobConnection from "@/common/components/JobConnection"
import DeleteModal from "@/common/components/modals/DeleteModal"

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
			jobs={jobs}
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
	const [showDeleteModal, setShowDeleteModal] = useState(false)
	const [deleteEntity, setDeleteEntity] = useState<Entity | null>(null)

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
								icon: <PencilSimpleLineIcon className="size-4" />,
								label: "Edit",
								onClick: () => onEdit(record.id.toString()),
							},
							{
								key: "delete",
								icon: <TrashIcon className="size-4" />,
								label: "Delete",
								danger: true,
								onClick: () => {
									if (!record.jobs || record.jobs.length === 0) {
										onDelete(record)
									} else {
										setDeleteEntity(record)
										setShowDeleteModal(true)
									}
								},
							},
						],
					}}
					trigger={["click"]}
					overlayStyle={{ minWidth: "170px" }}
				>
					<Button
						type="text"
						icon={<DotsThreeIcon className="size-5" />}
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
				<DeleteModal
					open={showDeleteModal}
					onClose={() => setShowDeleteModal(false)}
					entity={deleteEntity ?? undefined}
					fromSource={true}
					onDelete={() => {
						if (deleteEntity) onDelete(deleteEntity)
						setShowDeleteModal(false)
					}}
				/>
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

import { useState } from "react"
import { Table, Input, Button, Dropdown, Pagination, Tooltip } from "antd"
import { Source } from "../../../types"
import { DotsThree, PencilSimpleLine, Trash } from "@phosphor-icons/react"
import { getConnectorImage } from "../../../utils/utils"
import React from "react"
import DeleteModal from "../../common/Modals/DeleteModal"

interface SourceTableProps {
	sources: Source[]
	loading: boolean
	onEdit: (id: string) => void
	onDelete: (source: Source) => void
}

const SourceTable: React.FC<SourceTableProps> = ({
	sources,
	loading,
	onEdit,
	onDelete,
}) => {
	const [searchText, setSearchText] = useState("")
	const [currentPage, setCurrentPage] = useState(1)
	const pageSize = 5
	const { Search } = Input

	const columns = [
		{
			title: () => <span className="font-medium">Actions</span>,
			key: "actions",
			width: 80,

			render: (_: any, record: Source) => (
				<Dropdown
					menu={{
						items: [
							{
								key: "edit",
								icon: <PencilSimpleLine className="size-4" />,
								label: "Edit",
								onClick: () => onEdit(record.id),
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
			render: (_: any, record: Source) => {
				if (!record.associatedJobs || record.associatedJobs.length === 0) {
					return <div className="text-gray-500">No associated jobs</div>
				}
				return (
					<div className="flex-end flex w-fit flex-col items-end gap-3">
						<div className="mb-1 flex items-center">
							<div
								key={`job-${record.associatedJobs[0].jobName}}`}
								className="flex items-center gap-3"
							>
								<div className="flex items-center">
									<img
										src={getConnectorImage(record.associatedJobs[0].source)}
										className="size-8"
									/>
									<div className="ml-2 text-[#A3A3A3]">-------</div>
									<div className="w-36 truncate rounded-[6px] border border-[#D9D9D9] bg-black bg-opacity-[2%] px-2 py-1 text-center text-black">
										{record.associatedJobs[0].jobName.length > 15 ? (
											<Tooltip title={record.associatedJobs[0].jobName}>
												{record.associatedJobs[0].jobName}
											</Tooltip>
										) : (
											record.associatedJobs[0].jobName
										)}
									</div>
									<div className="mr-2 text-[#A3A3A3]">-------</div>
									<img
										src={getConnectorImage(
											record.associatedJobs[0].destination,
										)}
										className="size-8"
									/>
								</div>
								<div className="cursor-pointer items-end text-sm font-bold text-[#203FDD] underline">
									{record.associatedJobs.length > 1
										? `+${record.associatedJobs.length - 1} more jobs`
										: ""}
								</div>
							</div>
						</div>
					</div>
				)
			},
		},
	]

	const filteredSources = sources.filter(
		source =>
			source.name.toLowerCase().includes(searchText.toLowerCase()) ||
			source.type.toLowerCase().includes(searchText.toLowerCase()),
	)

	// Calculate current page data for display
	const startIndex = (currentPage - 1) * pageSize
	const endIndex = Math.min(startIndex + pageSize, filteredSources.length)
	const currentPageData = filteredSources.slice(startIndex, endIndex)

	return (
		<>
			<div>
				<div className="mb-4">
					<Search
						placeholder="Search Sources"
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
				<DeleteModal fromSource={true} />
			</div>

			{/* Fixed pagination at bottom right */}
			<div
				style={{
					position: "fixed",
					bottom: 60,
					right: 40,
					display: "flex",
					justifyContent: "flex-end",
					padding: "8px 0",
					backgroundColor: "#fff",
					zIndex: 100,
				}}
			>
				<Pagination
					current={currentPage}
					onChange={setCurrentPage}
					total={filteredSources.length}
					pageSize={pageSize}
					showSizeChanger={false}
				/>
			</div>

			{/* Add padding at bottom to prevent content from being hidden behind fixed pagination */}
			<div style={{ height: "80px" }}></div>
		</>
	)
}

export default SourceTable

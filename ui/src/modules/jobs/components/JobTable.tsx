import { useState } from "react"
import { Table, Input, Button, Dropdown, Pagination } from "antd"
import { EntityBase, Job } from "../../../types"
import { useNavigate } from "react-router-dom"
import {
	ArrowsClockwise,
	ClockCounterClockwise,
	DotsThree,
	Gear,
	Pause,
	PencilSimple,
	Trash,
} from "@phosphor-icons/react"
import { getConnectorImage } from "../../../utils/utils"
import { getStatusIcon } from "../../../utils/statusIcons"
import { formatDistanceToNow } from "date-fns"

interface JobTableProps {
	jobs: Job[]
	loading: boolean
	onSync: (id: string) => void
	onEdit: (id: string) => void
	onPause: (id: string) => void
	onDelete: (id: string) => void
}

const JobTable: React.FC<JobTableProps> = ({
	jobs,
	loading,
	onSync,
	onEdit,
	onPause,
	onDelete,
}) => {
	const [searchText, setSearchText] = useState("")
	const [currentPage, setCurrentPage] = useState(1)
	const pageSize = 8
	const navigate = useNavigate()

	const { Search } = Input

	const handleViewHistory = (jobId: string) => {
		navigate(`/jobs/${jobId}/history`)
	}

	const handleViewSettings = (jobId: string) => {
		navigate(`/jobs/${jobId}/settings`)
	}

	const columns = [
		{
			title: "Actions",
			key: "actions",
			width: 80,
			render: (_: unknown, record: Job) => (
				<Dropdown
					menu={{
						items: [
							{
								key: "sync",
								icon: <ArrowsClockwise className="size-4" />,
								label: "Sync now",
								onClick: () => onSync(record.id.toString()),
							},
							{
								key: "edit",
								icon: <PencilSimple className="size-4" />,
								label: "Edit",
								onClick: () => onEdit(record.id.toString()),
							},
							{
								key: "pause",
								icon: <Pause className="size-4" />,
								label: "Pause job",
								onClick: () => onPause(record.id.toString()),
							},
							{
								key: "history",
								icon: <ClockCounterClockwise className="size-4" />,
								label: "Job history",
								onClick: () => handleViewHistory(record.id.toString()),
							},
							{
								key: "settings",
								icon: <Gear className="size-4" />,
								label: "Job settings",
								onClick: () => handleViewSettings(record.id.toString()),
							},
							{
								key: "delete",
								icon: <Trash className="size-4" />,
								label: "Delete",
								danger: true,
								onClick: () => onDelete(record.id.toString()),
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
			title: "Job Name",
			dataIndex: "name",
			key: "name",
		},
		{
			title: "Source",
			dataIndex: "source",
			key: "source",
			render: (text: EntityBase) => (
				<div className="flex items-center">
					<img
						src={getConnectorImage(text?.type)}
						className="mr-2 h-4 w-4"
					/>
					{text?.name}
				</div>
			),
		},
		{
			title: "Destination",
			dataIndex: "destination",
			key: "destination",
			render: (text: EntityBase) => (
				<div className="flex items-center">
					<img
						src={getConnectorImage(text?.type)}
						className="mr-2 h-4 w-4"
					/>
					{text?.name}
				</div>
			),
		},
		{
			title: "Last sync",
			dataIndex: "last_run_time",
			key: "last_run_time",
			render: (text?: string) => {
				if (!text) return <div className="pl-4">-</div>
				try {
					const date = new Date(text) // Use native Date instead of parseISO
					if (isNaN(date.getTime())) throw new Error("Invalid date")
					return formatDistanceToNow(date, { addSuffix: true })
				} catch {
					return "-"
				}
			},
		},
		{
			title: "Last sync status",
			dataIndex: "last_run_state",
			key: "last_run_state",
			render: (status: string) => {
				if (!status) return <div className="pl-4">-</div>
				return (
					<div
						className={`flex w-fit items-center justify-center gap-1 rounded-[6px] px-4 py-1 ${
							status === "success"
								? "bg-[#f6ffed] text-[#389E0D]"
								: status === "failed"
									? "bg-[#fff1f0 text-[#cf1322]"
									: ""
						}`}
					>
						{getStatusIcon(status)}
						<span>
							{status === "success"
								? "Success"
								: status === "failed"
									? "Failed"
									: status}
						</span>
					</div>
				)
			},
		},
	]

	const filteredJobs = jobs.filter(
		job =>
			job.name.toLowerCase().includes(searchText.toLowerCase()) ||
			job.source.name.toLowerCase().includes(searchText.toLowerCase()) ||
			job.destination.name.toLowerCase().includes(searchText.toLowerCase()),
	)

	// Calculate current page data for display
	const startIndex = (currentPage - 1) * pageSize
	const endIndex = Math.min(startIndex + pageSize, filteredJobs.length)
	const currentPageData = filteredJobs.slice(startIndex, endIndex)

	return (
		<>
			<div>
				<div className="mb-4">
					<Search
						placeholder="Search Jobs"
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
					total={filteredJobs.length}
					pageSize={pageSize}
					showSizeChanger={false}
				/>
			</div>

			{/* Add padding at bottom to prevent content from being hidden behind fixed pagination */}
			<div style={{ height: "80px" }}></div>
		</>
	)
}

export default JobTable

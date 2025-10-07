import { useState } from "react"
import { useNavigate } from "react-router-dom"
import { formatDistanceToNow } from "date-fns"
import { Table, Input, Button, Dropdown, Pagination } from "antd"
import {
	ArrowsClockwise,
	ClockCounterClockwise,
	DotsThree,
	Gear,
	Pause,
	PencilSimple,
	Play,
	Trash,
	XIcon,
} from "@phosphor-icons/react"

import { EntityBase, Job, JobTableProps } from "../../../types"
import {
	getConnectorImage,
	getStatusClass,
	getStatusLabel,
} from "../../../utils/utils"
import { getStatusIcon } from "../../../utils/statusIcons"
import { PAGE_SIZE } from "../../../utils/constants"

const formatLastSyncTime = (text?: string) => {
	if (!text) return <div className="pl-4">-</div>
	try {
		const date = new Date(text)
		if (isNaN(date.getTime())) throw new Error("Invalid date")
		return formatDistanceToNow(date, { addSuffix: true })
	} catch {
		return "-"
	}
}

const JobTable: React.FC<JobTableProps> = ({
	jobs,
	loading,
	jobType,
	onSync,
	onEdit,
	onPause,
	onDelete,
	onCancelJob,
}) => {
	const [searchText, setSearchText] = useState("")
	const [currentPage, setCurrentPage] = useState(1)
	const navigate = useNavigate()

	const handleViewHistory = (jobId: string) => {
		navigate(`/jobs/${jobId}/history`)
	}

	const handleViewSettings = (jobId: string) => {
		navigate(`/jobs/${jobId}/settings`)
	}

	const getTableColumns = () => [
		{
			title: "Actions",
			key: "actions",
			width: 80,
			render: (_: unknown, record: Job) => {
				const menuItems =
					jobType === "saved"
						? [
								{
									key: "edit",
									icon: <PencilSimple className="size-4" />,
									label: "Edit",
									onClick: () => onEdit(record.id.toString()),
								},
								{
									key: "delete",
									icon: <Trash className="size-4" />,
									label: "Delete",
									danger: true,
									onClick: () => onDelete(record.id.toString()),
								},
							]
						: [
								{
									key: "sync",
									icon: <ArrowsClockwise className="size-4" />,
									label: "Sync now",
									onClick: () => onSync(record.id.toString()),
								},
								{
									key: "edit",
									icon: <PencilSimple className="size-4" />,
									label: "Edit Streams",
									onClick: () => onEdit(record.id.toString()),
								},
								{
									key: "pause",
									icon: record.activate ? (
										<Pause className="size-4" />
									) : (
										<Play className="size-4" />
									),
									label: record.activate ? "Pause job" : "Resume job",
									onClick: () => onPause(record.id.toString(), record.activate),
								},
								{
									key: "cancel",
									icon: <XIcon className="size-4" />,
									label: "Cancel Run",
									disabled: record.last_run_state?.toLowerCase() !== "running",
									onClick: () => onCancelJob(record.id.toString()),
								},
								{
									key: "history",
									icon: <ClockCounterClockwise className="size-4" />,
									label: "Job Logs & History",
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
							]

				return (
					<Dropdown
						menu={{ items: menuItems }}
						trigger={["click"]}
						overlayStyle={{ minWidth: "170px" }}
					>
						<Button
							type="text"
							icon={<DotsThree className="size-5" />}
						/>
					</Dropdown>
				)
			},
		},
		{
			title: "Job ID",
			dataIndex: "id",
			key: "id",
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
					{text.name && (
						<img
							src={getConnectorImage(text.type)}
							className="mr-2 h-5 w-5"
							alt={`${text.name} connector`}
						/>
					)}
					{text.name}
				</div>
			),
		},
		{
			title: "Destination",
			dataIndex: "destination",
			key: "destination",
			render: (text: EntityBase) => (
				<div className="flex items-center">
					{text.name && (
						<img
							src={getConnectorImage(text.type)}
							className="mr-2 h-5 w-5"
							alt={`${text.name} connector`}
						/>
					)}
					{text.name}
				</div>
			),
		},
		{
			title: "Last sync",
			dataIndex: "last_run_time",
			key: "last_run_time",
			render: formatLastSyncTime,
		},
		{
			title: "Last sync status",
			dataIndex: "last_run_state",
			key: "last_run_state",
			render: (status: string) => {
				if (!status) return <div className="pl-4">-</div>
				return (
					<div
						className={`flex w-fit items-center justify-center gap-1 rounded-md px-4 py-1 ${getStatusClass(status)}`}
					>
						{getStatusIcon(status.toLowerCase())}
						<span>{getStatusLabel(status.toLowerCase())}</span>
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

	const startIndex = (currentPage - 1) * PAGE_SIZE
	const endIndex = Math.min(startIndex + PAGE_SIZE, filteredJobs.length)
	const currentPageData = filteredJobs.slice(startIndex, endIndex)

	return (
		<>
			<div>
				<div className="mb-4">
					<Input.Search
						placeholder="Search Jobs"
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
			</div>

			<div className="z-100 fixed bottom-[60px] right-[40px] flex justify-end bg-white p-2">
				<Pagination
					current={currentPage}
					onChange={setCurrentPage}
					total={filteredJobs.length}
					pageSize={PAGE_SIZE}
					showSizeChanger={false}
				/>
			</div>

			<div className="h-[80px]" />
		</>
	)
}

export default JobTable

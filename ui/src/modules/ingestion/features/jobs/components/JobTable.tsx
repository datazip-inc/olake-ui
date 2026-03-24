import {
	ArrowsClockwiseIcon,
	ClockCounterClockwiseIcon,
	DotsThreeIcon,
	GearIcon,
	PauseIcon,
	PencilSimpleIcon,
	PlayIcon,
	TrashIcon,
	XIcon,
} from "@phosphor-icons/react"
import { Table, Input, Button, Dropdown, Pagination, Tooltip } from "antd"
import { formatDistanceToNow } from "date-fns"
import { useState } from "react"
import { useNavigate } from "react-router-dom"

import { getStatusClass, getStatusLabel } from "@/common/utils"
import { getStatusIcon } from "@/modules/ingestion/common/components/statusIcons"
import { PAGE_SIZE } from "@/modules/ingestion/common/constants"
import { getConnectorImage } from "@/modules/ingestion/common/utils"

import { useJobStore } from "../stores"
import { Job, JobTableProps, JobType, SavedJobDraft } from "../types"
import { getJobTypeClass, getJobTypeLabel } from "../utils"

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
	onRefresh,
	onSync,
	onEdit,
	onPause,
	onDelete,
	onCancelJob,
}) => {
	const [searchText, setSearchText] = useState("")
	const [currentPage, setCurrentPage] = useState(1)
	const navigate = useNavigate()
	const { setSelectedJobId } = useJobStore()

	const handleViewHistory = (jobId: string) => {
		navigate(`/jobs/${jobId}/history`)
	}

	const handleViewSettings = (jobId: string) => {
		setSelectedJobId(jobId)
		navigate(`/jobs/${jobId}/settings`)
	}

	const getTableColumns = () => [
		{
			title: "Actions",
			key: "actions",
			width: 100,
			render: (_: unknown, record: Job | SavedJobDraft) => {
				const menuItems =
					jobType === "saved"
						? [
								{
									key: "edit",
									icon: <PencilSimpleIcon className="size-4" />,
									label: "Edit",
									onClick: () => onEdit(record.id.toString()),
								},
								{
									key: "delete",
									icon: <TrashIcon className="size-4" />,
									label: "Delete",
									danger: true,
									onClick: () => onDelete(record.id.toString()),
								},
							]
						: (() => {
								const j = record as Job
								return [
									{
										key: "sync",
										icon: <ArrowsClockwiseIcon className="size-4" />,
										label: "Sync now",
										disabled:
											j.last_run_state?.toLowerCase() === "running" ||
											!j.activate,
										onClick: () => onSync(record.id.toString()),
									},
									{
										key: "edit",
										icon: <PencilSimpleIcon className="size-4" />,
										label: "Edit Streams",
										disabled: !j.activate,
										onClick: () => onEdit(record.id.toString()),
									},
									{
										key: "pause",
										icon: j.activate ? (
											<PauseIcon className="size-4" />
										) : (
											<PlayIcon className="size-4" />
										),
										label: j.activate ? "Pause job" : "Resume job",
										disabled: j.last_run_state?.toLowerCase() === "running",
										onClick: () => onPause(record.id.toString(), j.activate),
									},
									{
										key: "cancel",
										icon: <XIcon className="size-4" />,
										label: "Cancel Run",
										disabled:
											!j.activate ||
											j.last_run_state?.toLowerCase() !== "running" ||
											(j.last_run_type === JobType.ClearDestination &&
												j.last_run_state?.toLowerCase() === "running"),
										onClick: () => onCancelJob(record.id.toString()),
									},
									{
										key: "history",
										icon: <ClockCounterClockwiseIcon className="size-4" />,
										label: "Job Logs & History",
										disabled: !j.activate,
										onClick: () => handleViewHistory(record.id.toString()),
									},
									{
										key: "settings",
										icon: <GearIcon className="size-4" />,
										label: "Job settings",
										disabled: !j.activate,
										onClick: () => handleViewSettings(record.id.toString()),
									},
									{
										key: "delete",
										icon: <TrashIcon className="size-4" />,
										label: "Delete",
										danger: true,
										onClick: () => onDelete(record.id.toString()),
									},
								]
							})()

				return (
					<Dropdown
						menu={{ items: menuItems }}
						trigger={["click"]}
						overlayStyle={{ minWidth: "170px" }}
					>
						<Button
							type="text"
							data-testid={`job-${record.name}`}
							icon={<DotsThreeIcon className="size-5" />}
						/>
					</Dropdown>
				)
			},
		},
		{
			title: "Job ID",
			dataIndex: "id",
			key: "id",
			width: 100,
		},
		{
			title: "Job Name",
			dataIndex: "name",
			key: "name",
			width: 180,
			render: (name: string) => (
				<Tooltip title={name}>
					<div className="truncate">{name}</div>
				</Tooltip>
			),
		},
		{
			title: "Source",
			dataIndex: "source",
			key: "source",
			width: 180,
			render: (text: { name: string; type: string }) => (
				<Tooltip title={text?.name}>
					<div className="flex min-w-0 items-center">
						{text?.name && (
							<img
								src={getConnectorImage(text.type)}
								className="mr-2 h-5 w-5 shrink-0"
								alt={`${text.name} connector`}
							/>
						)}
						<span className="min-w-0 flex-1 truncate">{text?.name}</span>
					</div>
				</Tooltip>
			),
		},
		{
			title: "Destination",
			dataIndex: "destination",
			key: "destination",
			width: 180,
			render: (text: { name: string; type: string }) => (
				<Tooltip title={text?.name}>
					<div className="flex min-w-0 items-center">
						{text?.name && (
							<img
								src={getConnectorImage(text.type)}
								className="mr-2 h-5 w-5 shrink-0"
								alt={`${text.name} connector`}
							/>
						)}
						<span className="min-w-0 flex-1 truncate">{text?.name}</span>
					</div>
				</Tooltip>
			),
		},
		{
			title: "Last Run",
			dataIndex: "last_run_time",
			key: "last_run_time",
			render: formatLastSyncTime,
		},
		{
			title: "Last Run status",
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
		{
			title: "Job Type",
			dataIndex: "last_run_type",
			key: "last_run_type",
			render: (lastRunType: JobType) => {
				if (!lastRunType) return <div className="pl-4">-</div>
				return (
					<div
						className={`flex w-fit items-center justify-center gap-1 rounded-md px-4 py-1 ${getJobTypeClass(lastRunType)}`}
					>
						<span>{getJobTypeLabel(lastRunType)}</span>
					</div>
				)
			},
		},
	]

	const filteredJobs = jobs.filter(
		job =>
			job.name.toLowerCase().includes(searchText.toLowerCase()) ||
			job.source?.name.toLowerCase().includes(searchText.toLowerCase()) ||
			job.destination?.name.toLowerCase().includes(searchText.toLowerCase()),
	)

	const startIndex = (currentPage - 1) * PAGE_SIZE
	const endIndex = Math.min(startIndex + PAGE_SIZE, filteredJobs.length)
	const currentPageData = filteredJobs.slice(startIndex, endIndex)

	return (
		<>
			<div>
				<div className="mb-4 flex gap-2">
					<Input.Search
						placeholder="Search Jobs"
						allowClear
						className="custom-search-input w-1/4"
						value={searchText}
						onChange={e => setSearchText(e.target.value)}
					/>

					<Tooltip title="Click to refetch">
						<Button
							icon={<ArrowsClockwiseIcon size={16} />}
							loading={loading}
							onClick={onRefresh}
							className="flex items-center"
						/>
					</Tooltip>
				</div>

				<div className="overflow-x-auto">
					<Table
						dataSource={currentPageData}
						columns={getTableColumns()}
						rowKey="id"
						loading={loading}
						pagination={false}
						tableLayout="fixed"
						scroll={{ x: 1200 }}
						rowClassName="no-hover"
					/>
				</div>
			</div>

			<div className="fixed bottom-[60px] right-[40px] z-50 flex justify-end bg-white p-2">
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

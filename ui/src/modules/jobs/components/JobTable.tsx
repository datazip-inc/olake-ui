import { useState } from "react"
import { Table, Input, Button, Dropdown } from "antd"
import { Job } from "../../../types"
import { useNavigate } from "react-router-dom"
import {
	ArrowsClockwise,
	ArrowsCounterClockwise,
	CheckCircle,
	ClockCounterClockwise,
	DotsThree,
	Gear,
	Pause,
	PencilSimple,
	Trash,
	XCircle,
} from "@phosphor-icons/react"
import { getConnectorImage } from "../../../utils/utils"

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
	const navigate = useNavigate()

	const { Search } = Input

	const handleViewHistory = (jobId: string) => {
		navigate(`/jobs/${jobId}/history`)
	}

	const handleViewSettings = (jobId: string) => {
		navigate(`/jobs/${jobId}/settings`)
	}

	const getStatusIcon = (status: string | undefined) => {
		if (status === "success") {
			return <CheckCircle className="text-green-500" />
		} else if (status === "failed") {
			return <XCircle className="text-red-500" />
		} else if (status === "running") {
			return <ArrowsCounterClockwise className="text-blue-500" />
		}
		return null
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
								onClick: () => onSync(record.id),
							},
							{
								key: "edit",
								icon: <PencilSimple className="size-4" />,
								label: "Edit",
								onClick: () => onEdit(record.id),
							},
							{
								key: "pause",
								icon: <Pause className="size-4" />,
								label: "Pause job",
								onClick: () => onPause(record.id),
							},
							{
								key: "history",
								icon: <ClockCounterClockwise className="size-4" />,
								label: "Job history",
								onClick: () => handleViewHistory(record.id),
							},
							{
								key: "settings",
								icon: <Gear className="size-4" />,
								label: "Job settings",
								onClick: () => handleViewSettings(record.id),
							},
							{
								key: "delete",
								icon: <Trash className="size-4" />,
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
			title: "Job Name",
			dataIndex: "name",
			key: "name",
		},
		{
			title: "Source",
			dataIndex: "source",
			key: "source",
			render: (text: string) => (
				<div className="flex items-center">
					<img
						src={getConnectorImage(text)}
						className="mr-2 h-4 w-4"
					/>
					{text}
				</div>
			),
		},
		{
			title: "Destination",
			dataIndex: "destination",
			key: "destination",
			render: (text: string) => (
				<div className="flex items-center">
					<img
						src={getConnectorImage(text)}
						className="mr-2 h-4 w-4"
					/>
					{text}
				</div>
			),
		},
		{
			title: "Last sync",
			dataIndex: "lastSync",
			key: "lastSync",
		},
		{
			title: "Last sync status",
			dataIndex: "lastSyncStatus",
			key: "lastSyncStatus",
			render: (status: string) => (
				<div
					className={`flex w-fit items-center justify-center rounded-[6px] px-4 py-1 ${
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
			),
		},
	]

	const filteredJobs = jobs.filter(
		job =>
			job.name.toLowerCase().includes(searchText.toLowerCase()) ||
			job.source.toLowerCase().includes(searchText.toLowerCase()) ||
			job.destination.toLowerCase().includes(searchText.toLowerCase()),
	)

	return (
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
				dataSource={filteredJobs}
				columns={columns}
				rowKey="id"
				loading={loading}
				pagination={{
					pageSize: 10,
					showSizeChanger: false,
				}}
				className="overflow-hidden rounded-xl"
				rowClassName="no-hover"
			/>
		</div>
	)
}

export default JobTable

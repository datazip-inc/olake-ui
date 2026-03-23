import {
	ChartDonutIcon,
	FolderSimpleIcon,
	TableIcon,
	ToggleRightIcon,
} from "@phosphor-icons/react"
import { Button } from "antd"
import { useNavigate } from "react-router-dom"

type StepCardProps = {
	title: string
	description: string
	icon: React.ReactNode
}

const StepCard: React.FC<StepCardProps> = ({ title, description, icon }) => {
	return (
		<div className="flex h-[90px] items-center gap-5 rounded-lg border border-olake-border bg-white px-6">
			<div className="flex size-8 items-center justify-center text-olake-primary">
				{icon}
			</div>
			<div className="flex flex-col gap-1">
				<p className="text-base leading-6 text-olake-heading-strong">{title}</p>
				<p className="text-sm leading-[22px] text-olake-body">{description}</p>
			</div>
		</div>
	)
}

const TableEmptyState: React.FC = () => {
	const navigate = useNavigate()

	return (
		<div className="min-h-full bg-white px-6 py-9">
			<div className="max-w-[1127px]">
				<div className="mb-10 max-w-[530px]">
					<p className="mb-3 text-xl leading-7 text-olake-body">
						Launching Optimization
					</p>
					<p className="mb-3 text-[30px] font-medium leading-[38px] text-olake-heading-strong">
						Optimize your tables for faster queries &amp; less storage
					</p>
				</div>

				<div className="mb-10">
					<h2 className="mb-6 text-xl font-medium leading-7 text-olake-heading-strong">
						Follow these steps to get started
					</h2>
					<div className="space-y-4">
						<StepCard
							title="Add Catalogs"
							description="Create and manage your catalogs efficiently"
							icon={<FolderSimpleIcon size={24} />}
						/>
						<StepCard
							title="View tables"
							description="Choose your catalogs & database to explore the available tables"
							icon={<TableIcon size={24} />}
						/>
						<StepCard
							title="Configure & Run Maintenance"
							description="Set up maintenance"
							icon={<ToggleRightIcon size={24} />}
						/>
						<StepCard
							title="View Logs & Metrics"
							description="Access logs and metrics for detailed insights"
							icon={<ChartDonutIcon size={24} />}
						/>
					</div>
				</div>

				<div>
					<p className="mb-4 text-base leading-6 text-olake-heading-strong">
						What would you like to do next ?
					</p>
					<div className="flex items-center gap-2">
						<Button
							className="h-10 min-w-24 rounded-lg border-olake-border text-base text-olake-heading-strong"
							onClick={() => navigate("/maintenance/catalogs")}
						>
							Add Catalog
						</Button>
					</div>
				</div>
			</div>
		</div>
	)
}

export default TableEmptyState

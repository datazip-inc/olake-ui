import {
	ClockCounterClockwiseIcon,
	DotsThreeIcon,
	SpinnerIcon,
	TrashIcon,
} from "@phosphor-icons/react"
import { Button, Dropdown, Spin, Switch } from "antd"
import type { MenuProps } from "antd/es/menu"

import type { ColumnDef } from "@/common/components"

import { RunStatusCell } from "../components"
import type { Table } from "../types"

const getHealthScoreColor = (score: number) =>
	score > 70 ? "text-olake-success" : "text-olake-warning"

export interface TableActions {
	onViewLogs: (row: Table) => void
	onCancelRun: (row: Table) => void
	onToggleOptimizing: (row: Table, enabled: boolean) => void
	onViewMetrics: (row: Table) => void
	onConfigure: (row: Table) => void
}

export interface TableColumnOptions {
	openActionRow: string | null
	setOpenActionRow: (id: string | null) => void
	isTogglePendingFor: (tableName: string) => boolean
	isCancelPendingFor: (tableName: string) => boolean
	actions: TableActions
}

export function getTableColumns(opts: TableColumnOptions): ColumnDef<Table>[] {
	const {
		openActionRow,
		setOpenActionRow,
		isTogglePendingFor,
		isCancelPendingFor,
		actions,
	} = opts

	const getActionMenuItems = (row: Table): MenuProps["items"] => [
		{
			key: `logs-${row.id}`,
			icon: <ClockCounterClockwiseIcon size={20} />,
			label: (
				<span className="text-sm leading-[22px]">View Logs &amp; Runs</span>
			),
			onClick: () => {
				setOpenActionRow(null)
				actions.onViewLogs(row)
			},
		},
		{
			key: `cancel-${row.id}`,
			icon: isCancelPendingFor(row.name) ? (
				<SpinnerIcon
					size={20}
					className="animate-spin text-olake-text-secondary"
				/>
			) : (
				<TrashIcon size={20} />
			),
			label: <span className="text-sm leading-[22px]">Cancel Run</span>,
			disabled: isCancelPendingFor(row.name),
			onClick: () => {
				setOpenActionRow(null)
				actions.onCancelRun(row)
			},
		},
	]

	return [
		{
			key: "actions",
			header: "Actions",
			width: 8,
			render: row => (
				<Dropdown
					menu={{ items: getActionMenuItems(row) }}
					trigger={["click"]}
					open={openActionRow === row.id}
					onOpenChange={isOpen => setOpenActionRow(isOpen ? row.id : null)}
				>
					<Button className="size-8 border-0 p-0">
						<DotsThreeIcon size={16} />
					</Button>
				</Dropdown>
			),
		},
		{
			key: "name",
			header: "Table",
			width: 24,
			render: row => (
				<div className="flex items-center gap-2">
					<p className="text-sm leading-6 text-olake-text">{row.name}</p>
					{row.byOLake && (
						<span className="inline-flex h-5 items-center rounded-[20px] bg-olake-primary-bg px-2 text-[10px] font-medium leading-5 text-olake-primary">
							OLake
						</span>
					)}
				</div>
			),
		},
		{
			key: "healthScore",
			header: "Health Score",
			width: 12,
			render: row => (
				<span
					className={`text-sm leading-6 ${getHealthScoreColor(row.healthScore)}`}
				>
					{row.healthScore}
				</span>
			),
		},
		{
			key: "lastRunStatus",
			header: "Last Run Status",
			width: 14,
			align: "center",
			render: row => <RunStatusCell row={row} />,
		},
		{
			key: "metrics",
			header: "Metrics",
			width: 12,
			align: "center",
			render: row => (
				<Button
					size="small"
					onClick={() => actions.onViewMetrics(row)}
				>
					View Metrics
				</Button>
			),
		},
		{
			key: "maintenance",
			header: "Maintenance",
			width: 12,
			align: "center",
			render: row => (
				<Button
					size="small"
					onClick={() => actions.onConfigure(row)}
				>
					Configure
				</Button>
			),
		},
		{
			key: "status",
			header: "Status",
			align: "center",
			render: row => {
				if (isTogglePendingFor(row.name)) {
					return (
						<div className="flex h-6 items-center justify-center">
							<Spin size="small" />
						</div>
					)
				}

				return (
					<Switch
						size="small"
						checked={row.enabled}
						onChange={checked => actions.onToggleOptimizing(row, checked)}
					/>
				)
			},
		},
	]
}

import {
	ClockCounterClockwiseIcon,
	DotsThreeIcon,
	QuestionIcon,
	SpinnerIcon,
	TrashIcon,
} from "@phosphor-icons/react"
import { Button, Dropdown, Spin, Switch, Tooltip } from "antd"
import type { MenuProps } from "antd/es/menu"
import clsx from "clsx"

import { Tag, type ColumnDef } from "@/common/components"

import { RunStatusCell } from "../components"
import type { Table } from "../types"
import { getCancelRunID } from "./tableUtils"

const getHealthScoreColor = (score: number) =>
	score > 70 ? "text-olake-success" : "text-olake-warning"

export interface TableActions {
	onViewLogs: (row: Table) => void
	onCancelRun: (row: Table) => void
	onToggleOptimizingStatus: (row: Table, enabled: boolean) => void
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
			disabled: isCancelPendingFor(row.name) || !getCancelRunID(row),
			onClick: () => {
				if (!getCancelRunID(row)) return
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
					{row.olakeCreated && <Tag>OLake</Tag>}
				</div>
			),
		},
		{
			key: "healthScore",
			header: (
				<div className="flex items-center gap-1">
					<span>Health Score</span>
					<Tooltip title="Overall table efficiency based on small files and deletes. Higher is better.">
						<QuestionIcon
							size={14}
							className="cursor-help text-olake-text-secondary"
						/>
					</Tooltip>
				</div>
			),
			width: 12,
			render: row =>
				row.healthScore < 0 ? (
					<Tag
						color="muted"
						className="text-xs"
					>
						Not determined
					</Tag>
				) : (
					<span
						className={clsx(
							"text-sm leading-6",
							getHealthScoreColor(row.healthScore),
						)}
					>
						{row.healthScore}%
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
			header: (
				<div className="flex items-center justify-center gap-1">
					<span>Status</span>
					<Tooltip title="Enables or disables all optimization runs.">
						<QuestionIcon
							size={14}
							className="cursor-help text-olake-text-secondary"
						/>
					</Tooltip>
				</div>
			),
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
						className={row.enabled ? "!bg-olake-success" : undefined}
						onChange={checked => actions.onToggleOptimizingStatus(row, checked)}
					/>
				)
			},
		},
	]
}

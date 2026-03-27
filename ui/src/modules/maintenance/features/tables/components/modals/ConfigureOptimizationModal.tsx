import { CaretUpIcon, QuestionIcon, TableIcon } from "@phosphor-icons/react"
import { Button, Input, Modal, Select, Spin, Tooltip } from "antd"
import clsx from "clsx"
import { useEffect, useState } from "react"

import ConfigurationSuccessModal from "./ConfigurationSuccessModal"
import { CRON_FREQUENCY_OPTIONS, DEFAULT_CRON_CONFIG } from "../../constants"
import { DEFAULT_TABLE_MODAL_STYLES } from "../../constants"
import { useTableDetails, useUpdateTableCronConfig } from "../../hooks"
import type {
	CronConfigOption,
	ScheduleSectionProps,
	UpdateTableCronApiRequest,
} from "../../types"
import { getCronFromConfig, getEarliestNextRun, getNextRuns } from "../../utils"

type ConfigureOptimizationModalProps = {
	open: boolean
	onClose: () => void
	catalog: string
	database: string
	tableName: string
	tableSize: string
}

const ScheduleSection: React.FC<ScheduleSectionProps> = ({
	title,
	value,
	onChange,
	isFirst = false,
	tooltip,
}) => {
	const nextRuns = getNextRuns(getCronFromConfig(value))

	return (
		<div
			className={clsx(
				"flex flex-col gap-4 py-8",
				!isFirst && "border-t border-olake-border-secondary",
			)}
		>
			{/* Section heading */}
			<div className="flex items-center gap-1">
				<span className="text-base font-medium leading-6 text-olake-text">
					{title}
				</span>
				{tooltip && (
					<Tooltip title={tooltip}>
						<QuestionIcon
							size={14}
							className="cursor-help text-olake-text-tertiary"
						/>
					</Tooltip>
				)}
			</div>

			{/* Form row */}
			<div className="flex gap-4">
				{/* Frequency */}
				<div className="flex w-[233px] flex-col gap-2">
					<label className="text-sm leading-[22px] text-olake-text">
						Frequency
					</label>
					<Select
						value={value.frequency}
						onChange={next =>
							onChange({ ...value, frequency: next, customCron: "" })
						}
						options={CRON_FREQUENCY_OPTIONS}
						className="w-full"
					/>
				</div>

				{value.frequency === "custom" && (
					<div className="flex w-[346px] flex-col gap-2">
						<label className="text-sm leading-[22px] text-olake-text">
							Cron Expression
						</label>
						<Input
							value={value.customCron}
							onChange={e => onChange({ ...value, customCron: e.target.value })}
							placeholder="e.g. 0 */6 * * *"
							className="font-mono text-[13px]"
						/>
					</div>
				)}
			</div>

			{/* Next 3 Runs */}
			{nextRuns.length > 0 && (
				<div className="flex flex-col gap-2">
					<span className="text-sm font-medium leading-6 text-olake-text">
						Next 3 Runs
					</span>
					<div className="flex items-center gap-2">
						{nextRuns.map(run => (
							<span
								key={`${title}-${run}`}
								className="w-48 rounded-[6px] bg-olake-surface-muted px-2 py-1 text-sm leading-[22px] text-olake-text-secondary"
							>
								{run}
							</span>
						))}
					</div>
				</div>
			)}
		</div>
	)
}

type ActiveModal = null | "success"

const ConfigureOptimizationModal: React.FC<ConfigureOptimizationModalProps> = ({
	open,
	onClose,
	catalog,
	database,
	tableName,
	tableSize,
}) => {
	const isTableIdentified = !!catalog && !!database && !!tableName
	const [minorCron, setMinorCron] =
		useState<CronConfigOption>(DEFAULT_CRON_CONFIG)
	const [majorCron, setMajorCron] =
		useState<CronConfigOption>(DEFAULT_CRON_CONFIG)
	const [fullCron, setFullCron] =
		useState<CronConfigOption>(DEFAULT_CRON_CONFIG)
	const [targetFileSize, setTargetFileSize] = useState(100)
	const [advancedOpen, setAdvancedOpen] = useState(true)
	const [activeModal, setActiveModal] = useState<ActiveModal>(null)
	const {
		data: tableCronConfig,
		isLoading: isConfigLoading,
		isError: isConfigError,
		refetch: refetchConfig,
	} = useTableDetails(catalog, database, tableName)
	const { mutate: updateTableCronConfig, isPending: isSaveLoading } =
		useUpdateTableCronConfig(catalog, database, tableName)

	useEffect(() => {
		if (!tableCronConfig) return

		const config = tableCronConfig
		setMinorCron(config.minorCron)
		setMajorCron(config.majorCron)
		setFullCron(config.fullCron)
		if (config.targetFileSize !== undefined && config.targetFileSize > 0) {
			setTargetFileSize(config.targetFileSize)
		}
	}, [tableCronConfig])

	useEffect(() => {
		if (!open) {
			setActiveModal(null)
		}
	}, [open])

	const handleSave = () => {
		if (!catalog || !database || !tableName) return

		const payload: UpdateTableCronApiRequest = {
			minorTriggerInterval: getCronFromConfig(minorCron),
			majorTriggerInterval: getCronFromConfig(majorCron),
			fullTriggerInterval: getCronFromConfig(fullCron),
			targetFileSize: targetFileSize,
		}

		updateTableCronConfig(payload, {
			onSuccess: () => {
				setActiveModal("success")
			},
		})
	}

	const handleClose = () => {
		setActiveModal(null)
		onClose()
	}

	const firstRunAt = getEarliestNextRun([minorCron, majorCron, fullCron])
	const isTargetFileSizeValid = targetFileSize > 0

	return (
		<>
			<Modal
				open={open && activeModal === null}
				onCancel={handleClose}
				footer={null}
				closable={false}
				centered
				width={696}
				destroyOnHidden
				styles={DEFAULT_TABLE_MODAL_STYLES}
			>
				<div className="flex h-[808px] flex-col overflow-hidden bg-white">
					{/* Header */}
					<div className="px-8 pt-10">
						<h2 className="text-xl font-medium leading-7 text-olake-text">
							Configure Optimization
						</h2>

						{/* Table chip */}
						<div className="mt-4 flex h-7 items-center justify-between rounded-[4px] bg-olake-surface-muted pl-3 pr-3">
							<div className="flex items-center gap-2">
								<TableIcon
									size={16}
									className="text-olake-text"
								/>
								<span className="text-sm leading-[22px] text-olake-text">
									{tableName}
								</span>
							</div>
							<span className="text-sm leading-[22px] text-olake-text">
								{tableSize}
							</span>
						</div>
					</div>

					{/* Scrollable content */}
					<div className="flex-1 overflow-y-auto px-8">
						{isConfigError ? (
							<div className="flex h-full flex-col items-center justify-center gap-1 text-center">
								<p className="text-xl font-medium leading-7 text-olake-heading-strong">
									Failed to load cron configuration
								</p>
								<p className="text-sm leading-[22px] text-olake-body">
									Unable to fetch the schedule configuration. Please try again.
								</p>
								<Button
									type="primary"
									className="mt-3"
									onClick={() => void refetchConfig()}
								>
									Retry
								</Button>
							</div>
						) : isConfigLoading ? (
							<div className="flex h-full items-center justify-center">
								<Spin size="large" />
							</div>
						) : (
							<>
								<ScheduleSection
									title="Lite"
									tooltip="Converts equality deletes to position deletes."
									value={minorCron}
									onChange={setMinorCron}
									isFirst
								/>
								<ScheduleSection
									title="Medium"
									tooltip="Merges small files into larger ones."
									value={majorCron}
									onChange={setMajorCron}
								/>
								<ScheduleSection
									title="Full"
									tooltip="Rewrites all data and applies deletes, creating files of the target size."
									value={fullCron}
									onChange={setFullCron}
								/>

								{/* Advanced Config */}
								<div className="border-t border-olake-border-secondary">
									<button
										type="button"
										className="flex w-full items-center justify-between py-2"
										onClick={() => setAdvancedOpen(prev => !prev)}
									>
										<span className="text-base font-medium leading-6 text-olake-text">
											Advanced Config
										</span>
										<CaretUpIcon
											size={16}
											className={clsx(
												"text-olake-text-tertiary transition-transform",
												!advancedOpen && "rotate-180",
											)}
										/>
									</button>

									{advancedOpen && (
										<div className="flex items-center justify-between pb-8">
											<div className="flex items-center justify-center gap-x-2">
												<span className="text-base leading-6 text-olake-text">
													Target file size
												</span>
												<Tooltip title="Desired file size after compaction; full optimization creates files of this size.">
													<QuestionIcon
														size={14}
														className="cursor-help text-olake-text-tertiary"
													/>
												</Tooltip>
											</div>
											<Input
												value={
													targetFileSize === 0 ? "" : String(targetFileSize)
												}
												onChange={e => {
													const numericValue = e.target.value.replace(
														/[^0-9]/g,
														"",
													)
													setTargetFileSize(
														numericValue === "" ? 0 : Number(numericValue),
													)
												}}
												inputMode="numeric"
												placeholder="100"
												addonAfter="MB"
												className="w-32"
											/>
										</div>
									)}
								</div>
							</>
						)}
					</div>

					{/* Footer */}
					<div className="flex h-20 shrink-0 items-center gap-3 border-t border-olake-border pl-8">
						<Button
							type="primary"
							onClick={handleSave}
							loading={isSaveLoading}
							disabled={
								!isTableIdentified ||
								isConfigLoading ||
								isConfigError ||
								!isTargetFileSizeValid
							}
						>
							Save
						</Button>
						<Button
							onClick={handleClose}
							disabled={isSaveLoading}
						>
							Cancel
						</Button>
					</div>
				</div>
			</Modal>

			<ConfigurationSuccessModal
				open={open && activeModal === "success"}
				onClose={handleClose}
				firstRunAt={firstRunAt}
			/>
		</>
	)
}

export default ConfigureOptimizationModal

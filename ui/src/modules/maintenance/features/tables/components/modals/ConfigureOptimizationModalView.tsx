import { CaretUpIcon, QuestionIcon } from "@phosphor-icons/react"
import { Button, Input, Modal, Spin, Tabs, Tooltip } from "antd"
import clsx from "clsx"
import { useEffect, useState } from "react"

import { ErrorLogsModal } from "@/common/components/modals"

import CleanupTabContent from "./CleanupTabContent"
import ConfigurationSuccessModal from "./ConfigurationSuccessModal"
import {
	DEFAULT_CLEANUP_CONFIG,
	DEFAULT_CRON_CONFIG,
	DEFAULT_TARGET_FILE_SIZE,
	MAINTENANCE_TAB,
	RUN_TYPE_LABEL,
} from "../../constants"
import { DEFAULT_TABLE_MODAL_STYLES } from "../../constants"
import type {
	CleanupConfigFormModel,
	CronConfigOption,
	MaintenanceTabKey,
	TableCronFormModel,
	UpdateTableCronApiRequest,
} from "../../types"
import { getCronFromConfig, getNextRunsAcross } from "../../utils"
import FeatureToggleRow from "../FeatureToggleRow"
import ScheduleSection from "../ScheduleSection"

export type SaveCallbacks = {
	onSuccess: () => void
	onError: (logs: string[]) => void
}

export type ConfigureOptimizationModalViewProps = {
	open: boolean
	onClose: () => void
	title?: string
	headerChip?: React.ReactNode
	disableSaveReason?: string
	isConfigLoading?: boolean
	isConfigError?: boolean
	onRetryConfig?: () => void
	initialConfig?: TableCronFormModel
	onSave: (payload: UpdateTableCronApiRequest, callbacks: SaveCallbacks) => void
	isSaving: boolean
	/** Previously saved cleanup config; undefined means the defaults apply. */
	initialCleanupConfig?: CleanupConfigFormModel
	/** Persists the cleanup tab. UI-only today: callers write to the cleanup store. */
	onSaveCleanup: (config: CleanupConfigFormModel) => void
}

enum ActiveOptimizationModalState {
	SUCCESS = "success",
	ERROR_LOGS = "error-logs",
}

const ConfigureOptimizationModalView: React.FC<
	ConfigureOptimizationModalViewProps
> = ({
	open,
	onClose,
	title = "Configure Optimization",
	headerChip,
	disableSaveReason,
	isConfigLoading = false,
	isConfigError = false,
	onRetryConfig,
	initialConfig,
	onSave,
	isSaving,
	initialCleanupConfig,
	onSaveCleanup,
}) => {
	const [activeTab, setActiveTab] = useState<MaintenanceTabKey>(
		MAINTENANCE_TAB.OPTIMIZE,
	)
	const [minorCron, setMinorCron] =
		useState<CronConfigOption>(DEFAULT_CRON_CONFIG)
	const [majorCron, setMajorCron] =
		useState<CronConfigOption>(DEFAULT_CRON_CONFIG)
	const [fullCron, setFullCron] =
		useState<CronConfigOption>(DEFAULT_CRON_CONFIG)
	const [targetFileSize, setTargetFileSize] = useState(DEFAULT_TARGET_FILE_SIZE)
	const [advancedOpen, setAdvancedOpen] = useState(true)
	// Both features start switched on; the user can turn either off before saving.
	const [optimizationEnabled, setOptimizationEnabled] = useState(true)
	const [cleanupConfig, setCleanupConfig] = useState<CleanupConfigFormModel>(
		DEFAULT_CLEANUP_CONFIG,
	)
	const [activeModal, setActiveModal] =
		useState<ActiveOptimizationModalState | null>(null)
	const [errorLogs, setErrorLogs] = useState<string[]>([])

	// Populate form when config loads.
	useEffect(() => {
		if (!initialConfig || !open) return
		setMinorCron(initialConfig.minorCron)
		setMajorCron(initialConfig.majorCron)
		setFullCron(initialConfig.fullCron)
		if (initialConfig.targetFileSize) {
			setTargetFileSize(initialConfig.targetFileSize)
		}
	}, [initialConfig, open])

	useEffect(() => {
		if (!open) return
		setCleanupConfig(initialCleanupConfig ?? DEFAULT_CLEANUP_CONFIG)
	}, [initialCleanupConfig, open])

	// Reset all form and sub-modal state on close.
	useEffect(() => {
		if (!open) {
			setActiveTab(MAINTENANCE_TAB.OPTIMIZE)
			setActiveModal(null)
			setMinorCron(DEFAULT_CRON_CONFIG)
			setMajorCron(DEFAULT_CRON_CONFIG)
			setFullCron(DEFAULT_CRON_CONFIG)
			setTargetFileSize(DEFAULT_TARGET_FILE_SIZE)
			setAdvancedOpen(true)
			setOptimizationEnabled(true)
			setCleanupConfig(DEFAULT_CLEANUP_CONFIG)
			setErrorLogs([])
		}
	}, [open])

	// One save for both tabs: optimization goes to the API, cleanup to its store once that succeeds.
	const handleSave = () => {
		const payload: UpdateTableCronApiRequest = {
			minor_cron: getCronFromConfig(minorCron),
			major_cron: getCronFromConfig(majorCron),
			full_cron: getCronFromConfig(fullCron),
			target_file_size: targetFileSize,
			enabled_for_optimization: String(optimizationEnabled),
		}

		onSave(payload, {
			onSuccess: () => {
				onSaveCleanup(cleanupConfig)
				setActiveModal(ActiveOptimizationModalState.SUCCESS)
			},
			onError: logs => {
				setErrorLogs(logs)
				setActiveModal(ActiveOptimizationModalState.ERROR_LOGS)
			},
		})
	}

	const handleClose = () => {
		if (isSaving) return
		setActiveModal(null)
		onClose()
	}

	const optimizeRuns = optimizationEnabled
		? getNextRunsAcross([minorCron, majorCron, fullCron])
		: []
	const cleanupRuns = cleanupConfig.expireSnapshots
		? getNextRunsAcross([cleanupConfig.cron])
		: []
	const isTargetFileSizeValid = targetFileSize > 0
	const isCleanupValid =
		!cleanupConfig.expireSnapshots ||
		(cleanupConfig.olderThanDays > 0 && cleanupConfig.minSnapshotsToKeep > 0)

	const optimizeTabContent = (
		<>
			<div className="pt-6">
				<FeatureToggleRow
					label="Optimization"
					checked={optimizationEnabled}
					onChange={setOptimizationEnabled}
				/>
			</div>
			<ScheduleSection
				title={RUN_TYPE_LABEL.LITE}
				tooltip="Converts equality deletes to position deletes and merge small files"
				value={minorCron}
				onChange={setMinorCron}
				isFirst
			/>
			<ScheduleSection
				title={RUN_TYPE_LABEL.MEDIUM}
				tooltip="Applies deletes and merges data files"
				value={majorCron}
				onChange={setMajorCron}
			/>
			<ScheduleSection
				title={RUN_TYPE_LABEL.FULL}
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
							value={targetFileSize === 0 ? "" : String(targetFileSize)}
							onChange={e => {
								const numericValue = e.target.value.replace(/[^0-9]/g, "")
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
	)

	return (
		<>
			<Modal
				open={open && activeModal === null}
				onCancel={handleClose}
				footer={null}
				closable={false}
				maskClosable={!isSaving}
				keyboard={!isSaving}
				centered
				width={696}
				destroyOnHidden
				styles={DEFAULT_TABLE_MODAL_STYLES}
			>
				<div className="flex h-[808px] flex-col overflow-hidden bg-white">
					{/* Header */}
					<div className="px-8 pt-10">
						<h2 className="text-xl font-medium leading-7 text-olake-text">
							{title}
						</h2>
						{headerChip}
						<Tabs
							activeKey={activeTab}
							onChange={key => setActiveTab(key as MaintenanceTabKey)}
							className="mt-2 [&_.ant-tabs-nav]:mb-0"
							items={[
								{ key: MAINTENANCE_TAB.OPTIMIZE, label: "Optimize" },
								{ key: MAINTENANCE_TAB.CLEANUP, label: "Cleanup" },
							]}
						/>
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
									onClick={onRetryConfig}
								>
									Retry
								</Button>
							</div>
						) : isConfigLoading ? (
							<div className="flex h-full items-center justify-center">
								<Spin size="large" />
							</div>
						) : activeTab === MAINTENANCE_TAB.OPTIMIZE ? (
							optimizeTabContent
						) : (
							<CleanupTabContent
								value={cleanupConfig}
								onChange={setCleanupConfig}
							/>
						)}
					</div>

					{/* Footer: one save for both tabs */}
					<div className="flex h-20 shrink-0 items-center gap-3 border-t border-olake-border pl-8">
						<Tooltip title={disableSaveReason}>
							<Button
								type="primary"
								onClick={handleSave}
								loading={isSaving}
								disabled={
									!!disableSaveReason ||
									isConfigLoading ||
									isConfigError ||
									!isTargetFileSizeValid ||
									!isCleanupValid
								}
							>
								Complete Configuration
							</Button>
						</Tooltip>
						<Button
							onClick={handleClose}
							disabled={isSaving}
						>
							Cancel
						</Button>
					</div>
				</div>
			</Modal>

			<ConfigurationSuccessModal
				open={open && activeModal === ActiveOptimizationModalState.SUCCESS}
				onClose={handleClose}
				optimizeRuns={optimizeRuns}
				cleanupRuns={cleanupRuns}
			/>

			<ErrorLogsModal
				open={open && activeModal === ActiveOptimizationModalState.ERROR_LOGS}
				onClose={() => setActiveModal(null)}
				title="Failed to update optimization configuration"
				error={errorLogs.join("\n")}
				onAction={() => {
					setActiveModal(null)
					handleSave()
				}}
				actionButtonText="Retry"
			/>
		</>
	)
}

export default ConfigureOptimizationModalView

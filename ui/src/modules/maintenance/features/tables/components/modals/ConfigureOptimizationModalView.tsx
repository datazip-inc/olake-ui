import { Button, Modal, Spin, Tabs, Tooltip } from "antd"
import { useEffect, useState } from "react"

import { ErrorLogsModal } from "@/common/components/modals"

import CleanupTabContent from "./CleanupTabContent"
import ConfigurationSuccessModal from "./ConfigurationSuccessModal"
import OptimizeTabContent, {
	type OptimizeFormModel,
} from "./OptimizeTabContent"
import {
	DEFAULT_CLEANUP_CONFIG,
	DEFAULT_CRON_CONFIG,
	DEFAULT_TABLE_MODAL_STYLES,
	DEFAULT_TARGET_FILE_SIZE,
	MAINTENANCE_TAB,
	TABLE_MODAL_WIDTH,
} from "../../constants"
import type {
	CleanupConfigFormModel,
	MaintenanceTabKey,
	TableCronFormModel,
	UpdateTableCronApiRequest,
} from "../../types"
import {
	getCronConfigError,
	getCronFromConfig,
	getNextRunsAcross,
	mapCleanupConfigToApiRequest,
} from "../../utils"

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
	/** Saved settings of one table. Absent in bulk mode, where the form starts from defaults. */
	initialConfig?: TableCronFormModel
	onSave: (payload: UpdateTableCronApiRequest, callbacks: SaveCallbacks) => void
	isSaving: boolean
}

interface ConfigureFormState {
	optimize: OptimizeFormModel
	cleanup: CleanupConfigFormModel
	/** Bulk mode only writes cleanup when the user changed it, so other tables keep their own. */
	cleanupTouched: boolean
}

const INITIAL_FORM: ConfigureFormState = {
	optimize: {
		optimizationEnabled: true,
		minorCron: DEFAULT_CRON_CONFIG,
		majorCron: DEFAULT_CRON_CONFIG,
		fullCron: DEFAULT_CRON_CONFIG,
		targetFileSize: DEFAULT_TARGET_FILE_SIZE,
	},
	cleanup: DEFAULT_CLEANUP_CONFIG,
	cleanupTouched: false,
}

const toFormState = (config: TableCronFormModel): ConfigureFormState => ({
	optimize: {
		optimizationEnabled: config.optimizationEnabled,
		minorCron: config.minorCron,
		majorCron: config.majorCron,
		fullCron: config.fullCron,
		targetFileSize: config.targetFileSize || DEFAULT_TARGET_FILE_SIZE,
	},
	cleanup: config.cleanup,
	cleanupTouched: false,
})

enum ActiveOptimizationModalState {
	SUCCESS = "success",
	ERROR_LOGS = "error-logs",
}

const TAB_ITEMS = [
	{ key: MAINTENANCE_TAB.OPTIMIZE, label: "Optimize" },
	{ key: MAINTENANCE_TAB.CLEANUP, label: "Cleanup" },
]

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
}) => {
	const [activeTab, setActiveTab] = useState<MaintenanceTabKey>(
		MAINTENANCE_TAB.OPTIMIZE,
	)
	const [form, setForm] = useState<ConfigureFormState>(INITIAL_FORM)
	const [activeModal, setActiveModal] =
		useState<ActiveOptimizationModalState | null>(null)
	const [errorLogs, setErrorLogs] = useState<string[]>([])

	// Populate when the saved config arrives; reset everything on close.
	useEffect(() => {
		if (!open) {
			setActiveTab(MAINTENANCE_TAB.OPTIMIZE)
			setActiveModal(null)
			setErrorLogs([])
			setForm(INITIAL_FORM)
			return
		}
		if (initialConfig) setForm(toFormState(initialConfig))
	}, [initialConfig, open])

	// Single mode always writes cleanup; bulk mode only after the user touched that tab.
	const includeCleanup = initialConfig !== undefined || form.cleanupTouched
	const { optimize, cleanup } = form

	// One save for both tabs: every property goes out in a single ALTER TABLE batch.
	const handleSave = () => {
		const payload: UpdateTableCronApiRequest = {
			minor_cron: getCronFromConfig(optimize.minorCron),
			major_cron: getCronFromConfig(optimize.majorCron),
			full_cron: getCronFromConfig(optimize.fullCron),
			target_file_size: optimize.targetFileSize,
			enabled_for_optimization: String(optimize.optimizationEnabled),
			...(includeCleanup ? mapCleanupConfigToApiRequest(cleanup) : {}),
		}

		onSave(payload, {
			onSuccess: () => setActiveModal(ActiveOptimizationModalState.SUCCESS),
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

	const compactionCrons = [
		optimize.minorCron,
		optimize.majorCron,
		optimize.fullCron,
	]
	const activeCrons = [
		...compactionCrons,
		...(includeCleanup && cleanup.expireSnapshots ? [cleanup.cron] : []),
	]
	const hasInvalidCron = activeCrons.some(
		cron => getCronConfigError(cron) !== null,
	)
	const isCleanupValid =
		!includeCleanup ||
		!cleanup.expireSnapshots ||
		(cleanup.keepDuration.value > 0 && cleanup.minSnapshotsToKeep > 0)
	const canSave =
		!disableSaveReason &&
		!isConfigLoading &&
		!isConfigError &&
		optimize.targetFileSize > 0 &&
		isCleanupValid &&
		!hasInvalidCron

	const optimizeRuns = optimize.optimizationEnabled
		? getNextRunsAcross(compactionCrons)
		: []
	const cleanupRuns = includeCleanup
		? cleanup.expireSnapshots
			? getNextRunsAcross([cleanup.cron])
			: []
		: undefined

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
				width={TABLE_MODAL_WIDTH}
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
							items={TAB_ITEMS}
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
							<OptimizeTabContent
								value={optimize}
								onChange={next =>
									setForm(prev => ({ ...prev, optimize: next }))
								}
							/>
						) : (
							<CleanupTabContent
								value={cleanup}
								onChange={next =>
									setForm(prev => ({
										...prev,
										cleanup: next,
										cleanupTouched: true,
									}))
								}
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
								disabled={!canSave}
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

import { Button, Modal, Spin, Tabs, Tooltip } from "antd"
import { useEffect, useState } from "react"

import { ErrorLogsModal } from "@/common/components/modals"

import CleanupTabContent from "./CleanupTabContent"
import ConfigurationSuccessModal from "./ConfigurationSuccessModal"
import OptimizeTabContent from "./OptimizeTabContent"
import {
	CONFIGURE_MODE,
	DEFAULT_TABLE_CONFIG_FORM,
	DEFAULT_TABLE_MODAL_STYLES,
	MAINTENANCE_TAB,
	MAINTENANCE_TAB_ITEMS,
	MIN_SNAPSHOTS_TO_KEEP_LOWEST,
	OPTIMIZE_SCHEDULES,
} from "../../constants"
import type {
	ConfigureMode,
	MaintenanceTabKey,
	TableConfigFormModel,
	UpdateTableConfigApiRequest,
} from "../../types"
import {
	getCronConfigError,
	getNextRunsAcross,
	mapCleanupConfigToApiRequest,
	mapOptimizeConfigToApiRequest,
} from "../../utils"

export type SaveCallbacks = {
	onSuccess: () => void
	onError: (logs: string[]) => void
}

export type ConfigureOptimizationModalViewProps = {
	open: boolean
	onClose: () => void
	mode: ConfigureMode
	title?: string
	headerChip?: React.ReactNode
	disableSaveReason?: string
	isConfigLoading?: boolean
	isConfigError?: boolean
	onRetryConfig?: () => void
	// Saved settings of the table in single mode. Bulk mode starts from defaults.
	initialConfig?: TableConfigFormModel
	onSave: (
		payload: UpdateTableConfigApiRequest,
		callbacks: SaveCallbacks,
	) => void
	isSaving: boolean
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
	mode,
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
	const [form, setForm] = useState<TableConfigFormModel>(
		DEFAULT_TABLE_CONFIG_FORM,
	)
	// Bulk mode only writes cleanup when the user changed it, so other tables keep their own.
	const [cleanupTouched, setCleanupTouched] = useState(false)
	const [activeModal, setActiveModal] =
		useState<ActiveOptimizationModalState | null>(null)
	const [errorLogs, setErrorLogs] = useState<string[]>([])

	// Populate when the saved config arrives; reset everything on close.
	useEffect(() => {
		if (!open) {
			setActiveTab(MAINTENANCE_TAB.OPTIMIZE)
			setActiveModal(null)
			setErrorLogs([])
			setForm(DEFAULT_TABLE_CONFIG_FORM)
			setCleanupTouched(false)
			return
		}
		if (initialConfig) {
			const { optimize, cleanup } = initialConfig
			setForm({ optimize, cleanup })
		}
	}, [initialConfig, open])

	const includeCleanup = mode === CONFIGURE_MODE.SINGLE || cleanupTouched
	const { optimize, cleanup } = form

	// One save for both tabs: every property goes out in a single ALTER TABLE batch.
	const handleSave = () => {
		const payload: UpdateTableConfigApiRequest = {
			...mapOptimizeConfigToApiRequest(optimize),
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

	const compactionCrons = OPTIMIZE_SCHEDULES.map(
		schedule => optimize[schedule.key],
	)
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
		(cleanup.keepDuration.value > 0 &&
			cleanup.minSnapshotsToKeep >= MIN_SNAPSHOTS_TO_KEEP_LOWEST)
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
							items={MAINTENANCE_TAB_ITEMS}
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
								onChange={next => {
									setCleanupTouched(true)
									setForm(prev => ({ ...prev, cleanup: next }))
								}}
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

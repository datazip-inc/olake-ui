import { InfoIcon, SlidersIcon, WarningIcon } from "@phosphor-icons/react"
import { InputNumber, Select, Spin, Tag, Tooltip } from "antd"
import clsx from "clsx"
import React from "react"

import { restrictNumericInput } from "@/common/utils"
import { useAvailableQueryEngines } from "@/modules/ingestion/features/sources/hooks"

import { useJobConfigurationStore } from "../stores"

interface AdvancedSettingsCardProps {
	sourceType?: string
	sourceVersion?: string
	queryEnginesDisabled?: boolean
}

const AdvancedSettingsCard: React.FC<AdvancedSettingsCardProps> = ({
	sourceType,
	sourceVersion,
	queryEnginesDisabled = false,
}) => {
	const { advancedSettings, setAdvancedSettings, selectedSource } =
		useJobConfigurationStore()

	const {
		data: queryEngines = [],
		isLoading: isLoadingQueryEngines,
		isFetching: isFetchingQueryEngines,
		isError: isQueryEnginesError,
		refetch: refetchQueryEngines,
	} = useAvailableQueryEngines(
		(sourceType ?? selectedSource?.type ?? "").toLowerCase(),
		sourceVersion ?? selectedSource?.version ?? "",
	)

	const targetQueryEngines = advancedSettings?.target_query_engines ?? []
	const setQueryEngines = (target_query_engines: string[]) =>
		setAdvancedSettings({ ...advancedSettings, target_query_engines })

	return (
		<div className="mt-5 rounded-xl border border-olake-border p-6">
			<div className="mb-6 flex items-center gap-2">
				<SlidersIcon className="size-5" />
				<span className="text-base font-medium text-gray-900">
					Advanced Settings
				</span>
			</div>
			{(isLoadingQueryEngines ||
				isQueryEnginesError ||
				queryEngines.length > 0) && (
				<div className="mb-6 border-b border-olake-border pb-6">
					<div className="w-1/2">
						<div className="mb-2 flex items-center gap-1">
							<label className="text-sm text-gray-600">
								Target Query Engines
							</label>
							<Tooltip
								title={
									queryEnginesDisabled
										? "To change these, open Edit Streams, go back to Job Config, then select Next to re-discover streams."
										: "Query engines that will query the tables written by this job."
								}
							>
								<InfoIcon
									size={16}
									className="cursor-help text-slate-900"
								/>
							</Tooltip>
						</div>
						{isLoadingQueryEngines ? (
							<div className="flex items-center gap-x-2">
								<Spin size="small" />
								<span className="">Fetching available query engines</span>
							</div>
						) : isQueryEnginesError ? (
							<div className="flex items-center gap-2 text-sm text-danger">
								<WarningIcon className="size-4 shrink-0" />
								Failed to fetch available query engines.
								<button
									type="button"
									className="font-medium underline underline-offset-2 hover:opacity-80"
									onClick={() => refetchQueryEngines()}
								>
									Retry
								</button>
							</div>
						) : (
							<Select
								mode="multiple"
								allowClear
								className="w-full"
								placeholder="Select Query Engines"
								loading={isFetchingQueryEngines}
								disabled={queryEnginesDisabled}
								maxTagCount="responsive"
								maxTagPlaceholder={omitted => `+${omitted.length} more`}
								value={targetQueryEngines}
								onChange={setQueryEngines}
								options={queryEngines.map(({ engine, label }) => ({
									label,
									value: engine,
								}))}
								tagRender={({ label, closable, onClose }) => (
									<Tag
										closable={closable}
										// Keeps the dropdown from toggling when the tag is closed.
										onMouseDown={event => event.preventDefault()}
										onClose={onClose}
										className={clsx(
											"mr-1 rounded font-medium",
											queryEnginesDisabled
												? "border-neutral-disabled bg-neutral-light text-gray-500"
												: "border-primary-100 bg-primary-50 text-brand-blue",
										)}
									>
										{label}
									</Tag>
								)}
							/>
						)}
					</div>
				</div>
			)}

			<div className="mb-4 text-sm font-medium text-gray-900">
				Additional Settings
			</div>

			<div className="flex w-2/5 flex-wrap gap-x-12 gap-y-6">
				{/* Max Discover Threads */}
				<div className="w-full">
					<div className="mb-2 flex items-center gap-1">
						<label className="text-sm text-gray-600">
							Max Discover Threads
						</label>
						<Tooltip title="Max number of parallel threads for discovery of table in database">
							<InfoIcon
								size={16}
								className="cursor-help text-slate-900"
							/>
						</Tooltip>
					</div>
					<InputNumber
						min={1}
						precision={0}
						className="w-full"
						value={advancedSettings?.max_discover_threads}
						onChange={val =>
							setAdvancedSettings({
								...advancedSettings,
								max_discover_threads: val,
							})
						}
						placeholder="50"
						onKeyDown={restrictNumericInput}
					/>
				</div>
			</div>
		</div>
	)
}

export default AdvancedSettingsCard

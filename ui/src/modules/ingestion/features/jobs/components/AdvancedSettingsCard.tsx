import { InfoIcon, SlidersIcon, WarningIcon } from "@phosphor-icons/react"
import { InputNumber, Select, Spin, Tag, Tooltip } from "antd"
import React from "react"
import semver from "semver"

import { restrictNumericInput } from "@/common/utils"
import { MIN_QUERY_ENGINES_VERSION } from "@/modules/ingestion/common/constants"
import { useAvailableQueryEngines } from "@/modules/ingestion/features/sources/hooks"

import { useJobConfigurationStore } from "../stores"

interface AdvancedSettingsCardProps {
	sourceType?: string
	sourceVersion?: string
}

const AdvancedSettingsCard: React.FC<AdvancedSettingsCardProps> = ({
	sourceType,
	sourceVersion,
}) => {
	const {
		advancedSettings,
		setAdvancedSettings,
		selectedSource,
		savedAdvancedSettings,
		isEditMode,
	} = useJobConfigurationStore()

	const version = sourceVersion ?? selectedSource?.version ?? ""
	// An unknown or unparsable version is treated as supported.
	const isVersionSupported =
		!semver.valid(version) || semver.gte(version, MIN_QUERY_ENGINES_VERSION)

	const {
		data: queryEngines = [],
		isLoading: isLoadingQueryEngines,
		isFetching: isFetchingQueryEngines,
		isError: isQueryEnginesError,
		refetch: refetchQueryEngines,
	} = useAvailableQueryEngines(
		(sourceType ?? selectedSource?.type ?? "").toLowerCase(),
		isVersionSupported ? version : "",
	)

	const targetQueryEngines = advancedSettings?.target_query_engines ?? []
	const setQueryEngines = (target_query_engines: string[]) =>
		setAdvancedSettings({ ...advancedSettings, target_query_engines })

	const savedQueryEngines = savedAdvancedSettings?.target_query_engines ?? []
	// Hidden in create mode, or when Snowflake was already in the saved selection.
	const showSnowflakeWarning =
		isEditMode &&
		targetQueryEngines.some(
			engine =>
				!savedQueryEngines.includes(engine) &&
				engine.toLowerCase().includes("snowflake"),
		)

	return (
		<div className="mt-5 rounded-xl border border-olake-border p-6">
			<div className="mb-6 flex items-center gap-2">
				<SlidersIcon className="size-5" />
				<span className="text-base font-medium text-gray-900">
					Advanced Settings
				</span>
			</div>
			{(!isVersionSupported ||
				isLoadingQueryEngines ||
				isQueryEnginesError ||
				queryEngines.length > 0) && (
				<div className="mb-6 border-b border-olake-border pb-6">
					<div className="w-1/2">
						<div className="mb-2 flex items-center gap-1">
							<label className="text-sm text-gray-600">
								Target Query Engines
							</label>
							<Tooltip title="Query engines that will query the tables written by this job.">
								<InfoIcon
									size={16}
									className="cursor-help text-slate-900"
								/>
							</Tooltip>
						</div>
						{!isVersionSupported ? (
							<div className="flex items-center gap-2 text-sm text-gray-500">
								<InfoIcon className="size-4 shrink-0" />
								Upgrade your source version to {MIN_QUERY_ENGINES_VERSION} or
								later to use this option.
							</div>
						) : isLoadingQueryEngines ? (
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
										className="mr-1 rounded border-primary-100 bg-primary-50 font-medium text-brand-blue"
									>
										{label}
									</Tag>
								)}
							/>
						)}
						{showSnowflakeWarning && (
							<div className="mt-4 flex gap-3 rounded-lg border border-warning/40 bg-warning-light p-4 text-xs text-warning-dark">
								<InfoIcon className="mt-0.5 size-5 shrink-0" />
								<p>
									We highly recommend running{" "}
									<span className="font-semibold">Clear Destination</span>{" "}
									before proceeding. This is due to a{" "}
									<span className="font-semibold">
										Snowflake-side limitation
									</span>{" "}
									that can cause stale data to be read from old Iceberg
									snapshots if the destination is not cleared.
								</p>
							</div>
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

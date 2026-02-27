import React from "react"
import { InputNumber, Tooltip } from "antd"
import { InfoIcon, SlidersIcon } from "@phosphor-icons/react"

import { AdvancedSettings } from "../../../types"
import { restrictNumericInput } from "../../../utils/utils"

interface AdvancedSettingsCardProps {
	advancedSettings: AdvancedSettings | null
	setAdvancedSettings: (settings: AdvancedSettings | null) => void
}

const AdvancedSettingsCard: React.FC<AdvancedSettingsCardProps> = ({
	advancedSettings,
	setAdvancedSettings,
}) => {
	return (
		<div className="mt-5 rounded-xl border border-[#D9D9D9] p-6">
			<div className="mb-6 flex items-center gap-2">
				<SlidersIcon className="size-5" />
				<span className="text-base font-medium text-gray-900">
					Advanced Settings
				</span>
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

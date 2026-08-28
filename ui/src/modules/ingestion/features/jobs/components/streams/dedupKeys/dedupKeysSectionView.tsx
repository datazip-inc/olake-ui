import { InfoIcon, WarningIcon } from "@phosphor-icons/react"
import { Select, Tooltip } from "antd"
import clsx from "clsx"

interface DedupKeysSectionViewProps {
	options: string[]
	value: string[]
	disabled?: boolean
	isDirty?: boolean
	onChange: (keys: string[]) => void
}

const DedupKeysSectionView = ({
	options,
	value,
	disabled,
	isDirty,
	onChange,
}: DedupKeysSectionViewProps) => {
	return (
		<div
			className={clsx("mb-4", disabled ? "text-gray-500" : "text-neutral-text")}
		>
			<div className="mb-3">
				<div className="flex items-center gap-1">
					{isDirty && <WarningIcon className="size-4 text-orange-500" />}
					<label className="block w-full">Dedup Keys:</label>
					<Tooltip title="Selecting only _kafka_key enables tombstone deletes. Other selections have Upsert without deletes">
						<InfoIcon className="size-4" />
					</Tooltip>
				</div>
				<div className="text-xs text-neutral-700">
					Required for Upsert. Identifies the entity in destination
				</div>
			</div>
			<Select
				mode="multiple"
				className="w-full"
				disabled={disabled}
				placeholder="Select dedup keys"
				options={options.map(o => ({ label: o, value: o }))}
				value={value}
				onChange={onChange}
			/>
		</div>
	)
}

export default DedupKeysSectionView

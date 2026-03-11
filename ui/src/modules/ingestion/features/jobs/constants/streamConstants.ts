export const PARTITIONING_COLUMNS = [
	{
		title: "Column name",
		dataIndex: "name",
		key: "name",
	},
	{
		title: "Granularity",
		dataIndex: "granularity",
		key: "granularity",
	},
	{
		title: "Default",
		dataIndex: "default",
		key: "default",
	},
]

// fallback defaults for streams
export const STREAM_DEFAULTS = {
	append_mode: false,
	normalization: false,
	partition_regex: "",
	filter: "",
} as const

export const SYNC_MODE_MAP = {
	FULL_REFRESH: "full_refresh",
	INCREMENTAL: "incremental",
	CDC: "cdc",
	STRICT_CDC: "strict_cdc",
}

export const PartitioningRegexTooltip =
	"Choose a column to partition your data for faster reads and better performance"

export const DESTINATION_TABLE_TOOLTIP_TEXT =
	"Defines the table’s appearance and its destination database where it will be stored"

export const DESTINATATION_DATABASE_TOOLTIP_TEXT =
	"The name of the destination database where synced tables will be accessible for querying"

export const FORMAT_OPTIONS = {
	DYNAMIC: "dynamic",
	CUSTOM: "custom",
} as const

export const NAMESPACE_PLACEHOLDER = "_${source_namespace}"

export const LABELS = {
	S3: {
		title: "S3 Folder Name",
		folderType: "S3",
	},
	ICEBERG: {
		title: "Iceberg Database Name",
		folderType: "Iceberg DB",
	},
} as const

export const TAB_STYLES = {
	active: "border border-primary bg-white text-primary rounded-md py-1 px-2",
	inactive: "bg-background-primary text-slate-900 py-1 px-2",
	hover: "hover:text-primary",
}

export const CARD_STYLE = "rounded-xl border border-[#E3E3E3] p-3"

export const COLORS = {
	selected: {
		border: "#203FDD",
		text: "#203FDD",
	},
	unselected: {
		border: "#D9D9D9",
		text: "#575757",
	},
} as const

import { UpsertType } from "@/modules/ingestion/common/types"

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

export const UPSERT_TYPE_OPTIONS = [
	{
		label: "Equality Deletes",
		value: UpsertType.EQUALITY,
		icebergFormatVersion: "V2",
		tooltip:
			"Rows are matched and deleted by primary key values before the new rows are written",
	},
	{
		label: "Positional Deletes",
		value: UpsertType.POSITIONAL,
		icebergFormatVersion: "V2",
		tooltip:
			"Rows are deleted by their position in the existing data files before the new rows are written",
	},
	{
		label: "Deletion Vector",
		value: UpsertType.DELETION_VECTOR,
		icebergFormatVersion: "V3",
		tooltip:
			"Deleted rows are marked in a per-file bitmap instead of a delete file, which Iceberg V3 readers resolve",
	},
]

// Catalogs discovered before target query engines existed carry no
// available_update_types; back then only these two formats were writable.
export const DEFAULT_AVAILABLE_UPDATE_TYPES = [
	UpsertType.EQUALITY,
	UpsertType.POSITIONAL,
]

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

export const CARD_STYLE = "rounded-xl border border-neutral-border p-3"

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

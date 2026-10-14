import { CpuIcon, DatabaseIcon } from "@phosphor-icons/react"

import {
	DESTINATION_INTERNAL_TYPES,
	SOURCE_INTERNAL_TYPES,
} from "@/modules/ingestion/common/constants"

import { IngestionMode } from "../enums"
import type { TaskLogSource } from "../types"

/**
 * Matches a single or compound filter expression of the form:
 *  - column operator value
 *  - column operator value (and|or) column operator value
 *
 * Operators: >=, <=, !=, >, <, =
 * Value can be a quoted string ("..."), a number (int/float), or a word (\w+)
 *
 * Capture groups:
 *  1: first column name
 *  2: first operator
 *  3: first value
 *  4: logical operator (and|or) [optional]
 *  5: second column name [optional]
 *  6: second operator [optional]
 *  7: second value [optional]
 */
export const FILTER_REGEX =
	/^(\w+)\s*(>=|<=|!=|>|<|=)\s*("[^"]+"|\d*\.?\d+|\w+)\s*(?:(and|or)\s*(\w+)\s*(>=|<=|!=|>|<|=)\s*("[^"]+"|\d*\.?\d+|\w+))?\s*$/

const DB_STANDARD_MODES = [IngestionMode.APPEND, IngestionMode.UPSERT] as const
const APPEND_ONLY_MODE = [IngestionMode.APPEND] as const

export const SOURCE_SUPPORTED_INGESTION_MODES = {
	[SOURCE_INTERNAL_TYPES.MONGODB]: DB_STANDARD_MODES,
	[SOURCE_INTERNAL_TYPES.POSTGRES]: DB_STANDARD_MODES,
	[SOURCE_INTERNAL_TYPES.MYSQL]: DB_STANDARD_MODES,
	[SOURCE_INTERNAL_TYPES.ORACLE]: DB_STANDARD_MODES,
	[SOURCE_INTERNAL_TYPES.DB2]: DB_STANDARD_MODES,
	[SOURCE_INTERNAL_TYPES.MSSQL]: DB_STANDARD_MODES,
	[SOURCE_INTERNAL_TYPES.KAFKA]: APPEND_ONLY_MODE,
	[SOURCE_INTERNAL_TYPES.S3]: APPEND_ONLY_MODE,
} as const

export const DESTINATION_SUPPORTED_INGESTION_MODES = {
	[DESTINATION_INTERNAL_TYPES.S3]: APPEND_ONLY_MODE,
	[DESTINATION_INTERNAL_TYPES.ICEBERG]: DB_STANDARD_MODES,
} as const

export const operatorOptions = [
	{ label: "=", value: "=" },
	{ label: "!=", value: "!=" },
	{ label: ">", value: ">" },
	{ label: "<", value: "<" },
	{ label: ">=", value: ">=" },
	{ label: "<=", value: "<=" },
]

// Logs pagination configuration
export const LOGS_CONFIG = {
	DEFAULT_CURSOR: "", // empty cursor means start from the end of the logs
	INITIAL_BATCH_SIZE: 1000, // First load
	SUBSEQUENT_BATCH_SIZE: 500, // Subsequent loads
	MAX_LOGS_IN_MEMORY: 10000, // Maximum logs to keep in memory
	VIRTUAL_LIST_START_INDEX: 1000000, // High starting index for virtualized log list
	OVERSCAN: 1000, // Number of items to render outside visible area
} as const

// Task log sources: per-source styling (badge, short accent bar, row tint)
// and the tabs of the logs page.
export const LOG_SOURCE_STYLES = {
	sync: {
		label: "SYNC",
		Icon: DatabaseIcon,
		badgeClass: "border-indigo-300 bg-indigo-50 text-indigo-700",
		accentClass: "bg-indigo-400",
		rowClass: "",
	},
	worker: {
		label: "WORKER",
		Icon: CpuIcon,
		badgeClass: "border-teal-300 bg-teal-50 text-teal-700",
		accentClass: "bg-teal-600",
		rowClass: "bg-teal-50/30",
	},
} as const

export const LOG_SOURCE_TABS: { value: TaskLogSource; label: string }[] = [
	{ value: "all", label: "All logs" },
	{ value: "sync", label: "Sync logs" },
	{ value: "worker", label: "Worker logs" },
]

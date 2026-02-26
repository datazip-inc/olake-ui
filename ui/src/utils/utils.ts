import React from "react"
import { message } from "antd"
import parser from "cron-parser"

import {
	CronParseResult,
	JobType,
	IngestionMode,
	SelectedStream,
	CursorFieldValues,
	LogEntry,
	TaskLogEntry,
	SelectedStreamsByNamespace,
	FilterConfig,
	FilterConfigCondition,
} from "../types"
import type { StreamData } from "../types/streamTypes"
import {
	ReleasesResponse,
	ReleaseType,
	ReleaseTypeData,
} from "../types/platformTypes"
import {
	DAYS_MAP,
	DESTINATION_INTERNAL_TYPES,
	DESTINATION_LABELS,
	FILTER_REGEX,
	SOURCE_SUPPORTED_INGESTION_MODES,
	DESTINATION_SUPPORTED_INGESTION_MODES,
} from "./constants"
import {
	AWSS3,
	ApacheIceBerg,
	DB2,
	Kafka,
	MongoDB,
	MySQL,
	Oracle,
	Postgres,
	MSSQL,
} from "../assets"

// Normalizes old connector types to their current internal types
export const normalizeConnectorType = (connectorType: string): string => {
	const lowerType = connectorType.toLowerCase()

	switch (lowerType) {
		case "s3":
			return "s3"
		case "amazon s3":
			return "parquet"
		case "iceberg":
		case "apache iceberg":
			return "iceberg"
		default:
			return connectorType
	}
}

// These are used to show in connector dropdowns
export const getConnectorImage = (connector: string) => {
	const normalizedConnector = normalizeConnectorType(connector).toLowerCase()

	switch (normalizedConnector) {
		case "mongodb":
			return MongoDB
		case "postgres":
			return Postgres
		case "mysql":
			return MySQL
		case "oracle":
			return Oracle
		case DESTINATION_INTERNAL_TYPES.S3:
			return AWSS3
		case DESTINATION_INTERNAL_TYPES.ICEBERG:
			return ApacheIceBerg
		case "kafka":
			return Kafka
		case "s3":
			return AWSS3
		case "db2":
			return DB2
		case "mssql":
			return MSSQL
		default:
			// Default placeholder
			return MongoDB
	}
}

// These are used to show documentation path for the connector
export const getConnectorDocumentationPath = (
	connector: string,
	catalog: string | null,
) => {
	switch (connector) {
		case "Amazon S3":
			return "s3/config"
		case "Apache Iceberg":
			switch (catalog) {
				case "glue":
					return "iceberg/catalog/glue"
				case "rest":
					return "iceberg/catalog/rest"
				case "jdbc":
					return "iceberg/catalog/jdbc"
				case "hive":
					return "iceberg/catalog/hive"
				default:
					return "iceberg/catalog/glue"
			}
		default:
			return undefined
	}
}

export const getStatusClass = (status: string) => {
	switch (status.toLowerCase()) {
		case "success":
		case "completed":
			return "text-[#52C41A] bg-[#F6FFED]"
		case "failed":
			return "text-[#F5222D] bg-[#FFF1F0]"
		case "canceled":
			return "text-amber-700 bg-amber-50"
		case "running":
			return "text-primary-700 bg-primary-200"
		case "scheduled":
			return "text-[rgba(0,0,0,88)] bg-neutral-light"
		default:
			return "text-[rgba(0,0,0,88)] bg-transparent"
	}
}

export const getJobTypeClass = (jobType: JobType) => {
	switch (jobType) {
		case JobType.Sync:
			return "text-[#52C41A] bg-[#F6FFED]"
		case JobType.ClearDestination:
			return "text-amber-700 bg-amber-50"
		default:
			return "text-[rgba(0,0,0,88)] bg-transparent"
	}
}

export const getConnectorInLowerCase = (connector?: string | null) => {
	const normalizedConnector = normalizeConnectorType(connector || "")
	const lowerConnector = normalizedConnector.toLowerCase()

	switch (lowerConnector) {
		case DESTINATION_INTERNAL_TYPES.S3:
		case DESTINATION_LABELS.AMAZON_S3:
			return DESTINATION_INTERNAL_TYPES.S3
		case DESTINATION_INTERNAL_TYPES.ICEBERG:
		case DESTINATION_LABELS.APACHE_ICEBERG:
			return DESTINATION_INTERNAL_TYPES.ICEBERG
		case "s3":
			return "s3"
		case "mongodb":
			return "mongodb"
		case "postgres":
			return "postgres"
		case "mysql":
			return "mysql"
		case "oracle":
			return "oracle"
		case "db2":
			return "db2"
		case "mssql":
			return "mssql"
		default:
			return lowerConnector
	}
}

export const getStatusLabel = (status: string) => {
	switch (status) {
		case "success":
			return "Success"
		case "failed":
			return "Failed"
		case "canceled":
			return "Canceled"
		case "running":
			return "Running"
		case "scheduled":
			return "Scheduled"
		case "completed":
			return "Completed"
		default:
			return status
	}
}

export const getJobTypeLabel = (lastRunType: JobType) => {
	switch (lastRunType) {
		case JobType.Sync:
			return "Sync"
		case JobType.ClearDestination:
			return "Clear Destination"
		default:
			return lastRunType
	}
}

export const getConnectorLabel = (type: string): string => {
	switch (type) {
		case "mongodb":
		case "MongoDB":
			return "MongoDB"
		case "postgres":
		case "Postgres":
			return "Postgres"
		case "mysql":
		case "MySQL":
			return "MySQL"
		case "oracle":
		case "Oracle":
			return "Oracle"
		case "kafka":
			return "Kafka"
		case "s3":
		case "S3":
			return "S3"
		case "db2":
			return "DB2"
		case "mssql":
			return "MSSQL"
		default:
			return "MongoDB"
	}
}

export const getFrequencyValue = (frequency: string) => {
	if (frequency.includes(" ")) {
		const parts = frequency.split(" ")
		const unit = parts[1].toLowerCase()

		switch (true) {
			case unit.includes("hour"):
				return "hours"
			case unit.includes("minute"):
				return "minutes"
			case unit.includes("day"):
				return "days"
			case unit.includes("week"):
				return "weeks"
			case unit.includes("month"):
				return "months"
			case unit.includes("year"):
				return "years"
			default:
				return "hours"
		}
	}

	switch (frequency) {
		case "hourly":
		case "hours":
			return "hours"
		case "daily":
		case "days":
			return "days"
		case "weekly":
		case "weeks":
			return "weeks"
		case "monthly":
		case "months":
			return "months"
		case "yearly":
		case "years":
			return "years"
		case "minutes":
			return "minutes"
		case "custom":
			return "custom"
		default:
			return "hours"
	}
}

// removes the saved job from local storage when user deletes the job or completes entire flow and create
export const removeSavedJobFromLocalStorage = (jobId: string) => {
	const savedJobs = localStorage.getItem("savedJobs")
	if (savedJobs) {
		const jobs = JSON.parse(savedJobs)
		const filteredJobs = jobs.filter((job: any) => job.id !== jobId)
		localStorage.setItem("savedJobs", JSON.stringify(filteredJobs))
	}
}

export const getReplicationFrequency = (replicationFrequency: string) => {
	if (replicationFrequency.includes(" ")) {
		const parts = replicationFrequency.split(" ")
		const value = parts[0]
		const unit = parts[1].toLowerCase()

		if (unit.includes("minute")) return `${value} minutes`
		if (unit.includes("hour")) return "hourly"
		if (unit.includes("day")) return "daily"
		if (unit.includes("week")) return "weekly"
		if (unit.includes("month")) return "monthly"
		if (unit.includes("year")) return "yearly"
	}

	if (replicationFrequency === "minutes") {
		return "minutes"
	} else if (replicationFrequency === "hours") {
		return "hourly"
	} else if (replicationFrequency === "days") {
		return "daily"
	} else if (replicationFrequency === "weeks") {
		return "weekly"
	} else if (replicationFrequency === "months") {
		return "monthly"
	} else if (replicationFrequency === "years") {
		return "yearly"
	}
}

export const getLogLevelClass = (level: string) => {
	switch (level) {
		case "debug":
			return "text-blue-600 bg-[#F0F5FF]"
		case "info":
			return "text-[#531DAB] bg-[#F9F0FF]"
		case "warning":
		case "warn":
			return "text-[#FAAD14] bg-[#FFFBE6]"
		case "error":
		case "fatal":
			return "text-red-500 bg-[#FFF1F0]"
		default:
			return "text-gray-600"
	}
}

export const getLogTextColor = (level: string) => {
	switch (level) {
		case "warning":
		case "warn":
			return "text-[#FAAD14]"
		case "error":
		case "fatal":
			return "text-[#F5222D]"
		default:
			return "text-[#000000"
	}
}

export const mapLogEntriesToTaskLogEntries = (
	logs: LogEntry[],
): TaskLogEntry[] => {
	return logs.map(log => {
		const level = log.level ?? ""
		const message = log.message ?? ""
		const timeRaw = log.time ?? ""

		let date = ""
		let time = ""

		if (timeRaw) {
			const dateObj = new Date(timeRaw)
			date = dateObj.toLocaleDateString()
			time = dateObj.toLocaleTimeString("en-US", {
				timeZone: "UTC",
				hour12: false,
			})
		}

		return {
			level,
			message,
			time,
			date,
		}
	})
}

export const getDayNumber = (day: string): number => {
	return DAYS_MAP[day as keyof typeof DAYS_MAP]
}

export const generateCronExpression = (
	frequency: string,
	time: string,
	ampm: "AM" | "PM",
	day: string,
) => {
	let hour = parseInt(time)
	if (ampm === "PM" && hour !== 12) {
		hour += 12
	} else if (ampm === "AM" && hour === 12) {
		hour = 0
	}

	let cronExp = ""
	switch (frequency) {
		case "minutes":
			cronExp = "* * * * *" // Every minute
			break
		case "hours":
			cronExp = "0 * * * *" // Every hour at minute 0
			break
		case "days":
			cronExp = `0 ${hour} * * *` // Every day at specified hour
			break
		case "weeks":
			const dayNumber = getDayNumber(day)
			cronExp = `0 ${hour} * * ${dayNumber}` // Every week on specified day at specified hour
			break
		default:
			cronExp = "* * * * *" // Default to every minute if no frequency specified
	}
	return cronExp
}

export const operatorOptions = [
	{ label: "=", value: "=" },
	{ label: "!=", value: "!=" },
	{ label: ">", value: ">" },
	{ label: "<", value: "<" },
	{ label: ">=", value: ">=" },
	{ label: "<=", value: "<=" },
]

export const isValidCronExpression = (cron: string): boolean => {
	// Check if the cron has exactly 5 parts
	const parts = cron.trim().split(" ")
	if (parts.length !== 5) return false

	try {
		parser.parse(cron)
		return true
	} catch {
		return false
	}
}

export const parseCronExpression = (
	cronExpression: string,
	DAYS: string[],
): CronParseResult => {
	try {
		const parts = cronExpression.split(" ")
		if (parts.length !== 5) {
			return { frequency: "custom", customCronExpression: cronExpression }
		}

		const [minute, hour, dayOfMonth, month, dayOfWeek] = parts

		// Check if it's a custom pattern first
		if (
			!(
				// Minutes pattern
				(
					(minute === "*" &&
						hour === "*" &&
						dayOfMonth === "*" &&
						month === "*" &&
						dayOfWeek === "*") ||
					// Hours pattern
					(minute === "0" &&
						hour === "*" &&
						dayOfMonth === "*" &&
						month === "*" &&
						dayOfWeek === "*") ||
					// Days pattern
					(minute === "0" &&
						/^\d+$/.test(hour) &&
						dayOfMonth === "*" &&
						month === "*" &&
						dayOfWeek === "*") ||
					// Weeks pattern
					(minute === "0" &&
						/^\d+$/.test(hour) &&
						dayOfMonth === "*" &&
						month === "*" &&
						/^[0-6]$/.test(dayOfWeek))
				)
			)
		) {
			return { frequency: "custom", customCronExpression: cronExpression }
		}

		// Determine frequency and set states based on cron pattern
		if (minute === "*" && hour === "*") {
			return { frequency: "minutes" }
		}

		if (minute === "0" && hour === "*") {
			return { frequency: "hours" }
		}

		if (
			minute === "0" &&
			dayOfMonth === "*" &&
			month === "*" &&
			dayOfWeek === "*"
		) {
			const hourNum = parseInt(hour)
			return {
				frequency: "days",
				selectedTime:
					hourNum > 12
						? (hourNum - 12).toString()
						: hourNum === 0
							? "12"
							: hourNum.toString(),
				selectedAmPm: hourNum >= 12 ? "PM" : "AM",
			}
		}

		if (
			minute === "0" &&
			dayOfMonth === "*" &&
			month === "*" &&
			/^[0-6]$/.test(dayOfWeek)
		) {
			const hourNum = parseInt(hour)
			return {
				frequency: "weeks",
				selectedTime:
					hourNum > 12
						? (hourNum - 12).toString()
						: hourNum === 0
							? "12"
							: hourNum.toString(),
				selectedAmPm: hourNum >= 12 ? "PM" : "AM",
				selectedDay: DAYS[parseInt(dayOfWeek)],
			}
		}

		return { frequency: "custom", customCronExpression: cronExpression }
	} catch (error) {
		console.error("Error parsing cron expression:", error)
		return { frequency: "custom", customCronExpression: cronExpression }
	}
}

export const validateCronExpression = (cronExpression: string): boolean => {
	if (!cronExpression.trim()) {
		message.error("Cron expression is required")
		return false
	}
	if (!isValidCronExpression(cronExpression)) {
		message.error("Invalid cron expression")
		return false
	}
	return true
}

export type AbortableFunction<T> = (signal: AbortSignal) => Promise<T>

// used to cancel old requests when new one is made which helps in removing the old data
export const withAbortController = <T>(
	fn: AbortableFunction<T>,
	onSuccess: (data: T) => void,
	onError?: (error: unknown) => void,
	onFinally?: () => void,
) => {
	let isMounted = true
	const abortController = new AbortController()

	const execute = async () => {
		try {
			const response = await fn(abortController.signal)
			if (isMounted) {
				onSuccess(response)
			}
		} catch (error: unknown) {
			if (isMounted && error instanceof Error && error.name !== "AbortError") {
				if (onError) {
					onError(error)
				} else {
					console.error("Error in abortable function:", error)
				}
			}
		} finally {
			if (isMounted && onFinally) {
				onFinally()
			}
		}
	}

	execute()

	return () => {
		isMounted = false
		abortController.abort()
		if (onFinally) {
			onFinally()
		}
	}
}

// for small screen items shown will be 6 else 8
export const getResponsivePageSize = () => {
	const screenHeight = window.innerHeight
	return screenHeight >= 900 ? 8 : 6
}

// validate alphanumeric underscore
export const validateAlphanumericUnderscore = (
	value: string,
): { validValue: string; errorMessage: string } => {
	const validValue = value.replace(/[^a-z0-9_]/g, "")
	return {
		validValue,
		errorMessage:
			validValue !== value
				? "Only lowercase letters, numbers and underscores allowed"
				: "",
	}
}

// restricts input to only numbers and control keys
export const restrictNumericInput = (
	event: React.KeyboardEvent<HTMLInputElement>,
) => {
	const allowedKeys = [
		"Backspace",
		"Delete",
		"ArrowLeft",
		"ArrowRight",
		"Tab",
		"Home",
		"End",
	]

	if (!/[0-9]/.test(event.key) && !allowedKeys.includes(event.key)) {
		event.preventDefault()
	}
}

export const handleSpecResponse = (
	response: any,
	setSchema: (schema: any) => void,
	setUiSchema: (uiSchema: any) => void,
	errorType: "source" | "destination" = "source",
) => {
	try {
		if (response?.spec?.jsonschema) {
			setSchema(response.spec.jsonschema)
			setUiSchema(JSON.parse(response.spec.uischema))
		} else {
			console.error(`Failed to get ${errorType} spec:`, response.message)
		}
	} catch {
		setSchema({})
		setUiSchema({})
	}
}

// Casts a single filter condition's value to the appropriate native type
// (number, boolean, object, array) based on the column's type schema.
export const castFilterConditionValue = (
	cond: FilterConfigCondition,
	columnSchema?: { type: string | string[] },
): FilterConfigCondition => {
	if (cond.value === null || cond.value === "<null>") {
		return { ...cond, value: null }
	}

	if (!columnSchema) return cond

	// Find primary non-null type
	const nonNullTypes = (
		Array.isArray(columnSchema.type) ? columnSchema.type : [columnSchema.type]
	).filter(t => t !== "null")

	if (nonNullTypes.length === 0) return cond

	const type = nonNullTypes[0] // take the primary type for casting
	let castValue: any = String(cond.value).trim()

	switch (type) {
		case "integer_small":
		case "integer":
			castValue = castValue === "" ? null : parseInt(castValue, 10)
			break
		case "number_small":
		case "number":
			castValue = castValue === "" ? null : parseFloat(castValue)
			break
		case "boolean":
			castValue = castValue === "true"
			break
		// arrays and objects are sent as string
	}

	return { ...cond, value: castValue }
}

// Filters out disabled streams
export const getSelectedStreams = (selectedStreams: {
	[key: string]: SelectedStream[]
}): { [key: string]: SelectedStream[] } => {
	const result: { [key: string]: SelectedStream[] } = {}

	Object.keys(selectedStreams).forEach(key => {
		result[key] = selectedStreams[key].filter(stream => !stream.disabled)
	})

	return result
}

// Applies type casting to filter_config values for their correct native types
export const formatSelectedStreamsPayload = (
	selectedStreams: { [key: string]: SelectedStream[] },
	streams?: StreamData[],
): { [key: string]: SelectedStream[] } => {
	const filteredStreams = getSelectedStreams(selectedStreams)

	const typeSchemaByName = new Map(
		streams?.map(s => [s.stream.name, s.stream.type_schema?.properties]) ?? [],
	)

	return Object.fromEntries(
		Object.entries(filteredStreams).map(([key, nsStreams]) => [
			key,
			nsStreams.map(stream => {
				const typeSchemaProps = typeSchemaByName.get(stream.stream_name)
				if (!stream.filter_config || !typeSchemaProps) return stream

				return {
					...stream,
					// Cast each condition's value to its schema-defined native type
					filter_config: {
						...stream.filter_config,
						conditions: stream.filter_config.conditions.map(cond =>
							castFilterConditionValue(cond, typeSchemaProps[cond.column]),
						),
					},
				}
			}),
		]),
	)
}

// validates filter expression
export const validateFilter = (filter: string): boolean => {
	if (!filter.trim()) return false
	return FILTER_REGEX.test(filter.trim())
}

// ISO 8601 validation
function isValidISO8601(str: string): boolean {
	const iso8601Regex =
		/^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}(?:\.\d+)?(?:Z|[+-]\d{2}:\d{2})$/
	if (!iso8601Regex.test(str)) return false

	const date = new Date(str)
	return !isNaN(date.getTime())
}

// Validates if value is compatible with any given DataType.
// Explicitly handles "null" values by checking if the column schema allows it.
// Converts other values to string internally to handle native types safely.
export const isValueValidForTypes = (
	value: any,
	type: string | string[],
): boolean => {
	const typeArray = Array.isArray(type) ? type : [type]
	if (typeArray.length === 0) return false

	const v = value === null ? "" : String(value)

	return typeArray.some(t => {
		switch (t) {
			case "null":
				return value === null || v === "<null>"

			case "integer_small":
			case "integer":
				return /^-?\d+$/.test(v) && Number.isInteger(Number(v))

			case "number_small":
			case "number":
				return !isNaN(Number(v)) && v !== ""

			case "boolean":
				return v === "true" || v === "false"

			case "timestamp":
			case "timestamp_milli":
			case "timestamp_micro":
			case "timestamp_nano":
				return isValidISO8601(v)

			case "array": {
				try {
					return Array.isArray(JSON.parse(v))
				} catch {
					return false
				}
			}

			case "object": {
				try {
					const parsed = JSON.parse(v)
					return (
						parsed !== null &&
						typeof parsed === "object" &&
						!Array.isArray(parsed)
					)
				} catch {
					return false
				}
			}

			// "string", "unknown", or any unrecognised type
			default:
				return true
		}
	})
}

// Validates a structured filter_config object.
// Returns null if valid, or a descriptive error string.
export const validateFilterConfig = (
	filterConfig: FilterConfig,
	typeSchemaProperties?: Record<string, { type: string | string[] }>,
): string | null => {
	if (!filterConfig.conditions || filterConfig.conditions.length === 0) {
		return "Filter conditions cannot be empty"
	}

	for (const cond of filterConfig.conditions) {
		if (typeof cond.column !== "string" || cond.column.trim() === "") {
			return "Filter condition is missing a column"
		}
		// Values can be null if the schema allows it, but they cannot be missing/undefined entirely
		if (cond.value === undefined) {
			return `Filter condition for "${cond.column}" is missing a value`
		}

		// Type-aware validation when schema is available
		if (typeSchemaProperties) {
			const columnSchema = typeSchemaProperties[cond.column]
			if (
				columnSchema &&
				!isValueValidForTypes(cond.value, columnSchema.type)
			) {
				const expectedTypes = (
					(Array.isArray(columnSchema.type)
						? columnSchema.type
						: [columnSchema.type]
					).filter(t => t !== "null")[0]
						? (Array.isArray(columnSchema.type)
								? columnSchema.type
								: [columnSchema.type]
							).filter(t => t !== "null")
						: ["null"]
				).join(" | ")
				return `Invalid value "${cond.value}" for column "${cond.column}" — expected type: ${expectedTypes}`
			}
		}
	}

	return null
}

// Returns null if all stream filters are valid, or a descriptive error string.
export const validateStreams = (
	selections: { [key: string]: SelectedStream[] },
	streams?: StreamData[],
): string | null => {
	// Map typeSchemaProperties by stream name for quick lookup
	const typeSchemaByName = new Map(
		streams?.map(s => [s.stream.name, s.stream.type_schema?.properties]) ?? [],
	)

	for (const nsStreams of Object.values(selections)) {
		for (const sel of nsStreams) {
			if (sel.filter && !validateFilter(sel.filter)) {
				return "Invalid filter expression"
			}
			if (sel.filter_config) {
				const typeSchemaProps = typeSchemaByName.get(sel.stream_name)
				const error = validateFilterConfig(sel.filter_config, typeSchemaProps)
				if (error) return error
			}
		}
	}

	return null
}

export const getIngestionMode = (
	selectedStreams: SelectedStreamsByNamespace,
	sourceType?: string,
): IngestionMode => {
	// Fallback to APPEND if source doesn't support UPSERT
	if (!isSourceIngestionModeSupported(IngestionMode.UPSERT, sourceType)) {
		return IngestionMode.APPEND
	}

	const selectedStreamsObj = getSelectedStreams(selectedStreams)
	const allSelectedStreams: SelectedStream[] = []

	// Flatten all streams from all namespaces
	Object.values(selectedStreamsObj).forEach((streams: SelectedStream[]) => {
		allSelectedStreams.push(...streams)
	})

	if (allSelectedStreams.length === 0) return IngestionMode.UPSERT

	const appendCount = allSelectedStreams.filter(
		s => s.append_mode === true,
	).length
	const upsertCount = allSelectedStreams.filter(s => !s.append_mode).length

	if (appendCount === allSelectedStreams.length) return IngestionMode.APPEND
	if (upsertCount === allSelectedStreams.length) return IngestionMode.UPSERT
	return IngestionMode.CUSTOM
}

// Checks if the source connector supports a specific ingestion mode
export const isSourceIngestionModeSupported = (
	mode: IngestionMode,
	sourceType?: string,
): boolean => {
	if (!sourceType) return false

	const normSourceType = normalizeConnectorType(
		sourceType,
	).toLowerCase() as keyof typeof SOURCE_SUPPORTED_INGESTION_MODES
	const sourceModes = SOURCE_SUPPORTED_INGESTION_MODES[normSourceType]

	return sourceModes?.some(m => m === mode) ?? false
}

// Checks if the destination connector supports a specific ingestion mode
export const isDestinationIngestionModeSupported = (
	mode: IngestionMode,
	destinationType?: string,
): boolean => {
	if (!destinationType) return false

	const normDestType = normalizeConnectorType(destinationType).toLowerCase()
	const destModes =
		DESTINATION_SUPPORTED_INGESTION_MODES[
			normDestType as keyof typeof DESTINATION_SUPPORTED_INGESTION_MODES
		]

	return destModes?.some(m => m === mode) ?? false
}

// recursively trims all string values in form data used to remove leading/trailing whitespaces from configuration fields
export const trimFormDataStrings = (data: any): any => {
	if (data === null || data === undefined) {
		return data
	}

	if (typeof data === "string") {
		return data.trim()
	}

	if (Array.isArray(data)) {
		return data.map(item => trimFormDataStrings(item))
	}

	if (typeof data === "object") {
		const trimmedObject: Record<string, any> = {}
		for (const key in data) {
			if (Object.prototype.hasOwnProperty.call(data, key)) {
				trimmedObject[key] = trimFormDataStrings(data[key])
			}
		}
		return trimmedObject
	}

	return data
}

export const getCursorFieldValues = (
	cursorValue?: string,
): CursorFieldValues => {
	if (!cursorValue) {
		return {
			primary: "",
			fallback: "",
		}
	}

	const [primary, fallback] = cursorValue.split(":")

	return {
		primary,
		fallback: fallback || "",
	}
}

// Parses a date string into a timestamp (ms since epoch); handles ISO and legacy formats; returns null if parsing fails
export const parseDateToTimestamp = (timeStr: string): number | null => {
	if (!timeStr) {
		return null
	}

	const timestamp = new Date(timeStr).getTime()
	return isNaN(timestamp) ? null : timestamp
}

// Copies text to clipboard with modern API and fallback support
export async function copyToClipboard(textToCopy: string): Promise<void> {
	// Check if there's content to copy
	if (!textToCopy) {
		message.error("No content provided to copy.")
		console.error("Attempted to copy empty or null text.")
		return
	}

	// Try modern Clipboard API first
	try {
		if (navigator?.clipboard?.writeText) {
			await navigator.clipboard.writeText(textToCopy)
			message.success("Logs copied to clipboard!")
			return
		}

		// Throw to use fallback for HTTP/non-secure contexts
		throw new Error("Clipboard API not available or permitted")
	} catch (err) {
		console.warn(
			"Clipboard API failed, falling back to document.execCommand:",
			err,
		)

		// Fallback: use execCommand with temporary textarea
		try {
			const textarea = document.createElement("textarea")
			textarea.value = textToCopy
			textarea.setAttribute("readonly", "")
			textarea.style.position = "fixed"
			textarea.style.left = "-9999px"
			document.body.appendChild(textarea)
			textarea.select()
			const success = document.execCommand("copy")
			document.body.removeChild(textarea)

			if (!success) {
				throw new Error("Fallback copy failed.")
			}

			message.success("Logs copied to clipboard!")
		} catch (fallbackErr) {
			console.error("Failed to copy logs with both methods", fallbackErr)
			message.error("Failed to copy logs")
		}
	}
}

// Format date from ISO string to readable format (e.g., "Jan 17, 2026")
export const formatDate = (dateString: string): string => {
	try {
		const date = new Date(dateString)
		const options: Intl.DateTimeFormatOptions = {
			day: "numeric",
			month: "short",
			year: "numeric",
		}
		return date.toLocaleDateString("en-US", options)
	} catch {
		return dateString
	}
}

/* Processes release data for UI consumption
 * - Converts ISO dates to readable format: "2026-01-17T10:00:00Z" -> "Released on Jan 17, 2026"
 * - Converts kebab-case tags to Title Case: "new-release" -> "New Release"
 *
 * Before: {
 *   olake_ui_worker: { releases: [{ date: "2026-01-17T10:00:00Z", tags: ["new-release"] }] },
 *   ...
 * }
 *
 * After: {
 *   olake_ui_worker: { releases: [{ date: "Released on Jan 17, 2026", tags: ["New Release"] }] },
 *   ...
 * }
 */
export const processReleasesData = (
	releases: ReleasesResponse | null,
): ReleasesResponse | null => {
	if (!releases) {
		return null
	}

	const formatReleaseData = (releaseTypeData?: ReleaseTypeData) => {
		if (!releaseTypeData) {
			return undefined
		}
		return {
			...releaseTypeData,
			releases: releaseTypeData.releases.map(release => ({
				...release,
				date: `Released on ${formatDate(release.date)}`,
				tags: release.tags.map(tag =>
					tag
						.replace(/-/g, " ")
						.split(" ")
						.map(word => word.charAt(0).toUpperCase() + word.slice(1))
						.join(" "),
				),
			})),
		}
	}
	return {
		[ReleaseType.OLAKE_UI_WORKER]: formatReleaseData(
			releases[ReleaseType.OLAKE_UI_WORKER],
		),
		[ReleaseType.OLAKE_HELM]: formatReleaseData(
			releases[ReleaseType.OLAKE_HELM],
		),
		[ReleaseType.OLAKE]: formatReleaseData(releases[ReleaseType.OLAKE]),
		[ReleaseType.FEATURES]: formatReleaseData(releases[ReleaseType.FEATURES]),
	}
}

import { message } from "antd"
import parser from "cron-parser"

import {
	CronParseResult,
	JobType,
	IngestionMode,
	SelectedStream,
} from "../types"
import {
	DAYS_MAP,
	DESTINATION_INTERNAL_TYPES,
	DESTINATION_LABELS,
	FILTER_REGEX,
} from "./constants"
import MongoDB from "../assets/Mongo.svg"
import Postgres from "../assets/Postgres.svg"
import MySQL from "../assets/MySQL.svg"
import Oracle from "../assets/Oracle.svg"
import AWSS3 from "../assets/AWSS3.svg"
import ApacheIceBerg from "../assets/ApacheIceBerg.svg"
import Kafka from "../assets/Kafka.svg"

// These are used to show in connector dropdowns
export const getConnectorImage = (connector: string) => {
	const lowerConnector = connector.toLowerCase()

	switch (lowerConnector) {
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

export const getConnectorInLowerCase = (connector: string) => {
	const lowerConnector = connector.toLowerCase()

	switch (lowerConnector) {
		case DESTINATION_INTERNAL_TYPES.S3:
		case DESTINATION_LABELS.AMAZON_S3:
			return DESTINATION_INTERNAL_TYPES.S3
		case DESTINATION_INTERNAL_TYPES.ICEBERG:
		case DESTINATION_LABELS.APACHE_ICEBERG:
			return DESTINATION_INTERNAL_TYPES.ICEBERG
		case "mongodb":
			return "mongodb"
		case "postgres":
			return "postgres"
		case "mysql":
			return "mysql"
		case "oracle":
			return "oracle"
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

// Returns a copy of the selected streams map with all disabled streams removed
export const getSelectedStreams = (selectedStreams: {
	[key: string]: SelectedStream[]
}): { [key: string]: SelectedStream[] } => {
	return Object.fromEntries(
		Object.entries(selectedStreams).map(([key, streams]) => [
			key,
			streams.filter(stream => !stream.disabled),
		]),
	)
}

// validates filter expression
export const validateFilter = (filter: string): boolean => {
	if (!filter.trim()) return false
	return FILTER_REGEX.test(filter.trim())
}

export const validateStreams = (selections: {
	[key: string]: SelectedStream[]
}): boolean => {
	return !Object.values(selections).some(streams =>
		streams.some(sel => sel.filter && !validateFilter(sel.filter)),
	)
}

export const getIngestionMode = (selectedStreams: {
	[key: string]: SelectedStream[]
}): IngestionMode => {
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

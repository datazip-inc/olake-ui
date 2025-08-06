import MongoDB from "../assets/Mongo.svg"
import Postgres from "../assets/Postgres.svg"
import MySQL from "../assets/MySQL.svg"
import Oracle from "../assets/Oracle.svg"
import AWSS3 from "../assets/AWSS3.svg"
import ApacheIceBerg from "../assets/ApacheIceBerg.svg"
import { DAYS_MAP } from "./constants"
import { CronParseResult } from "../types"
import parser from "cron-parser"
import { message } from "antd"

export const getConnectorImage = (connector: string) => {
	const lowerConnector = connector.toLowerCase()

	if (lowerConnector === "mongodb") {
		return MongoDB
	} else if (lowerConnector === "postgres") {
		return Postgres
	} else if (lowerConnector === "mysql") {
		return MySQL
	} else if (lowerConnector === "oracle") {
		return Oracle
	} else if (lowerConnector === "s3" || lowerConnector === "amazon") {
		return AWSS3
	} else if (
		lowerConnector === "iceberg" ||
		lowerConnector === "apache iceberg"
	) {
		return ApacheIceBerg
	}

	// Default placeholder
	return MongoDB
}

export const getConnectorName = (connector: string, catalog: string | null) => {
	if (connector === "Amazon S3") {
		return "s3/config"
	} else if (connector === "Apache Iceberg") {
		if (catalog === "AWS Glue") {
			return "iceberg/catalog/glue"
		} else if (catalog === "REST Catalog") {
			return "iceberg/catalog/rest"
		} else if (catalog === "JDBC Catalog") {
			return "iceberg/catalog/jdbc"
		} else if (catalog === "Hive Catalog" || catalog === "HIVE Catalog") {
			return "iceberg/catalog/hive"
		}
	}
}

export const getStatusClass = (status: string) => {
	switch (status.toLowerCase()) {
		case "success":
		case "completed":
			return "text-[#52C41A] bg-[#F6FFED]"
		case "failed":
		case "cancelled":
			return "text-[#F5222D] bg-[#FFF1F0]"
		case "running":
			return "text-primary-700 bg-primary-200"
		case "scheduled":
			return "text-[rgba(0,0,0,88)] bg-neutral-light"
		default:
			return "text-[rgba(0,0,0,88)] bg-transparent"
	}
}

export const getConnectorInLowerCase = (connector: string) => {
	if (connector === "Amazon S3" || connector === "s3") {
		return "s3"
	} else if (connector === "Apache Iceberg" || connector === "iceberg") {
		return "iceberg"
	} else if (connector.toLowerCase() === "mongodb") {
		return "mongodb"
	} else if (connector.toLowerCase() === "postgres") {
		return "postgres"
	} else if (connector.toLowerCase() === "mysql") {
		return "mysql"
	} else if (connector.toLowerCase() === "oracle") {
		return "oracle"
	} else {
		return connector.toLowerCase()
	}
}

export const getCatalogInLowerCase = (catalog: string) => {
	if (catalog === "AWS Glue" || catalog === "glue") {
		return "glue"
	} else if (catalog === "REST Catalog" || catalog === "rest") {
		return "rest"
	} else if (catalog === "JDBC Catalog" || catalog === "jdbc") {
		return "jdbc"
	} else if (catalog === "Hive Catalog" || catalog === "hive") {
		return "hive"
	}
}

export const getStatusLabel = (status: string) => {
	switch (status) {
		case "success":
			return "Success"
		case "failed":
			return "Failed"
		case "cancelled":
			return "Cancelled"
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
		default:
			return "MongoDB"
	}
}

export const getFrequencyValue = (frequency: string) => {
	if (frequency.includes(" ")) {
		const parts = frequency.split(" ")
		const unit = parts[1].toLowerCase()

		if (unit.includes("hour")) return "hours"
		if (unit.includes("minute")) return "minutes"
		if (unit.includes("day")) return "days"
		if (unit.includes("week")) return "weeks"
		if (unit.includes("month")) return "months"
		if (unit.includes("year")) return "years"
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

export const getCatalogName = (catalogType: string) => {
	switch (catalogType?.toLowerCase()) {
		case "glue":
		case "aws glue":
			return "AWS Glue"
		case "rest":
		case "rest catalog":
			return "REST Catalog"
		case "jdbc":
		case "jdbc catalog":
			return "JDBC Catalog"
		case "hive":
		case "hive catalog":
			return "Hive Catalog"
		default:
			return null
	}
}

export const getDestinationType = (type: string) => {
	if (type.toLowerCase() === "amazon s3" || type.toLowerCase() === "s3") {
		return "PARQUET"
	} else if (
		type.toLowerCase() === "apache iceberg" ||
		type.toLowerCase() === "iceberg"
	) {
		return "ICEBERG"
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

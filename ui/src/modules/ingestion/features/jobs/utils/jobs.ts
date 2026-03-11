import { DAYS_MAP } from "../constants"
import { JobType } from "../types"
import { message } from "antd"
import { CronParseResult } from "@/modules/ingestion/features/jobs/types"
import { getConnectorInLowerCase } from "@/modules/ingestion/common/utils"
import parser from "cron-parser"

export const buildConnectorPayload = (
	entity: {
		type: string
		version?: string
		config?: string | object
	} | null,
) => ({
	type: entity ? getConnectorInLowerCase(entity.type) : "",
	version: entity?.version ?? "",
	config: entity?.config
		? typeof entity.config === "string"
			? entity.config
			: JSON.stringify(entity.config)
		: "{}",
})

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

// removes the saved job from local storage when user deletes the job or completes entire flow and create
export const removeSavedJobFromLocalStorage = (jobId: string) => {
	const savedJobs = localStorage.getItem("savedJobs")
	if (savedJobs) {
		const jobs = JSON.parse(savedJobs)
		const filteredJobs = jobs.filter((job: any) => job.id !== jobId)
		localStorage.setItem("savedJobs", JSON.stringify(filteredJobs))
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

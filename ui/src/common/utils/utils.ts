import React from "react"
import { message } from "antd"

import {
	ReleasesResponse,
	ReleaseType,
	ReleaseTypeData,
} from "../types/platformTypes"
import { DESTINATION_INTERNAL_TYPES, DESTINATION_LABELS } from "../constants"
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
} from "@/assets"
import { formatDate } from "@/utils"

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

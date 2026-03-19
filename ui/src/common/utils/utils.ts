import { message } from "antd"
import React from "react"

import { SpecResponse } from "@/common/types"

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
			return "text-[#000000]"
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
	response: SpecResponse,
	setSchema: (schema: object) => void,
	setUiSchema: (uiSchema: object) => void,
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

// Parses a date string into a timestamp (ms since epoch); handles ISO and legacy formats; returns null if parsing fails
export const parseDateToTimestamp = (timeStr: string): number | null => {
	if (!timeStr) {
		return null
	}

	const timestamp = new Date(timeStr).getTime()
	return isNaN(timestamp) ? null : timestamp
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

// Format epoch milliseconds to UTC date-time string: YYYY-MM-DD HH:mm:ss
const getUtcIsoString = (timestamp: number): string | null => {
	const date = new Date(timestamp)
	return Number.isNaN(date.getTime()) ? null : date.toISOString()
}

export const formatTimestampToUtcDateTime = (timestamp: number): string => {
	const isoString = getUtcIsoString(timestamp)
	if (!isoString) return "--"
	return isoString.slice(0, 19).replace("T", " ")
}

// Format epoch milliseconds to UTC time string: HH:mm:ss
export const formatTimestampToUtcTime = (timestamp: number): string => {
	const isoString = getUtcIsoString(timestamp)
	if (!isoString) return "--"
	return isoString.slice(11, 19)
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

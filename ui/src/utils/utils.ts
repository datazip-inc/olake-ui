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

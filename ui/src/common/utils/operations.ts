import { AxiosError } from "axios"

import { operationsService } from "@/common/services/operationsService"
import { OperationAccepted } from "@/common/types"

export const isAbortError = (error: unknown): boolean =>
	(error instanceof DOMException && error.name === "AbortError") ||
	(error instanceof AxiosError &&
		(error.code === "ERR_CANCELED" || error.message === "canceled"))

const toError = (error: unknown, fallback: string): Error => {
	if (error instanceof AxiosError) {
		return new Error(error.response?.data?.message ?? error.message ?? fallback)
	}
	return new Error(error instanceof Error ? error.message : fallback)
}

// The server holds each status poll for up to 10s and answers as soon as the operation
// closes, so this loop needs no delay of its own.
export async function awaitOperation<T>(
	operationId: string,
	signal?: AbortSignal,
): Promise<T> {
	for (;;) {
		if (signal?.aborted) throw new DOMException("Aborted", "AbortError")

		let status
		try {
			status = await operationsService.getStatus(operationId, signal)
		} catch (error) {
			if (isAbortError(error)) throw error
			throw toError(error, "Failed to check operation status")
		}

		if (status.status === "succeeded") {
			try {
				return await operationsService.getResult<T>(operationId, signal)
			} catch (error) {
				if (isAbortError(error)) throw error
				throw toError(error, "Failed to fetch operation result")
			}
		}
		if (status.status === "failed") {
			throw new Error(status.error || "Operation failed")
		}
	}
}

export async function runOperation<T>(
	start: () => Promise<OperationAccepted>,
	signal?: AbortSignal,
): Promise<T> {
	const accepted = await start()
	return awaitOperation<T>(accepted.operation_id, signal)
}

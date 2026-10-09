import { AxiosError } from "axios"

import { API_CONFIG } from "@/config"
import { api } from "@/core/api"

/**
 * Client for the async operations API.
 *
 * Work that runs a connector container is submitted, acknowledged with an operation id,
 * and then polled. No request here stays open for the length of the work: the server
 * holds a status poll for at most POLL_WAIT_SECONDS and answers the instant the operation
 * finishes. That is what keeps these flows working behind corporate proxies and VPNs,
 * which cut idle connections long before a discovery or image pull completes.
 */

export type OperationStatusValue =
	| "running"
	| "succeeded"
	| "failed"
	| "canceled"
	| "timed_out"

export interface OperationAccepted {
	operation_id: string
	kind: string
	status: OperationStatusValue
}

export interface OperationStatus {
	operation_id: string
	kind: string
	status: OperationStatusValue
	started_at?: string
	finished_at?: string
	error?: string
}

/** Must stay below the server's OPERATION_MAX_WAIT and any proxy idle timeout. */
const POLL_WAIT_SECONDS = 10
/** Request timeout with headroom over the server-side hold. */
const POLL_REQUEST_TIMEOUT_MS = 30_000
/** Floor between polls, in case the server answers immediately. */
const MIN_POLL_INTERVAL_MS = 1_000

const PAYLOAD_TRANSFER_TIMEOUT_MS = 0

const operationsURL = () =>
	API_CONFIG.ENDPOINTS.ETL.OPERATIONS(API_CONFIG.PROJECT_ID)

const operationURL = (operationId: string) =>
	`${operationsURL()}/${encodeURIComponent(operationId)}`

export const isAbortError = (error: unknown): boolean =>
	(error instanceof DOMException && error.name === "AbortError") ||
	(error instanceof AxiosError &&
		(error.code === "ERR_CANCELED" || error.message === "canceled"))

/**
 * Failure of an operation, or of a poll against it. A distinct type so callers can tell it
 * apart from an ordinary AxiosError and pull the message out of the right place.
 */
export class OperationError extends Error {
	constructor(message: string) {
		super(message)
		this.name = "OperationError"
	}
}

const asOperationError = (error: unknown, fallback: string): OperationError =>
	new OperationError(errorMessage(error, fallback))

const delay = (ms: number, signal?: AbortSignal) =>
	new Promise<void>((resolve, reject) => {
		const timer = setTimeout(() => {
			signal?.removeEventListener("abort", onAbort)
			resolve()
		}, ms)
		const onAbort = () => {
			clearTimeout(timer)
			reject(new DOMException("Aborted", "AbortError"))
		}
		signal?.addEventListener("abort", onAbort, { once: true })
	})

const errorMessage = (error: unknown, fallback: string): string => {
	if (error instanceof AxiosError) {
		return error.response?.data?.message ?? error.message ?? fallback
	}
	return error instanceof Error ? error.message : fallback
}

export const operationsService = {
	getStatus: async (
		operationId: string,
		signal?: AbortSignal,
	): Promise<OperationStatus> => {
		const response = await api.get<OperationStatus>(operationURL(operationId), {
			params: { wait: POLL_WAIT_SECONDS },
			timeout: POLL_REQUEST_TIMEOUT_MS,
			signal,
			disableErrorNotification: true,
		})
		return response.data
	},
	getResult: async <T>(
		operationId: string,
		signal?: AbortSignal,
	): Promise<T> => {
		const response = await api.get<T>(`${operationURL(operationId)}/result`, {
			timeout: PAYLOAD_TRANSFER_TIMEOUT_MS,
			signal,
			disableErrorNotification: true,
		})
		return response.data
	},

	/**
	 * Stops the operation and the container behind it. Not wired to unmount — only an
	 * explicit user action should kill work that is already running.
	 */
	cancel: async (operationId: string): Promise<void> => {
		try {
			await api.delete(operationURL(operationId), {
				disableErrorNotification: true,
			})
		} catch (error) {
			// Best effort: the operation may already have finished on its own.
			console.error("Error cancelling operation:", error)
		}
	},
}

interface AwaitOptions {
	signal?: AbortSignal
	onStatus?: (status: OperationStatus) => void
}

/** Polls an operation to completion and returns its result. */
export async function awaitOperation<T>(
	operationId: string,
	{ signal, onStatus }: AwaitOptions = {},
): Promise<T> {
	for (;;) {
		if (signal?.aborted) throw new DOMException("Aborted", "AbortError")

		const startedAt = Date.now()
		let status: OperationStatus
		try {
			status = await operationsService.getStatus(operationId, signal)
		} catch (error) {
			if (isAbortError(error)) throw error
			throw asOperationError(error, "Failed to check operation status")
		}

		onStatus?.(status)

		if (status.status === "succeeded") {
			try {
				return await operationsService.getResult<T>(operationId, signal)
			} catch (error) {
				if (isAbortError(error)) throw error
				throw asOperationError(error, "Failed to fetch operation result")
			}
		}
		if (status.status !== "running") {
			// The operation itself ended badly — a failed image pull, a timeout, a
			// cancellation. `status.error` carries the reason the server recovered.
			throw new OperationError(status.error || `Operation ${status.status}`)
		}

		const elapsed = Date.now() - startedAt
		if (elapsed < MIN_POLL_INTERVAL_MS) {
			await delay(MIN_POLL_INTERVAL_MS - elapsed, signal)
		}
	}
}

/** Submits work, then polls it to completion. */
export async function runOperation<T>(
	start: () => Promise<OperationAccepted>,
	options: AwaitOptions = {},
): Promise<T> {
	const accepted = await start()
	return awaitOperation<T>(accepted.operation_id, options)
}

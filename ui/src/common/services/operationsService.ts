import {
	PAYLOAD_TRANSFER_TIMEOUT_MS,
	POLL_REQUEST_TIMEOUT_MS,
	POLL_WAIT_SECONDS,
} from "@/common/constants"
import { OperationStatus } from "@/common/types"
import { API_CONFIG } from "@/config"
import { api } from "@/core/api"

const operationURL = (operationId: string) =>
	`${API_CONFIG.ENDPOINTS.ETL.OPERATIONS(API_CONFIG.PROJECT_ID)}/${encodeURIComponent(operationId)}`

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
}

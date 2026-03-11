import { useMutation } from "@tanstack/react-query"
import { sourceService } from "../../services"
import { sourceKeys } from "../../constants/queryKeys"
import type {
	EntityBase,
	EntityTestRequest,
} from "@/modules/ingestion/common/types"

export const useCreateSource = () => {
	return useMutation({
		mutationKey: sourceKeys.lists(),
		mutationFn: (source: EntityBase) => sourceService.createSource(source),
	})
}

export const useUpdateSource = (id: string) => {
	return useMutation({
		mutationKey: sourceKeys.detail(id),
		mutationFn: (source: EntityBase) => sourceService.updateSource(id, source),
	})
}

export const useDeleteSource = () => {
	return useMutation({
		mutationKey: sourceKeys.lists(),
		mutationFn: (id: string) => sourceService.deleteSource(id),
	})
}

export const useTestSourceConnection = () => {
	return useMutation({
		mutationFn: ({
			source,
			existing = false,
		}: {
			source: EntityTestRequest
			existing?: boolean
		}) => sourceService.testSourceConnection(source, existing),
	})
}

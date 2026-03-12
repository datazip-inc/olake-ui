import { useMutation } from "@tanstack/react-query"

import type {
	EntityBase,
	EntityTestRequest,
} from "@/modules/ingestion/common/types"

import { destinationKeys } from "../../constants/queryKeys"
import { destinationService } from "../../services"

export const useCreateDestination = () => {
	return useMutation({
		mutationKey: destinationKeys.lists(),
		mutationFn: (destination: Omit<EntityBase, "id" | "createdAt">) =>
			destinationService.createDestination(destination),
	})
}

export const useUpdateDestination = (id: string) => {
	return useMutation({
		mutationKey: destinationKeys.lists(),
		mutationFn: (destination: EntityBase) =>
			destinationService.updateDestination(id, destination),
	})
}

export const useDeleteDestination = () => {
	return useMutation({
		mutationKey: destinationKeys.lists(),
		mutationFn: (id: string) => {
			const numericId = typeof id === "string" ? parseInt(id, 10) : id
			return destinationService.deleteDestination(numericId)
		},
	})
}

export const useTestDestinationConnection = () => {
	return useMutation({
		mutationFn: ({
			destination,
			existing = false,
			sourceType = "",
			sourceVersion = "",
		}: {
			destination: EntityTestRequest
			existing?: boolean
			sourceType?: string
			sourceVersion?: string
		}) =>
			destinationService.testDestinationConnection(
				destination,
				existing,
				sourceType,
				sourceVersion,
			),
	})
}

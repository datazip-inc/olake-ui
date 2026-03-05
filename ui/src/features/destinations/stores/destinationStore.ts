import { create } from "zustand"
import type { Entity, TestConnectionError } from "@/common/types"

// Client-side UI state only. Server state is handled by TanStack Query.
export interface DestinationState {
	selectedDestination: Entity
	setSelectedDestination: (destination: Entity) => void
	destinationTestConnectionError: TestConnectionError | null
	setDestinationTestConnectionError: (error: TestConnectionError | null) => void
}

export const useDestinationStore = create<DestinationState>()(set => ({
	selectedDestination: {} as Entity,
	setSelectedDestination: destination =>
		set({ selectedDestination: destination }),
	destinationTestConnectionError: null,
	setDestinationTestConnectionError: error =>
		set({ destinationTestConnectionError: error }),
}))

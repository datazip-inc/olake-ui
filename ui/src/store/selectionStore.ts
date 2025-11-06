import { StateCreator } from "zustand"
import type { Entity } from "../types"
import { jobService } from "../api"
import { ModalSlice } from "./modalStore"
export interface SelectionSlice {
	selectedJobId: string | null
	selectedHistoryId: string | null
	selectedSource: Entity
	selectedDestination: Entity
	selectedClearDestinationRunning: boolean
	isClearDestinationStatusLoading: boolean
	setSelectedJobId: (id: string | null) => void
	setSelectedHistoryId: (id: string | null) => void
	setSelectedSource: (source: Entity) => void
	setSelectedDestination: (destination: Entity) => void
	fetchSelectedClearDestinationStatus: () => void
}

export const createSelectionSlice: StateCreator<
	SelectionSlice & ModalSlice,
	[],
	[],
	SelectionSlice
> = (set, get) => ({
	selectedJobId: null,
	selectedHistoryId: null,
	selectedSource: {} as Entity,
	selectedDestination: {} as Entity,
	selectedClearDestinationRunning: false,
	isClearDestinationStatusLoading: false,
	setSelectedJobId: id => set({ selectedJobId: id }),
	setSelectedHistoryId: id => set({ selectedHistoryId: id }),
	setSelectedSource: source => set({ selectedSource: source }),
	setSelectedDestination: destination =>
		set({ selectedDestination: destination }),
	fetchSelectedClearDestinationStatus: async () => {
		const jobId = get().selectedJobId

		if (!jobId) return

		try {
			set({
				isClearDestinationStatusLoading: true,
			})
			const status = await jobService.getClearDestinationStatus(jobId)

			if (status.running) {
				set({ showStreamEditDisabledModal: true })
			}

			set({
				selectedClearDestinationRunning: status.running,
				isClearDestinationStatusLoading: false,
			})
		} catch (error) {
			console.error("Error fetching clear destination status:", error)
			set({
				isClearDestinationStatusLoading: false,
			})
		}
	},
})

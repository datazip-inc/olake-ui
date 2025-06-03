import { StateCreator } from "zustand"
import type { Entity } from '../types'

export interface SelectionSlice {
	selectedJobId: string | null
	selectedHistoryId: string | null
	selectedSource: Entity
	selectedDestination: Entity
	setSelectedJobId: (id: string | null) => void
	setSelectedHistoryId: (id: string | null) => void
	setSelectedSource: (source: Entity) => void
	setSelectedDestination: (destination: Entity) => void
}

export const createSelectionSlice: StateCreator<SelectionSlice> = set => ({
	selectedJobId: null,
	selectedHistoryId: null,
	selectedSource: {} as Entity,
	selectedDestination: {} as Entity,
	setSelectedJobId: id => set({ selectedJobId: id }),
	setSelectedHistoryId: id => set({ selectedHistoryId: id }),
	setSelectedSource: source => set({ selectedSource: source }),
	setSelectedDestination: destination =>
    set({ selectedDestination: destination }),
})

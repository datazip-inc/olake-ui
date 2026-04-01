import { create } from "zustand"

import type { TestConnectionError } from "@/common/types"
import type { Entity } from "@/modules/ingestion/common/types"

// Client-side UI state only. Server state is handled by TanStack Query.
export interface SourceState {
	selectedSource: Entity
	setSelectedSource: (source: Entity) => void
	sourceTestConnectionError: TestConnectionError | null
	setSourceTestConnectionError: (error: TestConnectionError | null) => void
}

export const useSourceStore = create<SourceState>()(set => ({
	selectedSource: {} as Entity,
	setSelectedSource: source => set({ selectedSource: source }),
	sourceTestConnectionError: null,
	setSourceTestConnectionError: error =>
		set({ sourceTestConnectionError: error }),
}))

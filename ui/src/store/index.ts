import { create } from "zustand"
import { AuthSlice, createAuthSlice } from "./authStore"
import { JobSlice, createJobSlice } from "./jobStore"
import { SourceSlice, createSourceSlice } from "./sourceStore"
import { DestinationSlice, createDestinationSlice } from "./destinationStore"
import { TaskSlice, createTaskSlice } from "./taskStore"
import { ModalSlice, createModalSlice } from "./modalStore"
import { SelectionSlice, createSelectionSlice } from "./selectionStore"
import { createSettingsSlice, SettingsSlice } from "./settingsStore"

export type AppState = AuthSlice &
	JobSlice &
	SourceSlice &
	DestinationSlice &
	TaskSlice &
	ModalSlice &
	SelectionSlice &
	SettingsSlice

export const useAppStore = create<AppState>()((...a) => ({
	...createAuthSlice(...a),
	...createJobSlice(...a),
	...createSourceSlice(...a),
	...createDestinationSlice(...a),
	...createTaskSlice(...a),
	...createModalSlice(...a),
	...createSelectionSlice(...a),
	...createSettingsSlice(...a),
}))

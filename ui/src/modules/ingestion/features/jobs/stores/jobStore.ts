import { create } from "zustand"
import { IngestionMode } from "@/modules/ingestion/features/jobs/enums"

// Client-side UI state only. Server state is handled by TanStack Query.
interface JobUIState {
	// Selection
	selectedJobId: string | null
	setSelectedJobId: (id: string | null) => void

	// Modal visibility
	showDeleteJobModal: boolean
	showClearDataModal: boolean
	showClearDestinationModal: boolean
	showStreamDifferenceModal: boolean
	showStreamEditDisabledModal: boolean
	showDestinationDatabaseModal: boolean
	showResetStreamsModal: boolean
	showIngestionModeChangeModal: boolean
	ingestionMode: IngestionMode

	setShowDeleteJobModal: (show: boolean) => void
	setShowClearDataModal: (show: boolean) => void
	setShowClearDestinationModal: (show: boolean) => void
	setShowStreamDifferenceModal: (show: boolean) => void
	setShowStreamEditDisabledModal: (show: boolean) => void
	setShowDestinationDatabaseModal: (show: boolean) => void
	setShowResetStreamsModal: (show: boolean) => void
	setShowIngestionModeChangeModal: (show: boolean) => void
	setIngestionMode: (mode: IngestionMode) => void
}

export const useJobStore = create<JobUIState>()(set => ({
	selectedJobId: null,
	setSelectedJobId: id => set({ selectedJobId: id }),

	showDeleteJobModal: false,
	showClearDataModal: false,
	showClearDestinationModal: false,
	showStreamDifferenceModal: false,
	showStreamEditDisabledModal: false,
	showDestinationDatabaseModal: false,
	showResetStreamsModal: false,
	showIngestionModeChangeModal: false,
	ingestionMode: IngestionMode.UPSERT,

	setShowDeleteJobModal: show => set({ showDeleteJobModal: show }),
	setShowClearDataModal: show => set({ showClearDataModal: show }),
	setShowClearDestinationModal: show =>
		set({ showClearDestinationModal: show }),
	setShowStreamDifferenceModal: show =>
		set({ showStreamDifferenceModal: show }),
	setShowStreamEditDisabledModal: show =>
		set({ showStreamEditDisabledModal: show }),
	setShowDestinationDatabaseModal: show =>
		set({ showDestinationDatabaseModal: show }),
	setShowResetStreamsModal: show => set({ showResetStreamsModal: show }),
	setShowIngestionModeChangeModal: show =>
		set({ showIngestionModeChangeModal: show }),
	setIngestionMode: mode => set({ ingestionMode: mode }),
}))

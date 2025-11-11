import { StateCreator } from "zustand"
import { IngestionMode } from "../types/commonTypes"

export interface ModalSlice {
	showTestingModal: boolean
	showSuccessModal: boolean
	showFailureModal: boolean
	showEntitySavedModal: boolean
	showSourceCancelModal: boolean
	showDeleteModal: boolean
	showDeleteJobModal: boolean
	showClearDataModal: boolean
	showClearDestinationModal: boolean
	showStreamDifferenceModal: boolean
	showStreamEditDisabledModal: boolean
	showEditSourceModal: boolean
	showEditDestinationModal: boolean
	showDestinationDatabaseModal: boolean
	showResetStreamsModal: boolean
	showIngestionModeChangeModal: boolean
	ingestionMode: IngestionMode
	showSpecFailedModal: boolean
	setShowTestingModal: (show: boolean) => void
	setShowSuccessModal: (show: boolean) => void
	setShowFailureModal: (show: boolean) => void
	setShowEntitySavedModal: (show: boolean) => void
	setShowSourceCancelModal: (show: boolean) => void
	setShowDeleteModal: (show: boolean) => void
	setShowDeleteJobModal: (show: boolean) => void
	setShowClearDataModal: (show: boolean) => void
	setShowClearDestinationModal: (show: boolean) => void
	setShowStreamDifferenceModal: (show: boolean) => void
	setShowStreamEditDisabledModal: (show: boolean) => void
	setShowEditSourceModal: (show: boolean) => void
	setShowEditDestinationModal: (show: boolean) => void
	setShowDestinationDatabaseModal: (show: boolean) => void
	setShowResetStreamsModal: (show: boolean) => void
	setShowIngestionModeChangeModal: (show: boolean) => void
	setIngestionMode: (mode: IngestionMode) => void
	setShowSpecFailedModal: (show: boolean) => void
}

export const createModalSlice: StateCreator<ModalSlice> = set => ({
	showTestingModal: false,
	showSuccessModal: false,
	showFailureModal: false,
	showEntitySavedModal: false,
	showSourceCancelModal: false,
	showDeleteModal: false,
	showDeleteJobModal: false,
	showClearDataModal: false,
	showClearDestinationModal: false,
	showStreamDifferenceModal: false,
	showStreamEditDisabledModal: false,
	showEditSourceModal: false,
	showEditDestinationModal: false,
	showDestinationDatabaseModal: false,
	showResetStreamsModal: false,
	showIngestionModeChangeModal: false,
	ingestionMode: IngestionMode.UPSERT,
	showSpecFailedModal: false,
	setShowTestingModal: show => set({ showTestingModal: show }),
	setShowSuccessModal: show => set({ showSuccessModal: show }),
	setShowFailureModal: show => set({ showFailureModal: show }),
	setShowEntitySavedModal: show => set({ showEntitySavedModal: show }),
	setShowSourceCancelModal: show => set({ showSourceCancelModal: show }),
	setShowDeleteModal: show => set({ showDeleteModal: show }),
	setShowDeleteJobModal: show => set({ showDeleteJobModal: show }),
	setShowClearDataModal: show => set({ showClearDataModal: show }),
	setShowClearDestinationModal: show =>
		set({ showClearDestinationModal: show }),
	setShowStreamDifferenceModal: show =>
		set({ showStreamDifferenceModal: show }),
	setShowStreamEditDisabledModal: show =>
		set({ showStreamEditDisabledModal: show }),
	setShowEditSourceModal: show => set({ showEditSourceModal: show }),
	setShowEditDestinationModal: show => set({ showEditDestinationModal: show }),
	setShowDestinationDatabaseModal: show =>
		set({ showDestinationDatabaseModal: show }),
	setShowResetStreamsModal: show => set({ showResetStreamsModal: show }),
	setShowIngestionModeChangeModal: show =>
		set({ showIngestionModeChangeModal: show }),
	setIngestionMode: mode => set({ ingestionMode: mode }),
	setShowSpecFailedModal: show => set({ showSpecFailedModal: show }),
})

import { StateCreator } from "zustand"

export interface ModalSlice {
	showTestingModal: boolean
	showSuccessModal: boolean
	showFailureModal: boolean
	showEntitySavedModal: boolean
	showSourceCancelModal: boolean
	showDeleteModal: boolean
	showDeleteJobModal: boolean
	showClearDataModal: boolean
	showClearDestinationAndSyncModal: boolean
	showEditSourceModal: boolean
	showEditDestinationModal: boolean
	showDestinationDatabaseModal: boolean
	setShowTestingModal: (show: boolean) => void
	setShowSuccessModal: (show: boolean) => void
	setShowFailureModal: (show: boolean) => void
	setShowEntitySavedModal: (show: boolean) => void
	setShowSourceCancelModal: (show: boolean) => void
	setShowDeleteModal: (show: boolean) => void
	setShowDeleteJobModal: (show: boolean) => void
	setShowClearDataModal: (show: boolean) => void
	setShowClearDestinationAndSyncModal: (show: boolean) => void
	setShowEditSourceModal: (show: boolean) => void
	setShowEditDestinationModal: (show: boolean) => void
	setShowDestinationDatabaseModal: (show: boolean) => void
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
	showClearDestinationAndSyncModal: false,
	showEditSourceModal: false,
	showEditDestinationModal: false,
	showDestinationDatabaseModal: false,

	setShowTestingModal: show => set({ showTestingModal: show }),
	setShowSuccessModal: show => set({ showSuccessModal: show }),
	setShowFailureModal: show => set({ showFailureModal: show }),
	setShowEntitySavedModal: show => set({ showEntitySavedModal: show }),
	setShowSourceCancelModal: show => set({ showSourceCancelModal: show }),
	setShowDeleteModal: show => set({ showDeleteModal: show }),
	setShowDeleteJobModal: show => set({ showDeleteJobModal: show }),
	setShowClearDataModal: show => set({ showClearDataModal: show }),
	setShowClearDestinationAndSyncModal: show =>
		set({ showClearDestinationAndSyncModal: show }),
	setShowEditSourceModal: show => set({ showEditSourceModal: show }),
	setShowEditDestinationModal: show => set({ showEditDestinationModal: show }),
	setShowDestinationDatabaseModal: show =>
		set({ showDestinationDatabaseModal: show }),
})

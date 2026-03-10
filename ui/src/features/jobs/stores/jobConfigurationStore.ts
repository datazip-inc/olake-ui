import { create } from "zustand"
import { AdvancedSettings } from "../types"
import { Entity } from "@/common/types"

interface JobConfigurationState {
	jobName: string
	cronExpression: string
	selectedSource: Entity | null
	selectedDestination: Entity | null
	advancedSettings: AdvancedSettings | null
	isEditMode: boolean
}

interface JobConfigurationActions {
	setJobName: (name: string) => void
	setCronExpression: (expr: string) => void
	setSelectedSource: (source: Entity | null) => void
	setSelectedDestination: (destination: Entity | null) => void
	setAdvancedSettings: (settings: AdvancedSettings | null) => void
	setIsEditMode: (isEdit: boolean) => void
	reset: () => void
}

type JobConfigurationStore = JobConfigurationState & JobConfigurationActions

const initialState: JobConfigurationState = {
	jobName: "",
	cronExpression: "* * * * *",
	selectedSource: null,
	selectedDestination: null,
	advancedSettings: null,
	isEditMode: false,
}

export const useJobConfigurationStore = create<JobConfigurationStore>(set => ({
	...initialState,
	setJobName: jobName => set({ jobName }),
	setCronExpression: cronExpression => set({ cronExpression }),
	setSelectedSource: selectedSource => set({ selectedSource }),
	setSelectedDestination: selectedDestination => set({ selectedDestination }),
	setAdvancedSettings: advancedSettings => set({ advancedSettings }),
	setIsEditMode: isEditMode => set({ isEditMode }),
	reset: () => set(initialState),
}))

import { StateCreator } from "zustand"
import type {
	APIResponse,
	Entity,
	EntityBase,
	TestConnectionError,
} from "../types"
import { sourceService } from "../api"
export interface SourceSlice {
	sources: Entity[]
	sourcesError: string | null
	isLoadingSources: boolean
	sourceTestConnectionError: TestConnectionError | null
	setSourceTestConnectionError: (error: TestConnectionError | null) => void
	fetchSources: () => Promise<Entity[]>
	addSource: (source: EntityBase) => Promise<APIResponse<EntityBase>>
	updateSource: (id: string, source: EntityBase) => Promise<APIResponse<Entity>>
	deleteSource: (id: string) => Promise<void>
}

export const createSourceSlice: StateCreator<SourceSlice> = set => ({
	sourceTestConnectionError: null,
	sources: [],
	isLoadingSources: false,
	sourcesError: null,

	setSourceTestConnectionError: error =>
		set({ sourceTestConnectionError: error }),

	fetchSources: async () => {
		set({ isLoadingSources: true, sourcesError: null })
		try {
			const sources = await sourceService.getSources()
			set({ sources, isLoadingSources: false })
			return sources
		} catch (error) {
			set({
				isLoadingSources: false,
				sourcesError:
					error instanceof Error ? error.message : "Failed to fetch sources",
			})
			throw error
		}
	},

	addSource: async sourceData => {
		try {
			const newSource = await sourceService.createSource(sourceData)
			set(state => ({ sources: [...state.sources, newSource.data as Entity] }))
			return newSource
		} catch (error) {
			set({
				sourcesError:
					error instanceof Error ? error.message : "Failed to add source",
			})
			throw error
		}
	},

	updateSource: async (id, sourceData) => {
		try {
			const updatedSource = await sourceService.updateSource(id, sourceData)
			const updatedSourceData = updatedSource.data as Entity

			set(state => ({
				sources: state.sources.map(source =>
					source.id.toString() === id ? updatedSourceData : source,
				),
			}))
			return updatedSource
		} catch (error) {
			set({
				sourcesError:
					error instanceof Error ? error.message : "Failed to update source",
			})
			throw error
		}
	},

	deleteSource: async id => {
		try {
			const numericId = typeof id === "string" ? parseInt(id, 10) : id
			await sourceService.deleteSource(numericId.toString())
			set(state => ({
				sources: state.sources.filter(source => source.id !== numericId),
			}))
		} catch (error) {
			set({
				sourcesError:
					error instanceof Error ? error.message : "Failed to delete source",
			})
			throw error
		}
	},
})

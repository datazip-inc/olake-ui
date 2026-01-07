import { StateCreator } from "zustand"
import type { Entity, EntityBase, TestConnectionError } from "../types"
import { sourceService } from "../api"
export interface SourceSlice {
	sources: Entity[]
	sourcesError: string | null
	isLoadingSources: boolean
	source: Entity | null
	isLoadingSource: boolean
	sourceError: string | null
	sourceTestConnectionError: TestConnectionError | null
	setSourceTestConnectionError: (error: TestConnectionError | null) => void
	fetchSources: () => Promise<Entity[]>
	fetchSource: (id: string) => Promise<void>
	addSource: (source: EntityBase) => Promise<EntityBase>
	updateSource: (id: string, source: EntityBase) => Promise<Entity>
	deleteSource: (id: string) => Promise<void>
}

export const createSourceSlice: StateCreator<SourceSlice> = set => ({
	sourceTestConnectionError: null,
	sources: [],
	isLoadingSources: false,
	sourcesError: null,
	source: null,
	isLoadingSource: false,
	sourceError: null,

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

	fetchSource: async (id: string) => {
		set({ source: null, isLoadingSource: true, sourceError: null })
		try {
			const source = await sourceService.getSource(id)
			set({
				source: source,
				isLoadingSource: false,
			})
		} catch (error) {
			set({
				isLoadingSource: false,
				sourceError:
					error instanceof Error ? error.message : "Failed to fetch source",
			})
			throw error
		}
	},

	addSource: async sourceData => {
		try {
			const newSource = await sourceService.createSource(sourceData)
			set(state => ({ sources: [...state.sources, newSource as Entity] }))
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
			const updatedSourceData = updatedSource as Entity

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

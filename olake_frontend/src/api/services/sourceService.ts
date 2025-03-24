import api from "../axios"
import { Source } from "../../types"
import { mockSources } from "../mockData"

// Flag to use mock data instead of real API
const useMockData = true

export const sourceService = {
	// Get all sources
	getSources: async () => {
		if (useMockData) {
			// Return mock data with a small delay to simulate network request
			return new Promise<Source[]>(resolve => {
				setTimeout(() => resolve(mockSources), 500)
			})
		}

		const response = await api.get<Source[]>("/sources")
		return response.data
	},

	// Get source by id
	getSourceById: async (id: string) => {
		if (useMockData) {
			const source = mockSources.find(source => source.id === id)
			if (!source) throw new Error("Source not found")

			return new Promise<Source>(resolve => {
				setTimeout(() => resolve(source), 300)
			})
		}

		const response = await api.get<Source>(`/sources/${id}`)
		return response.data
	},

	// Create new source
	createSource: async (source: Omit<Source, "id" | "createdAt">) => {
		if (useMockData) {
			const newSource: Source = {
				...source,
				id: Math.random().toString(36).substring(2, 9),
				createdAt: new Date(),
			}

			mockSources.push(newSource)

			return new Promise<Source>(resolve => {
				setTimeout(() => resolve(newSource), 400)
			})
		}

		const response = await api.post<Source>("/sources", source)
		return response.data
	},

	// Update source
	updateSource: async (id: string, source: Partial<Source>) => {
		if (useMockData) {
			const index = mockSources.findIndex(s => s.id === id)
			if (index === -1) throw new Error("Source not found")

			const updatedSource = { ...mockSources[index], ...source }
			mockSources[index] = updatedSource

			return new Promise<Source>(resolve => {
				setTimeout(() => resolve(updatedSource), 300)
			})
		}

		const response = await api.put<Source>(`/sources/${id}`, source)
		return response.data
	},

	// Delete source
	deleteSource: async (id: string) => {
		if (useMockData) {
			const index = mockSources.findIndex(s => s.id === id)
			if (index === -1) throw new Error("Source not found")

			mockSources.splice(index, 1)

			return new Promise<void>(resolve => {
				setTimeout(() => resolve(), 300)
			})
		}

		const response = await api.delete(`/sources/${id}`)
		return response.data
	},

	// Test source connection
	testConnection: async (id: string) => {
		if (useMockData) {
			const source = mockSources.find(s => s.id === id)
			if (!source) throw new Error("Source not found")

			return new Promise<{ success: boolean; message: string }>(resolve => {
				setTimeout(
					() =>
						resolve({
							success: true,
							message: "Connection successful",
						}),
					800,
				)
			})
		}

		const response = await api.post(`/sources/${id}/test`)
		return response.data
	},
}

import api from "../axios"
import { Destination } from "../../types"
import { mockDestinations } from "../mockData"

// Flag to use mock data instead of real API
const useMockData = true

export const destinationService = {
	// Get all destinations
	getDestinations: async () => {
		if (useMockData) {
			// Return mock data with a small delay to simulate network request
			return new Promise<Destination[]>(resolve => {
				setTimeout(() => resolve(mockDestinations), 500)
			})
		}

		const response = await api.get<Destination[]>("/destinations")
		return response.data
	},

	// Get destination by id
	getDestinationById: async (id: string) => {
		if (useMockData) {
			const destination = mockDestinations.find(
				destination => destination.id === id,
			)
			if (!destination) throw new Error("Destination not found")

			return new Promise<Destination>(resolve => {
				setTimeout(() => resolve(destination), 300)
			})
		}

		const response = await api.get<Destination>(`/destinations/${id}`)
		return response.data
	},

	// Create new destination
	createDestination: async (
		destination: Omit<Destination, "id" | "createdAt">,
	) => {
		if (useMockData) {
			const newDestination: Destination = {
				...destination,
				id: Math.random().toString(36).substring(2, 9),
				createdAt: new Date(),
			}

			mockDestinations.push(newDestination)

			return new Promise<Destination>(resolve => {
				setTimeout(() => resolve(newDestination), 400)
			})
		}

		const response = await api.post<Destination>("/destinations", destination)
		return response.data
	},

	// Update destination
	updateDestination: async (id: string, destination: Partial<Destination>) => {
		if (useMockData) {
			const index = mockDestinations.findIndex(d => d.id === id)
			if (index === -1) throw new Error("Destination not found")

			const updatedDestination = { ...mockDestinations[index], ...destination }
			mockDestinations[index] = updatedDestination

			return new Promise<Destination>(resolve => {
				setTimeout(() => resolve(updatedDestination), 300)
			})
		}

		const response = await api.put<Destination>(
			`/destinations/${id}`,
			destination,
		)
		return response.data
	},

	// Delete destination
	deleteDestination: async (id: string) => {
		if (useMockData) {
			const index = mockDestinations.findIndex(d => d.id === id)
			if (index === -1) throw new Error("Destination not found")

			mockDestinations.splice(index, 1)

			return new Promise<void>(resolve => {
				setTimeout(() => resolve(), 300)
			})
		}

		const response = await api.delete(`/destinations/${id}`)
		return response.data
	},

	// Test destination connection
	testConnection: async (id: string) => {
		if (useMockData) {
			const destination = mockDestinations.find(d => d.id === id)
			if (!destination) throw new Error("Destination not found")

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

		const response = await api.post(`/destinations/${id}/test`)
		return response.data
	},
}

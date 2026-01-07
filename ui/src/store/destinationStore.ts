import { StateCreator } from "zustand"
import type { TestConnectionError, EntityBase, Entity } from "../types"
import { destinationService } from "../api"

export interface DestinationSlice {
	destinations: Entity[]
	isLoadingDestinations: boolean
	destinationsError: string | null
	destination: Entity | null
	isLoadingDestination: boolean
	destinationError: string | null
	destinationTestConnectionError: TestConnectionError | null
	setDestinationTestConnectionError: (error: TestConnectionError | null) => void

	fetchDestinations: () => Promise<Entity[]>
	fetchDestination: (id: string) => Promise<void>
	addDestination: (destination: EntityBase) => Promise<EntityBase>
	updateDestination: (
		id: string,
		destination: Partial<Entity>,
	) => Promise<EntityBase>
	deleteDestination: (id: string) => Promise<void>
}

export const createDestinationSlice: StateCreator<DestinationSlice> = set => ({
	destinations: [],
	isLoadingDestinations: false,
	destinationsError: null,
	destination: null,
	isLoadingDestination: false,
	destinationError: null,
	destinationTestConnectionError: null,

	fetchDestinations: async () => {
		set({ isLoadingDestinations: true, destinationsError: null })
		try {
			const destinations = await destinationService.getDestinations()
			set({ destinations, isLoadingDestinations: false })
			return destinations
		} catch (error) {
			set({
				isLoadingDestinations: false,
				destinationsError:
					error instanceof Error
						? error.message
						: "Failed to fetch destinations",
			})
			throw error
		}
	},

	fetchDestination: async (id: string) => {
		set({
			isLoadingDestination: true,
			destinationError: null,
			destination: null,
		})
		try {
			const destination = await destinationService.getDestination(id)
			set({
				destination: destination,
				isLoadingDestination: false,
			})
		} catch (error) {
			set({
				isLoadingDestination: false,
				destinationError:
					error instanceof Error
						? error.message
						: "Failed to fetch destination",
			})
			throw error
		}
	},

	addDestination: async destinationData => {
		try {
			const newDestination =
				await destinationService.createDestination(destinationData)
			set(state => ({
				destinations: [
					...state.destinations,
					newDestination as unknown as Entity,
				],
			}))
			return newDestination as EntityBase
		} catch (error) {
			set({
				destinationsError:
					error instanceof Error ? error.message : "Failed to add destination",
			})
			throw error
		}
	},

	updateDestination: async (id, destinationData) => {
		try {
			const updatedDestination = await destinationService.updateDestination(
				id,
				destinationData as EntityBase,
			)

			set(state => ({
				destinations: state.destinations.map(destination =>
					destination.id.toString() === id
						? (updatedDestination as unknown as Entity)
						: destination,
				),
			}))
			return updatedDestination
		} catch (error) {
			set({
				destinationsError:
					error instanceof Error
						? error.message
						: "Failed to update destination",
			})
			throw error
		}
	},

	deleteDestination: async id => {
		try {
			const numericId = typeof id === "string" ? parseInt(id, 10) : id
			await destinationService.deleteDestination(numericId)
			set(state => ({
				destinations: state.destinations.filter(
					destination => destination.id !== numericId,
				),
			}))
		} catch (error) {
			set({
				destinationsError:
					error instanceof Error
						? error.message
						: "Failed to delete destination",
			})
			throw error
		}
	},
	setDestinationTestConnectionError: error =>
		set({ destinationTestConnectionError: error }),
})

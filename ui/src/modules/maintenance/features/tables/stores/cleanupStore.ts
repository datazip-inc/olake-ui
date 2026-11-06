import { create } from "zustand"
import { createJSONStorage, persist } from "zustand/middleware"

import type { CleanupConfigFormModel } from "../types"

/**
 * Client-side home for the cleanup (expire snapshots) configuration.
 * Fusion has no API for it yet, so saved configs live here per table,
 * keyed by catalog/database/table, and survive a reload via sessionStorage.
 * Swap the setters for mutations once the backend lands.
 */
interface CleanupState {
	configs: Record<string, CleanupConfigFormModel>
	setCleanupConfig: (key: string, config: CleanupConfigFormModel) => void
	disableCleanup: (key: string) => void
}

export const getCleanupKey = (
	catalog: string,
	database: string,
	tableName: string,
) => `${catalog}/${database}/${tableName}`

export const useCleanupStore = create<CleanupState>()(
	persist(
		set => ({
			configs: {},
			setCleanupConfig: (key, config) =>
				set(state => ({ configs: { ...state.configs, [key]: config } })),
			disableCleanup: key =>
				set(state => {
					const existing = state.configs[key]
					if (!existing?.expireSnapshots) return state
					return {
						configs: {
							...state.configs,
							[key]: { ...existing, expireSnapshots: false },
						},
					}
				}),
		}),
		{
			name: "maintenance-cleanup-storage",
			storage: createJSONStorage(() => sessionStorage),
		},
	),
)

/** Saved cleanup config for a table, or undefined when it was never saved. */
export const useCleanupConfig = (
	catalog: string,
	database: string,
	tableName: string,
): CleanupConfigFormModel | undefined =>
	useCleanupStore(
		state => state.configs[getCleanupKey(catalog, database, tableName)],
	)

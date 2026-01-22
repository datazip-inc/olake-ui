import { create } from "zustand"
import { persist, createJSONStorage } from "zustand/middleware"
import { platformService } from "../api/services/platformService"
import { ReleasesResponse, ReleaseType } from "../types/platformTypes"
import { processReleasesData } from "../utils/utils"

interface PlatformState {
	releases: ReleasesResponse | null
	releasesError: string | null
	isLoadingReleases: boolean
	hasSeenUpdates: boolean
	seenCategories: ReleaseType[]
	fetchReleases: (limit?: number) => Promise<void>
	setHasSeenUpdates: (seen: boolean) => void
	markCategoryAsSeen: (category: ReleaseType) => void
}

export const usePlatformStore = create<PlatformState>()(
	persist(
		(set, get) => ({
			releases: null,
			releasesError: null,
			isLoadingReleases: false,
			hasSeenUpdates: false,
			seenCategories: [],

			fetchReleases: async (limit?: number) => {
				// do not refetch if already present
				if (get().releases) return

				set({ isLoadingReleases: true, releasesError: null })

				try {
					const apiResponse = await platformService.getReleases(limit)
					const releases = processReleasesData(apiResponse)

					set({
						releases,
						isLoadingReleases: false,
					})
				} catch (error) {
					set({
						isLoadingReleases: false,
						releasesError:
							error instanceof Error
								? error.message
								: "Failed to fetch releases",
					})
					throw error
				}
			},

			setHasSeenUpdates: (seen: boolean) => set({ hasSeenUpdates: seen }),

			markCategoryAsSeen: (category: ReleaseType) =>
				set(state => ({
					seenCategories: state.seenCategories.includes(category)
						? state.seenCategories
						: [...state.seenCategories, category],
				})),
		}),
		{
			name: "platform-storage",
			storage: createJSONStorage(() => sessionStorage),
			partialize: state => ({
				releases: state.releases,
				hasSeenUpdates: state.hasSeenUpdates,
				seenCategories: state.seenCategories,
			}),
		},
	),
)

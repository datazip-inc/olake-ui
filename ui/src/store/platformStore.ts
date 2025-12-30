import { StateCreator } from "zustand"
import { platformService } from "../api/services/platformService"
import { ReleasesResponse } from "../types/platformTypes"
import { processReleasesData } from "../utils/utils"

export interface PlatformSlice {
	releases: ReleasesResponse | null
	releasesError: string | null
	isLoadingReleases: boolean
	fetchReleases: (limit?: number) => Promise<void>
}

export const createPlatformSlice: StateCreator<PlatformSlice> = set => ({
	releases: null,
	releasesError: null,
	isLoadingReleases: false,

	fetchReleases: async (limit?: number) => {
		set({ isLoadingReleases: true, releasesError: null })
		try {
			const apiResponse = await platformService.getReleases(limit)
			const releases = processReleasesData(apiResponse)
			set({ releases, isLoadingReleases: false })
		} catch (error) {
			set({
				isLoadingReleases: false,
				releasesError:
					error instanceof Error ? error.message : "Failed to fetch releases",
			})
			throw error
		}
	},
})

import api from "../axios"
import { API_CONFIG } from "../config"
import { ReleasesResponse } from "../../types/platformTypes"

export const platformService = {
	getReleases: async (limit?: number): Promise<ReleasesResponse> => {
		try {
			const queryParams: Record<string, string> = {}
			if (limit) {
				queryParams.limit = String(limit)
			}
			const query = new URLSearchParams(queryParams)

			const response = await api.get<ReleasesResponse>(
				`${API_CONFIG.ENDPOINTS.PLATFORM}/releases?${query.toString()}`,
			)
			return response.data
		} catch (error) {
			console.error("Error fetching releases from API:", error)
			throw error
		}
	},
}

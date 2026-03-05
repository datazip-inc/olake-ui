import api from "@/api/axios"
import { API_CONFIG } from "@/api/config"
import { ENTITY_TYPES } from "@/common/constants/constants"

export const validationService = {
	checkUniqueName: async (
		name: string,
		entityType: (typeof ENTITY_TYPES)[keyof typeof ENTITY_TYPES],
	): Promise<boolean | null> => {
		try {
			const response = await api.post<{ unique: boolean }>(
				`${API_CONFIG.ENDPOINTS.PROJECT(API_CONFIG.PROJECT_ID)}/check-unique`,
				{ name, entity_type: entityType },
			)
			return response.data.unique
		} catch (error) {
			console.error(`Error checking ${entityType} name uniqueness:`, error)
			return null
		}
	},
}

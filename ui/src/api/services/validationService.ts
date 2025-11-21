import { jobService } from "./jobService"
import { sourceService } from "./sourceService"
import { destinationService } from "./destinationService"
import { ENTITY_TYPES } from "../../utils/constants"

export const validationService = {
	checkUniqueName: async (
		name: string,
		entityType: (typeof ENTITY_TYPES)[keyof typeof ENTITY_TYPES],
	): Promise<boolean | null> => {
		try {
			switch (entityType) {
				case ENTITY_TYPES.JOB: {
					const response = await jobService.checkJobNameUnique(name)
					return response.unique
				}
				case ENTITY_TYPES.SOURCE: {
					const response = await sourceService.checkSourceNameUnique(name)
					return response.unique
				}
				case ENTITY_TYPES.DESTINATION: {
					const response = await destinationService.checkDestinationNameUnique(name)
					return response.unique
				}
				default:
					throw new Error("Invalid type")
			}
		} catch (error) {
			console.error(`Error checking ${entityType} name uniqueness:`, error)
			return null 
		}
	},
}


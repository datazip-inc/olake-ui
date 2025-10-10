import { CreateSourcePage } from "../pages/CreateSourcePage"
import { CreateDestinationPage } from "../pages/CreateDestinationPage"

/**
 * Verifies entity creation success by checking:
 * 1. Test connection modal appears
 * 2. Test connection succeeds
 * 3. Entity saved modal appears and navigates back
 */
export const verifyEntityCreationSuccessModal = async (
	pageObject: CreateSourcePage | CreateDestinationPage,
) => {
	await pageObject.expectTestConnectionModal()
	await pageObject.assertTestConnectionSucceeded()
	await pageObject.expectEntitySavedModal()
}

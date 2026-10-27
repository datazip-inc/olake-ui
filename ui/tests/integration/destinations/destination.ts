/** Steps any destination connector uses. */

import { CreateDestinationPage } from "../../pages/CreateDestinationPage"
import { DestinationsPage } from "../../pages/DestinationsPage"
import { DestinationFormConfig } from "../../types/PageConfig.types"
import { verifyEntityCreationSuccessModal } from "../../utils"

/** Creates a destination: fills the Create Destination form from config and saves it. */
export const createDestination = async (
	destinationsPage: DestinationsPage,
	createDestinationPage: CreateDestinationPage,
	config: DestinationFormConfig,
) => {
	await destinationsPage.navigateToDestinations()
	await destinationsPage.expectDestinationsPageVisible()
	await destinationsPage.clickCreateDestination()
	await createDestinationPage.expectCreateDestinationPageVisible()
	await createDestinationPage.fillDestinationForm(config)
	await createDestinationPage.clickCreate()
	await verifyEntityCreationSuccessModal(createDestinationPage)
	await destinationsPage.expectDestinationsPageVisible()
	await destinationsPage.expectDestinationExists(config.name)
}

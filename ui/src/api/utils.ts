import { EntityTestRequest, EntityTestResponse } from "../types"
import { TEST_CONNECTION_STATUS } from "../utils/constants"
import {
	EVENT_TEST_CONNECTION_DESTINATION,
	EVENT_TEST_CONNECTION_SOURCE,
} from "./constants"
import analyticsService from "./services/analyticsService"

export const trackTestConnection = async (
	isSource: boolean,
	req: EntityTestRequest,
	response: EntityTestResponse,
	isExisting: boolean = false,
) => {
	let catalogType

	if (!isSource) {
		const config = JSON.parse(req.config)
		catalogType = config.writer.catalog_type
	}

	const properties = {
		type: req.type,
		version: req.version,
		success:
			response.connection_result.status === TEST_CONNECTION_STATUS.SUCCEEDED,
		existing: isExisting,
		...(catalogType && { catalog: catalogType }),
	}

	analyticsService.trackEvent(
		isSource ? EVENT_TEST_CONNECTION_SOURCE : EVENT_TEST_CONNECTION_DESTINATION,
		properties,
	)
}

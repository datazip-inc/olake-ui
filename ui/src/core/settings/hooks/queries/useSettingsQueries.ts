import { useQuery } from "@tanstack/react-query"

import { settingsKeys } from "../../constants/queryKeys"
import { settingsService } from "../../services"

export const useSystemSettings = () => {
	return useQuery({
		queryKey: settingsKeys.systemSettings(),
		queryFn: () => settingsService.getSystemSettings(),
	})
}

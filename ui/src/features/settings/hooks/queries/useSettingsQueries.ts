import { useQuery } from "@tanstack/react-query"
import { settingsService } from "../../services"
import { settingsKeys } from "../../constants/queryKeys"

export const useSystemSettings = () => {
	return useQuery({
		queryKey: settingsKeys.systemSettings(),
		queryFn: () => settingsService.getSystemSettings(),
	})
}

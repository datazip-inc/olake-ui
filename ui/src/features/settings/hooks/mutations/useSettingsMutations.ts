import { useMutation } from "@tanstack/react-query"
import { settingsService } from "../../services"
import { settingsKeys } from "../../constants/queryKeys"
import type { UpdateSystemSettingsRequest } from "../../types"

export const useUpdateSystemSettings = () => {
	return useMutation({
		mutationKey: settingsKeys.systemSettings(),
		mutationFn: (settings: UpdateSystemSettingsRequest) =>
			settingsService.updateSystemSettings(settings),
	})
}

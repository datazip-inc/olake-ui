import { useMutation } from "@tanstack/react-query"

import { settingsKeys } from "../../constants/queryKeys"
import { settingsService } from "../../services"
import type { UpdateSystemSettingsRequest } from "../../types"

export const useUpdateSystemSettings = () => {
	return useMutation({
		mutationKey: settingsKeys.systemSettings(),
		mutationFn: (settings: UpdateSystemSettingsRequest) =>
			settingsService.updateSystemSettings(settings),
	})
}

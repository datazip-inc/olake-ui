import { useEffect } from "react"
import { useLocation } from "react-router-dom"

import sessionStore from "@/common/utils/session"
import { trackEvent, AnalyticsEvent } from "@/core/analytics"

import { matchesPath, type NavModule } from "../nav-config"

const MODULE_OPENED_EVENTS: Record<string, AnalyticsEvent> = {
	maintenance: AnalyticsEvent.MaintenanceModuleOpened,
}

export const useModuleAnalytics = (modules: NavModule[]) => {
	const { pathname } = useLocation()

	useEffect(() => {
		const mod = modules.find(m =>
			m.items.some(i => matchesPath(pathname, i.path)),
		)
		const event = mod && MODULE_OPENED_EVENTS[mod.key]
		if (!event) return

		const key = `analytics_fired_${event}`
		if (sessionStore.get<boolean>(key)) return

		trackEvent(event, { path: pathname })
		sessionStore.set(key, true)
	}, [pathname, modules])
}

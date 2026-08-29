import { SlidersIcon } from "@phosphor-icons/react"

import type { NavModule } from "@/core/layout/nav-config"

export const settingsNavModule: NavModule = {
	key: "system",
	section: "System",
	items: [{ path: "/settings", label: "Settings", icon: SlidersIcon }],
}

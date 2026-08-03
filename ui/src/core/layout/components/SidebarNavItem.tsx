import clsx from "clsx"
import { NavLink } from "react-router-dom"

import { trackEvent, AnalyticsEvent } from "@/core/analytics"

const MAINTENANCE_PATHS = new Set([
	"/maintenance/tables",
	"/maintenance/catalogs",
])

const MAINTENANCE_SESSION_KEY = "maintenance_module_opened_session"

const SidebarNavItem: React.FC<{
	path: string
	label: string
	icon: React.ElementType
	iconSize?: number
	className?: string
}> = ({ path, label, icon: Icon, iconSize = 14, className }) => (
	<NavLink
		to={path}
		onClick={() => {
			if (
				MAINTENANCE_PATHS.has(path) &&
				!sessionStorage.getItem(MAINTENANCE_SESSION_KEY)
			) {
				console.log("Maintenance module opened", path)
				//if cookie does not have maintainance
				trackEvent(AnalyticsEvent.MaintenanceModuleOpened, { path })
				sessionStorage.setItem(MAINTENANCE_SESSION_KEY, "true")
			}
		}}
		className={({ isActive }) =>
			clsx(
				"flex items-center gap-[9px] rounded-md px-2 text-[14px] leading-[22px]",
				isActive
					? "bg-[#f5f5f5] text-[#141414]"
					: "text-[#595959] hover:bg-[#f5f5f5]",
				className,
			)
		}
	>
		<Icon size={iconSize} />
		<span>{label}</span>
	</NavLink>
)

export default SidebarNavItem

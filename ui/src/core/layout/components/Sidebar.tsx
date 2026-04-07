import { InfoIcon, SidebarSimpleIcon, SignOutIcon } from "@phosphor-icons/react"
import { Tooltip } from "antd"
import clsx from "clsx"
import { useEffect, useState } from "react"
import { Link, NavLink, useLocation } from "react-router-dom"

import { OLake, OlakeLogo } from "@/assets"

import { matchesPath, NavModule, SYSTEM_ITEMS } from "../nav-config"
import SidebarModuleGroup from "./SidebarModuleGroup"
import SidebarNavItem from "./SidebarNavItem"
import UpdateNotification from "./UpdateNotification"

const getOpenModules = (
	navModules: NavModule[],
	pathname: string,
): Record<string, boolean> =>
	Object.fromEntries(
		navModules.map(m => [
			m.key,
			m.items.some(item => matchesPath(pathname, item.path)),
		]),
	)

const Sidebar: React.FC<{
	collapsed: boolean
	onToggle: () => void
	onLogout: () => void
	onOpenUpdates: () => void
	navModules: NavModule[]
}> = ({ collapsed, onToggle, onLogout, onOpenUpdates, navModules }) => {
	const { pathname } = useLocation()

	const [openModules, setOpenModules] = useState<Record<string, boolean>>(() =>
		getOpenModules(navModules, pathname),
	)

	useEffect(() => {
		setOpenModules(getOpenModules(navModules, pathname))
	}, [pathname, navModules])

	const toggleModule = (key: string) =>
		setOpenModules(prev =>
			Object.fromEntries(
				navModules.map(m => [m.key, m.key === key ? !prev[key] : false]),
			),
		)

	const navSections = [...new Set(navModules.map(m => m.section))]

	return (
		<div
			className={clsx(
				"relative flex flex-col border-r border-olake-border bg-olake-surface font-sans transition-all duration-300 ease-in-out",
				collapsed ? "w-[72px]" : "w-[257px]",
			)}
		>
			{/* Header */}
			{collapsed ? (
				<div className="flex h-[72px] items-center justify-center">
					<button
						onClick={onToggle}
						className="rounded-md p-1 text-olake-icon-muted hover:bg-olake-surface-muted"
						aria-label="Expand sidebar"
					>
						<SidebarSimpleIcon size={20} />
					</button>
				</div>
			) : (
				<div className="flex items-center justify-between pl-4 pr-4 pt-6">
					<Link
						to="/jobs"
						className="mb-3 flex items-center gap-2"
					>
						<img
							src={OlakeLogo}
							alt="logo"
							className="h-6 w-6"
						/>
						<img
							src={OLake}
							alt="logo"
							className="h-[27px] w-[57px]"
						/>
					</Link>
					<button
						onClick={onToggle}
						className="mb-3 rounded-md p-1 text-olake-icon-muted hover:bg-olake-surface-muted"
						aria-label="Toggle sidebar"
					>
						<SidebarSimpleIcon size={16} />
					</button>
				</div>
			)}

			<nav className="flex flex-1 flex-col overflow-y-auto px-6 pb-6">
				{collapsed ? (
					<div className="flex h-full flex-col items-center px-0 pb-6 pt-1">
						<div className="flex flex-col items-center gap-4">
							{navModules.map((mod, moduleIndex) => (
								<div
									key={mod.key}
									className={clsx(
										"flex flex-col items-center gap-4",
										moduleIndex > 0 && "mt-1 pt-2",
									)}
								>
									{mod.items.map(({ path, label, icon: Icon }) => (
										<Tooltip
											key={path}
											title={label}
											placement="right"
										>
											<NavLink
												to={path}
												className={clsx(
													"flex items-center justify-center rounded-md p-1",
													matchesPath(pathname, path)
														? "bg-olake-surface-muted text-olake-heading-strong"
														: "text-olake-body hover:bg-olake-surface-muted",
												)}
											>
												<Icon size={20} />
											</NavLink>
										</Tooltip>
									))}
								</div>
							))}

							{SYSTEM_ITEMS.map(({ path, label, icon: Icon }) => (
								<Tooltip
									key={path}
									title={label}
									placement="right"
								>
									<NavLink
										to={path}
										className={({ isActive }) =>
											clsx(
												"flex items-center justify-center rounded-md p-1",
												isActive
													? "bg-olake-surface-muted text-olake-heading-strong"
													: "text-olake-body hover:bg-olake-surface-muted",
											)
										}
									>
										<Icon size={20} />
									</NavLink>
								</Tooltip>
							))}
						</div>

						<button
							onClick={onOpenUpdates}
							className="mt-auto flex items-center justify-center rounded-md p-1 text-olake-primary hover:bg-olake-surface-muted"
							aria-label="Open updates"
						>
							<InfoIcon
								size={16}
								weight="fill"
							/>
						</button>
					</div>
				) : (
					<>
						{navSections.map(section => (
							<div key={section}>
								<p className="mb-2 text-[12px] font-medium leading-5 text-olake-icon-muted">
									{section}
								</p>

								{navModules
									.filter(m => m.section === section)
									.map(mod => (
										<SidebarModuleGroup
											key={mod.key}
											mod={mod}
											isOpen={openModules[mod.key] ?? false}
											onToggle={() => toggleModule(mod.key)}
										/>
									))}
							</div>
						))}

						{/* System section */}
						<div>
							<p className="mb-2 text-[12px] font-medium leading-5 text-olake-icon-muted">
								System
							</p>
							{SYSTEM_ITEMS.map(({ path, label, icon }) => (
								<SidebarNavItem
									key={path}
									path={path}
									label={label}
									icon={icon}
									iconSize={16}
									className="mb-2 h-8"
								/>
							))}
						</div>

						{/* Bottom: update card + logout */}
						<div className="mt-auto">
							<div className="mb-4">
								<UpdateNotification onOpen={onOpenUpdates} />
							</div>
							<button
								onClick={onLogout}
								className="flex h-8 w-full items-center gap-[9px] rounded-md px-2 text-[14px] leading-[22px] text-olake-body hover:bg-olake-surface-muted"
							>
								<SignOutIcon size={16} />
								<span>Logout</span>
							</button>
						</div>
					</>
				)}
			</nav>
		</div>
	)
}

export default Sidebar

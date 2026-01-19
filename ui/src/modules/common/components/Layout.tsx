import { useState, useEffect } from "react"
import clsx from "clsx"
import { NavLink, Link, useNavigate } from "react-router-dom"
import { LayoutProps } from "antd"
import {
	ArrowsOutSimpleIcon,
	BellIcon,
	CaretLeftIcon,
	GearSixIcon,
	SignOutIcon,
} from "@phosphor-icons/react"

import { useAppStore } from "../../../store"
import { usePlatformStore } from "../../../store/platformStore"
import { NAV_ITEMS } from "../../../utils/constants"
import { ReleaseMetadataResponse } from "../../../types/platformTypes"
import { OlakeLogo, OLake } from "../../../assets"
import UpdatesModal from "../Modals/UpdatesModal"

// will be shown in the later period when we have new updates
const UpdateNotification: React.FC = () => {
	const { setShowUpdatesModal } = useAppStore()
	const { releases, isLoadingReleases, hasSeenUpdates, setHasSeenUpdates } =
		usePlatformStore()

	// Count new releases across all release types that have "New Release" tag
	const newUpdatesCount = releases
		? Object.values(releases).reduce((total, category) => {
				const count =
					category?.releases.filter((release: ReleaseMetadataResponse) =>
						release.tags.includes("New Release"),
					).length || 0
				return total + count
			}, 0)
		: 0

	const hasNewUpdates = newUpdatesCount > 0

	if (isLoadingReleases) {
		return null
	}

	const handleOpenModal = () => {
		setShowUpdatesModal(true)
		setHasSeenUpdates(true)
	}

	return (
		<>
			<div className="p-4">
				<div className="relative rounded-xl border border-[#efefef] bg-neutral-100 p-3">
					<button className="absolute right-2 top-2 rounded-full p-1 hover:bg-gray-200">
						<ArrowsOutSimpleIcon
							onClick={handleOpenModal}
							size={14}
							color="#383838"
						/>
					</button>
					<div className="flex w-[90%] flex-col gap-2">
						<div className="relative w-fit">
							{/* Red Dot - only show if user hasn't seen updates */}
							{hasNewUpdates && !hasSeenUpdates && (
								<div className="absolute right-0 top-0 h-2 w-2 rounded-full bg-red-500"></div>
							)}
							<BellIcon
								className=""
								size={17}
								color="#203FDD"
							/>
						</div>
						<div className="text-xs font-medium text-brand-blue">
							{hasNewUpdates ? (
								<>
									You have {newUpdatesCount} new update
									{newUpdatesCount !== 1 ? "s" : ""}
								</>
							) : (
								"You're all up to date!"
							)}
						</div>
						<div className="text-xs font-normal text-[#383838]">
							{hasNewUpdates
								? "Checkout the new fixes & updates"
								: "No new updates available at this time"}
						</div>
					</div>
				</div>
			</div>
			<UpdatesModal />
		</>
	)
}

const Sidebar: React.FC<{
	collapsed: boolean
	onToggle: () => void
	onLogout: () => void
}> = ({ collapsed, onToggle, onLogout }) => {
	return (
		<div
			className={clsx(
				"relative flex flex-col border-r border-gray-200 bg-white transition-all duration-300 ease-in-out",
				collapsed ? "w-20" : "w-64",
			)}
		>
			<div className="pl-4 pt-6">
				<Link
					to="/jobs"
					className="mb-3 flex items-center gap-2"
				>
					<img
						src={OlakeLogo}
						alt="logo"
						className={clsx(
							"transition-all duration-300 ease-in-out",
							collapsed ? "h-10 w-10 pl-1" : "h-6 w-6",
						)}
					/>
					{!collapsed && (
						<img
							src={OLake}
							alt="logo"
							className="h-[27px] w-[57px]"
						/>
					)}
				</Link>
			</div>

			<nav className="flex-1 space-y-2 p-4">
				{NAV_ITEMS.map(({ path, label, icon: Icon }) => (
					<NavLink
						key={path}
						to={path}
						className={({ isActive }) =>
							clsx(
								"flex items-center rounded-xl p-3",
								isActive
									? "bg-primary-100 text-primary hover:text-black"
									: "text-gray-700 hover:bg-gray-100 hover:text-black",
							)
						}
					>
						<Icon
							className="mr-3 flex-shrink-0"
							size={20}
						/>
						{!collapsed && <span>{label}</span>}
					</NavLink>
				))}
			</nav>

			{!collapsed && <UpdateNotification />}
			<div className="space-y-2 px-2 py-4">
				<div className="mt-auto px-4">
					<button
						onClick={onLogout}
						className="flex w-full items-center rounded-xl p-3 text-gray-700 hover:bg-gray-100 hover:text-black"
					>
						<SignOutIcon
							className="mr-3 flex-shrink-0"
							size={20}
						/>
						{!collapsed && <span>Logout</span>}
					</button>
				</div>
				<div className="px-4">
					<NavLink
						to="/settings"
						className={({ isActive }) =>
							clsx(
								"flex w-full items-center rounded-xl p-3",
								isActive
									? "bg-primary-100 text-primary hover:text-black"
									: "text-gray-700 hover:bg-gray-100 hover:text-black",
							)
						}
					>
						<GearSixIcon
							className="mr-3 flex-shrink-0"
							size={20}
						/>
						{!collapsed && <span>System Settings</span>}
					</NavLink>
				</div>
			</div>
			<button
				onClick={onToggle}
				className="absolute bottom-10 right-0 z-10 translate-x-1/2 rounded-xl border border-gray-200 bg-white p-2.5 text-gray-900 shadow-[0_6px_16px_0_rgba(0,0,0,0.08)] hover:text-gray-700 focus:outline-none"
			>
				<div
					className={clsx(
						"transition-transform duration-500",
						collapsed ? "rotate-180" : "rotate-0",
					)}
				>
					<CaretLeftIcon size={16} />
				</div>
			</button>
		</div>
	)
}

const Layout: React.FC<LayoutProps> = ({ children }) => {
	const [collapsed, setCollapsed] = useState(false)
	const { logout } = useAppStore()
	const { fetchReleases, releases } = usePlatformStore()
	const navigate = useNavigate()

	// Fetch releases if not in store (persist middleware handles sessionStorage)
	useEffect(() => {
		const hasReleases = releases && Object.keys(releases).length > 0
		if (!hasReleases) {
			fetchReleases()
		}
	}, []) // Only run once on mount

	const handleLogout = () => {
		logout()
		navigate("/login")
	}

	return (
		<div className="flex h-screen bg-gray-50">
			<Sidebar
				collapsed={collapsed}
				onToggle={() => setCollapsed(!collapsed)}
				onLogout={handleLogout}
			/>
			<div className="flex-1 overflow-auto bg-white">{children}</div>
		</div>
	)
}

export default Layout

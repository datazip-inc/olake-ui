import { useState } from "react"
import clsx from "clsx"
import { NavLink, Link, useNavigate } from "react-router-dom"
import { LayoutProps } from "antd"
import { CaretLeftIcon, InfoIcon, X, SignOutIcon } from "@phosphor-icons/react"

import { useAppStore } from "../../../store"
import { NAV_ITEMS } from "../../../utils/constants"
import OlakeLogo from "../../../assets/OlakeLogo.svg"
import Olake from "../../../assets/OLake.svg"

const UpdateNotification: React.FC<{ onClose: () => void }> = ({ onClose }) => (
	<div className="p-4">
		<div className="relative rounded-xl border border-gray-300 bg-gray-100 p-3">
			<button
				onClick={onClose}
				className="absolute right-2 top-2 rounded-full p-1 hover:bg-gray-200"
			>
				<X
					size={12}
					color="#383838"
				/>
			</button>
			<div className="flex items-center gap-2">
				<InfoIcon
					weight="fill"
					size={17}
					color="#203FDD"
				/>
				<span className="text-sm font-medium text-brand-blue">New Update</span>
			</div>
			<p className="mt-2 text-xs text-gray-900">
				We have made fixes to our ingestion flow & new UI is implemented
			</p>
		</div>
	</div>
)

const Sidebar: React.FC<{
	collapsed: boolean
	onToggle: () => void
	onLogout: () => void
	showUpdate: boolean
	onCloseUpdate: () => void
}> = ({ collapsed, onToggle, onLogout, showUpdate, onCloseUpdate }) => {
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
							src={Olake}
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

			{!collapsed && showUpdate && (
				<UpdateNotification onClose={onCloseUpdate} />
			)}

			<div className="mt-auto p-4">
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
	const [showUpdate, setShowUpdate] = useState(false)
	const { logout } = useAppStore()
	const navigate = useNavigate()

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
				showUpdate={showUpdate}
				onCloseUpdate={() => setShowUpdate(false)}
			/>
			<div className="flex-1 overflow-auto bg-white">{children}</div>
		</div>
	)
}

export default Layout

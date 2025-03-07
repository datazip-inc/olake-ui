import { useState } from "react"
import { NavLink } from "react-router-dom"
import OlakeLogo from "../../../assets/OlakeLogo.png"
import Olake from "../../../assets/Olake.svg"
import {
	CaretLeft,
	CaretRight,
	GitCommit,
	Info,
	LinktreeLogo,
	Path,
} from "@phosphor-icons/react"

interface LayoutProps {
	children: React.ReactNode
}

const Layout: React.FC<LayoutProps> = ({ children }) => {
	const [collapsed, setCollapsed] = useState(false)

	const toggleSidebar = () => {
		setCollapsed(!collapsed)
	}

	return (
		<div className="flex h-screen bg-gray-50">
			{/* Sidebar */}
			<div
				className={`${
					collapsed ? "w-20" : "w-64"
				} relative flex flex-col border-r border-gray-200 bg-white transition-all duration-300 ease-in-out`}
			>
				<div className="border-b border-gray-200 p-4">
					<div className="flex items-center gap-3">
						<img
							src={OlakeLogo}
							alt="logo"
							className="h-6 w-6"
						/>
						{!collapsed && (
							<img
								src={Olake}
								alt="logo"
								className="h-4 w-20"
							/>
						)}
					</div>
				</div>

				<nav className="flex-1 space-y-2 p-4">
					<NavLink
						to="/jobs"
						className={({ isActive }) =>
							`flex items-center rounded-lg p-3 ${
								isActive
									? "bg-blue-50 text-blue-600"
									: "text-gray-700 hover:bg-gray-100"
							}`
						}
					>
						<GitCommit
							className="mr-3"
							size={20}
						/>
						{!collapsed && <span>Jobs</span>}
					</NavLink>

					<NavLink
						to="/sources"
						className={({ isActive }) =>
							`flex items-center rounded-lg p-3 ${
								isActive
									? "bg-blue-50 text-blue-600"
									: "text-gray-700 hover:bg-gray-100"
							}`
						}
					>
						<LinktreeLogo
							className="mr-3"
							size={20}
						/>
						{!collapsed && <span>Sources</span>}
					</NavLink>

					<NavLink
						to="/destinations"
						className={({ isActive }) =>
							`flex items-center rounded-lg p-3 ${
								isActive
									? "bg-blue-50 text-blue-600"
									: "text-gray-700 hover:bg-gray-100"
							}`
						}
					>
						<Path
							className="mr-3"
							size={20}
						/>
						{!collapsed && <span>Destinations</span>}
					</NavLink>
				</nav>

				{!collapsed && (
					<div className="border-t border-gray-200 p-4">
						<div className="rounded-lg bg-blue-50 p-3">
							<div className="flex items-center gap-2">
								<Info size={16} />
								<span className="text-sm font-medium">New Update</span>
							</div>
							<p className="mt-2 text-xs text-gray-600">
								We have made fixes to our ingestion flow & new UI is implemented
							</p>
						</div>
					</div>
				)}
			</div>

			{/* Toggle button positioned at the edge of the sidebar */}
			<button
				onClick={toggleSidebar}
				className="absolute bottom-8 left-0 z-10 translate-x-1/2 transform rounded-full bg-white p-2 text-gray-500 shadow-md hover:text-gray-700 focus:outline-none"
			>
				{collapsed ? <CaretRight size={18} /> : <CaretLeft size={18} />}
			</button>

			{/* Main content */}
			<div className="flex-1 overflow-auto">{children}</div>
		</div>
	)
}

export default Layout

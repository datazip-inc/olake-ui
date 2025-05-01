import { useState } from "react"
import { NavLink, Link } from "react-router-dom"
import OlakeLogo from "../../../assets/OlakeLogo.svg"
import Olake from "../../../assets/OLake.svg"
import {
	CaretLeft,
	GitCommit,
	Info,
	LinktreeLogo,
	Path,
	X,
} from "@phosphor-icons/react"

interface LayoutProps {
	children: React.ReactNode
}

const Layout: React.FC<LayoutProps> = ({ children }) => {
	const [collapsed, setCollapsed] = useState(false)
	const [showUpdate, setShowUpdate] = useState(true)

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
				<div className="pl-4 pt-6">
					<Link
						to="/jobs"
						className="mb-3 flex items-center gap-2"
					>
						<img
							src={OlakeLogo}
							alt="logo"
							className={`transition-all duration-300 ease-in-out ${collapsed ? "h-10 w-10 pl-1" : "h-6 w-6"}`}
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
					<NavLink
						to="/jobs"
						className={({ isActive }) =>
							`flex items-center rounded-xl p-3 ${
								isActive
									? "bg-[#E9EBFC] text-[#203FDD] hover:text-black"
									: "text-gray-700 hover:bg-[#f6f6f6] hover:text-black"
							}`
						}
					>
						<GitCommit
							className="mr-3 flex-shrink-0"
							size={20}
						/>
						{!collapsed && <span>Jobs</span>}
					</NavLink>

					<NavLink
						to="/sources"
						className={({ isActive }) =>
							`flex items-center rounded-xl p-3 ${
								isActive
									? "bg-[#E9EBFC] text-[#203FDD] hover:text-black"
									: "text-gray-700 hover:bg-[#f6f6f6] hover:text-black"
							}`
						}
					>
						<LinktreeLogo
							className="mr-3 flex-shrink-0"
							size={20}
						/>
						{!collapsed && <span>Sources</span>}
					</NavLink>

					<NavLink
						to="/destinations"
						className={({ isActive }) =>
							`flex items-center rounded-xl p-3 ${
								isActive
									? "bg-[#E9EBFC] text-[#203FDD] hover:text-black"
									: "text-gray-700 hover:bg-[#f6f6f6] hover:text-black"
							}`
						}
					>
						<Path
							className="mr-3 flex-shrink-0"
							size={20}
						/>
						{!collapsed && <span>Destinations</span>}
					</NavLink>
				</nav>

				{!collapsed && showUpdate && (
					<div className="p-4">
						<div className="relative rounded-xl border-[1px] border-[#EFEFEF] bg-[#F6F6F6] p-3">
							<button
								onClick={() => setShowUpdate(false)}
								className="absolute right-2 top-2 rounded-full p-1 hover:bg-gray-200"
							>
								<X
									size={12}
									color="#383838"
								/>
							</button>
							<div className="flex items-center gap-2">
								<Info
									weight="fill"
									size={17}
									color="#203FDD"
								/>
								<span className="text-sm font-medium text-[#193AE6]">
									New Update
								</span>
							</div>
							<p className="mt-2 text-xs text-[#383838]">
								We have made fixes to our ingestion flow & new UI is implemented
							</p>
						</div>
					</div>
				)}

				{/* Toggle button positioned on the right border of sidebar */}
				<button
					onClick={toggleSidebar}
					className="absolute bottom-10 right-0 z-10 translate-x-1/2 rounded-xl border border-gray-200 bg-white p-2.5 text-[#383838] shadow-[0_6px_16px_0_rgba(0,0,0,0.08)] hover:text-gray-700 focus:outline-none"
				>
					<div
						className={`transition-transform duration-500 ${collapsed ? "rotate-180" : "rotate-0"}`}
					>
						<CaretLeft size={16} />
					</div>
				</button>
			</div>

			{/* Main content */}
			<div className="flex-1 overflow-auto bg-white">{children}</div>
		</div>
	)
}

export default Layout

import { LayoutProps } from "antd"
import { useEffect, useState } from "react"
import { useLocation, useNavigate } from "react-router-dom"

import { useAuthStore } from "@/core/auth/stores"
import { UpdatesModal } from "@/core/platform/components"
import { useCompactionStatus } from "@/core/platform/hooks/useCompactionStatus"
import { usePlatformStore } from "@/core/platform/stores"

import Sidebar from "./components/Sidebar"
import {
	getBreadcrumbModuleLabel,
	getBreadcrumbTrail,
	getNavModules,
} from "./nav-config"

const Layout: React.FC<LayoutProps> = ({ children }) => {
	const [collapsed, setCollapsed] = useState(false)
	const [showUpdatesModal, setShowUpdatesModal] = useState(false)
	const logout = useAuthStore(state => state.logout)
	const fetchReleases = usePlatformStore(state => state.fetchReleases)
	const releases = usePlatformStore(state => state.releases)
	const navigate = useNavigate()
	const location = useLocation()

	const { data: compactionStatus } = useCompactionStatus()

	const enabledFeatures = new Set<string>(
		compactionStatus?.enabled ? ["maintenance"] : [],
	)
	const navModules = getNavModules(enabledFeatures)

	const breadcrumbItems = getBreadcrumbTrail(location.pathname, navModules)

	useEffect(() => {
		const hasReleases = releases && Object.keys(releases).length > 0
		if (!hasReleases) {
			fetchReleases()
		}
	}, [])

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
				onOpenUpdates={() => setShowUpdatesModal(true)}
				navModules={navModules}
			/>
			<div className="flex min-w-0 flex-1 flex-col overflow-hidden bg-[#f5f5f5]">
				<div className="h-16 border-b border-[#d9d9d9] bg-white px-6">
					<div className="flex h-full items-center gap-2 text-sm text-[#8c8c8c]">
						<span>
							{getBreadcrumbModuleLabel(location.pathname, navModules)}
						</span>
						<span>/</span>
						{breadcrumbItems.map((item, index) => (
							<span
								key={`${item}-${index}`}
								className="contents"
							>
								{index > 0 && <span>/</span>}
								<span
									className={
										index === breadcrumbItems.length - 1
											? "text-[#595959]"
											: "text-[#8c8c8c]"
									}
								>
									{item}
								</span>
							</span>
						))}
					</div>
				</div>
				<div className="min-h-0 flex-1 overflow-auto bg-white">{children}</div>
			</div>
			<UpdatesModal
				open={showUpdatesModal}
				onClose={() => setShowUpdatesModal(false)}
			/>
		</div>
	)
}

export default Layout

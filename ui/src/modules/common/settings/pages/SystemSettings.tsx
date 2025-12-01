import { GearSixIcon } from "@phosphor-icons/react"
import { Spin, Tabs } from "antd"
import { useEffect, useState } from "react"
import AlertsAndNotifications from "../components/AlertsAndNotifications"
import { useAppStore } from "../../../../store"

const SystemSettings = () => {
	const [activeTab, setActiveTab] = useState("alerts")

	const { isLoadingSystemSettings, fetchSystemSettings } = useAppStore()

	const tabItems = [
		{
			key: "alerts",
			label: "Alerts and Notifications",
			children: <AlertsAndNotifications />,
		},
	]

	useEffect(() => {
		fetchSystemSettings()
	}, [])

	return (
		<div className="p-6">
			<div className="mb-4 flex items-center justify-between">
				<div className="flex items-center gap-2">
					<GearSixIcon className="mr-2 size-6" />
					<h1 className="text-2xl font-bold">System Settings</h1>
				</div>
			</div>

			<p className="mb-6 text-gray-600">
				Configure global system behaviour such as logging, data retention,
				backups, job defaults, and experimental feature flags
			</p>

			{isLoadingSystemSettings ? (
				<div className="flex items-center justify-center py-16">
					<Spin size="large" />
				</div>
			) : (
				<Tabs
					activeKey={activeTab}
					onChange={key => setActiveTab(key)}
					items={tabItems}
				/>
			)}
		</div>
	)
}

export default SystemSettings

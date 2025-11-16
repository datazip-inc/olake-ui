import { GearSixIcon } from "@phosphor-icons/react"
import { Tabs } from "antd"
import { useState } from "react"
import AlertsAndNotifications from "../components/AlertsAndNotifications"

const SystemSettings = () => {
	const [activeTab, setActiveTab] = useState("alerts")

	const tabItems = [
		{
			key: "alerts",
			label: "Alerts and Notifications",
			children: <AlertsAndNotifications />,
		},
	]

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

			<Tabs
				activeKey={activeTab}
				onChange={key => setActiveTab(key)}
				items={tabItems}
			/>
		</div>
	)
}

export default SystemSettings

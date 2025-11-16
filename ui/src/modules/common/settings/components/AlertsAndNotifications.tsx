import { Button } from "antd"
import { Input } from "antd/lib"

const AlertsAndNotifications = () => {
	return (
		<div className="mt-6">
			<div className="mb-1 flex flex-col gap-2">
				<div className="flex items-center gap-2">
					<h1 className="text-2xl font-bold">Alerts &amp; Notifications</h1>
				</div>
			</div>
			<p className="mb-6 text-gray-600">
				Configure alerts and notifications for your system
			</p>
			<div className="mb-6 rounded-xl border border-gray-200 bg-white px-6 pb-2">
				<div className="border-gray-200 pt-4">
					<div className="mb-2 flex flex-col gap-4">
						<div className="space-y-1">
							<div className="text-sm font-medium">Slack Alerts</div>
							<div className="text-sm text-text-tertiary">
								Configure Slack alerts for your system
							</div>
						</div>
						<div className="flex gap-2">
							<Input
								placeholder="Enter your Slack webhook URL"
								className="h-10 w-96"
							/>
							<Button
								type="default"
								className="h-10"
							>
								Save and Update
							</Button>
						</div>
					</div>
				</div>
			</div>
		</div>
	)
}

export default AlertsAndNotifications

import { Button, message } from "antd"
import { Input } from "antd/lib"
import { useAppStore } from "../../../../store"
import { useEffect, useMemo, useState } from "react"

const AlertsAndNotifications = () => {
	const { systemSettings, updateWebhookAlertUrl, isUpdatingSystemSettings } =
		useAppStore()

	const [webhookAlertUrl, setWebhookAlertUrl] = useState<string>("")

	const trimmedWebhookUrl = useMemo(
		() => webhookAlertUrl.trim(),
		[webhookAlertUrl],
	)

	useEffect(() => {
		if (systemSettings) {
			setWebhookAlertUrl(systemSettings.webhook_alert_url)
		}
	}, [systemSettings])

	const handleWebhookAlertUrlChange = (
		e: React.ChangeEvent<HTMLInputElement>,
	) => {
		setWebhookAlertUrl(e.target.value)
	}

	const handleSaveWebhookAlertUrl = () => {
		if (!trimmedWebhookUrl) {
			message.error("Please enter a webhook URL")
			return
		}
		try {
			new URL(trimmedWebhookUrl)
			updateWebhookAlertUrl(trimmedWebhookUrl)
		} catch {
			message.error("Please enter a valid webhook URL")
		}
	}

	const handleClearWebhookAlertUrl = () => {
		setWebhookAlertUrl("")
		updateWebhookAlertUrl("")
	}
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
			{/* TODO: After saving, lock these settings and show an "Edit" button to re-enable changes */}
			<div className="mb-6 rounded-xl border border-gray-200 bg-white px-6 pb-2">
				<div className="border-gray-200 pt-4">
					<div className="mb-2 flex flex-col gap-4">
						<div className="space-y-1">
							<div className="text-sm font-medium">Webhook Alerts</div>
							<div className="text-sm text-text-tertiary">
								Configure outgoing webhook alerts for your system (Slack, Teams,
								custom endpoints, etc.)
							</div>
						</div>
						<div className="flex gap-2">
							<Input
								placeholder="Enter your webhook URL"
								className="h-10 w-96 text-text-secondary"
								value={webhookAlertUrl}
								onChange={handleWebhookAlertUrlChange}
							/>
							<Button
								type="default"
								className="h-10"
								onClick={handleSaveWebhookAlertUrl}
								disabled={!trimmedWebhookUrl || isUpdatingSystemSettings}
							>
								Save
							</Button>
							<Button
								type="default"
								className="h-10"
								onClick={handleClearWebhookAlertUrl}
								disabled={isUpdatingSystemSettings || !trimmedWebhookUrl}
								aria-label="Clear webhook URL"
								title="Clear webhook URL"
							>
								Clear
							</Button>
						</div>
					</div>
				</div>
			</div>
		</div>
	)
}

export default AlertsAndNotifications

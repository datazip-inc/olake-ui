import { PencilSimpleIcon } from "@phosphor-icons/react"
import { Button, message } from "antd"
import { Input } from "antd/lib"
import { useEffect, useMemo, useState } from "react"

import { useSystemSettings, useUpdateSystemSettings } from "../hooks"

const AlertsAndNotifications = () => {
	const { data: systemSettings } = useSystemSettings()
	const updateSystemSettingsMutation = useUpdateSystemSettings()
	const isUpdatingSystemSettings = updateSystemSettingsMutation.isPending

	const [webhookAlertUrl, setWebhookAlertUrl] = useState<string>("")
	const [isEditingWebhookAlertUrl, setIsEditingWebhookAlertUrl] =
		useState<boolean>(false)

	const trimmedWebhookUrl = useMemo(
		() => webhookAlertUrl.trim(),
		[webhookAlertUrl],
	)

	useEffect(() => {
		if (systemSettings) {
			setWebhookAlertUrl(systemSettings.webhook_alert_url)
			// If the webhook alert URL is not set, allow editing
			setIsEditingWebhookAlertUrl(!systemSettings.webhook_alert_url)
		}
	}, [systemSettings])

	const handleWebhookAlertUrlChange = (
		e: React.ChangeEvent<HTMLInputElement>,
	) => {
		if (!isEditingWebhookAlertUrl) {
			return
		}
		setWebhookAlertUrl(e.target.value)
	}

	const handleSaveWebhookAlertUrl = () => {
		if (!trimmedWebhookUrl) {
			message.error("Please enter a webhook URL")
			return
		}
		if (!systemSettings) return
		try {
			new URL(trimmedWebhookUrl)
			updateSystemSettingsMutation.mutate(
				{ ...systemSettings, webhook_alert_url: trimmedWebhookUrl },
				{ onSuccess: () => setIsEditingWebhookAlertUrl(false) },
			)
		} catch {
			message.error("Please enter a valid webhook URL")
		}
	}

	const handleEnterEditWebhookAlertUrl = () => {
		setIsEditingWebhookAlertUrl(true)
	}

	const handleCancelEditWebhookAlertUrl = () => {
		if (!systemSettings?.webhook_alert_url) return
		if (systemSettings) {
			setWebhookAlertUrl(systemSettings.webhook_alert_url)
		}
		setIsEditingWebhookAlertUrl(false)
	}

	const handleClearWebhookAlertUrl = () => {
		setWebhookAlertUrl("")
		if (!systemSettings) return
		updateSystemSettingsMutation.mutate({
			...systemSettings,
			webhook_alert_url: "",
		})
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
								disabled={!isEditingWebhookAlertUrl || isUpdatingSystemSettings}
								onChange={handleWebhookAlertUrlChange}
							/>
							{isEditingWebhookAlertUrl ? (
								<>
									<Button
										type="primary"
										className="h-10"
										onClick={handleSaveWebhookAlertUrl}
										disabled={!trimmedWebhookUrl || isUpdatingSystemSettings}
									>
										Save changes
									</Button>
									<Button
										type="default"
										className="h-10"
										onClick={handleCancelEditWebhookAlertUrl}
										disabled={isUpdatingSystemSettings}
									>
										Cancel
									</Button>
								</>
							) : (
								<>
									<Button
										type="default"
										className="h-10"
										onClick={handleEnterEditWebhookAlertUrl}
										disabled={isUpdatingSystemSettings}
									>
										<span className="flex items-center gap-1">
											<PencilSimpleIcon className="size-4" />
											<span>Edit</span>
										</span>
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
								</>
							)}
						</div>
					</div>
				</div>
			</div>
		</div>
	)
}

export default AlertsAndNotifications

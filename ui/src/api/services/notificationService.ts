import { message } from "antd"

/**
 * A decoupled service for showing UI notifications.
 */
export const notificationService = {
	success: (msg: string) => {
		if (msg) {
			message.destroy()
			message.success(msg)
		}
	},
	error: (msg: string) => {
		if (msg) {
			message.destroy()
			message.error(msg)
		}
	},
}

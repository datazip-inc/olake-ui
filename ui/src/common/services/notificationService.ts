import { message } from "antd"

/**
 * A decoupled service for showing UI notifications.
 */

const ERROR_MESSAGE_DURATION = 6 // seconds

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
			message.error(msg, ERROR_MESSAGE_DURATION)
		}
	},
}

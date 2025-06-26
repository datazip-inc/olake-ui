import { AnalyticsBrowser } from "@segment/analytics-next"
import api from "../axios"

const analytics = AnalyticsBrowser.load({
	writeKey: "e2lmlXGqXwqBBkSAnP7BxsjBpAGZNbWk",
})

const getIPAddress = async (): Promise<string> => {
	try {
		const response = await fetch("https://api.ipify.org?format=json")
		const data = await response.json()
		return data.ip
	} catch (error) {
		console.error("Error fetching IP:", error)
		return ""
	}
}

const getLocationInfo = async (ip: string) => {
	try {
		const response = await fetch(`https://ipinfo.io/${ip}/json`)
		const data = await response.json()
		return {
			country: data.country,
			region: data.region,
			city: data.city,
		}
	} catch (error) {
		console.error("Error fetching location:", error)
		return null
	}
}

const getSystemInfo = async () => {
	const ip = await getIPAddress()
	const location = ip ? await getLocationInfo(ip) : null

	return {
		os: navigator.platform,
		arch: navigator.userAgent.includes("64") ? "x64" : "x86",
		device_cpu: navigator.hardwareConcurrency + " cores",
		ip_address: ip,
		location: location || "",
		timestamp: new Date().toISOString(),
	}
}

const getTelemetryID = async (): Promise<string> => {
	try {
		const response = await api.get("/telemetry-id")
		return response.data.data.telemetry_id
	} catch (error) {
		console.error("Error fetching telemetry ID:", error)
		return ""
	}
}

export const identifyUser = async () => {
	try {
		const username = localStorage.getItem("username")
		const systemInfo = await getSystemInfo()
		const telemetryId = await getTelemetryID()

		if (telemetryId) {
			await analytics.identify(telemetryId, {
				username,
				...systemInfo,
			})
			return true
		}
		return false
	} catch (error) {
		console.error("Error identifying user:", error)
		return false
	}
}

export const trackEvent = async (
	eventName: string,
	properties?: Record<string, any>,
) => {
	try {
		const username = localStorage.getItem("username")
		const systemInfo = await getSystemInfo()

		const eventProperties = {
			...properties,
			...systemInfo,
			...(username && { username }),
		}
		await analytics.track(eventName, eventProperties)
	} catch (error) {
		console.error("Error tracking event:", error)
	}
}

export default {
	trackEvent,
	identifyUser,
}

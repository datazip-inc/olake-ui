/**
 * AnalyticsService handles sending analytics events
 */

import api from "../axios"
import axios from "axios"

// endpoint which handles rate limiting and forwards the events to mixpanel
const ANALYTICS_ENDPOINT = "https://analytics.olake.io/mp/track"

const sendAnalyticsEvent = async (
	eventName: string,
	properties: Record<string, any>,
) => {
	try {
		const eventData = {
			event: eventName,
			properties,
		}

		await axios.post(ANALYTICS_ENDPOINT, eventData)
	} catch (error) {
		console.error("Failed to send analytics event:", error)
	}
}

const getIPAddress = async (): Promise<string> => {
	try {
		const response = await axios.get("https://api.ipify.org?format=json")
		return response.data.ip
	} catch (error) {
		console.error("Error fetching IP:", error)
		return ""
	}
}

const getLocationInfo = async (ip: string) => {
	try {
		const response = await axios.get(`https://ipinfo.io/${ip}/json`)
		return {
			country: response.data.country,
			region: response.data.region,
			city: response.data.city,
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

// returns a unique user id for the user to track them across sessions
const getTelemetryID = async (): Promise<string> => {
	try {
		const response = await api.get("/telemetry-id")
		return response.data.user_id || ""
	} catch (error) {
		console.error("Error fetching telemetry ID:", error)
		return ""
	}
}

export const trackEvent = async (
	eventName: string,
	properties?: Record<string, any>,
) => {
	try {
		const telemetryId = await getTelemetryID()
		if (!telemetryId || telemetryId === "") {
			return
		}

		// if user is already logged in we'll get the username from local storage
		const username = localStorage.getItem("username")
		const systemInfo = await getSystemInfo()

		const eventProperties = {
			distinct_id: telemetryId,
			event_original_name: eventName,
			...properties,
			...systemInfo,
			...(username && { username }),
		}

		await sendAnalyticsEvent(eventName, eventProperties)
	} catch (error) {
		console.error("Error tracking event:", error)
	}
}

export default {
	trackEvent,
}

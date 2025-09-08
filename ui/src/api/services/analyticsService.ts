import api from "../axios"

const ANALYTICS_ENDPOINT = "https://analytics.olake.io/mp/track"

const sendAnalyticsEvent = async (
	eventName: string,
	properties: Record<string, any>,
) => {
	const eventData = {
		event: eventName,
		properties,
	}

	const response = await fetch(ANALYTICS_ENDPOINT, {
		method: "POST",
		headers: {
			"Content-Type": "application/json",
		},
		body: JSON.stringify(eventData),
	})

	if (!response.ok) {
		throw new Error(`Failed to send analytics event: ${response.statusText}`)
	}
}

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
		return response.data.data.user_id || ""
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

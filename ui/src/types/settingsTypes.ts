export interface SystemSettings {
	id: number
	project_id: string
	webhook_alert_url: string
}

export interface UpdateSystemSettingsRequest {
	id?: number
	project_id: string
	webhook_alert_url: string
}

export interface GetSystemSettingsResponse {
	id: number
	project_id: string
	webhook_alert_url: string
}

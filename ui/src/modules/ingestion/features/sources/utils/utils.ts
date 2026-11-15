import { UPDATE_METHOD_CDC } from "@/common/components/form/cdc"

export const getConnectorLabel = (type: string): string => {
	switch (type) {
		case "mongodb":
		case "MongoDB":
			return "MongoDB"
		case "postgres":
		case "Postgres":
			return "Postgres"
		case "mysql":
		case "MySQL":
			return "MySQL"
		case "oracle":
		case "Oracle":
			return "Oracle"
		case "kafka":
			return "Kafka"
		case "s3":
		case "S3":
			return "S3"
		case "db2":
			return "DB2"
		case "mssql":
			return "MSSQL"
		default:
			return "Default"
	}
}

// Postgres preserve_toast_values can only be set when the source is created: it
// limits the source's jobs to update types that keep a row index. Edit shows it
// disabled with the saved value.
export const lockPreserveToastValues = (uiSchema: any, jsonSchema: any) => {
	const description = jsonSchema?.properties?.update_method?.oneOf?.find(
		(branch: any) => branch?.properties?.preserve_toast_values,
	)?.properties.preserve_toast_values.description
	return {
		...uiSchema,
		update_method: {
			...uiSchema?.update_method,
			preserve_toast_values: {
				...uiSchema?.update_method?.preserve_toast_values,
				"ui:disabled": true,
				"ui:description":
					`${description ?? ""} Can only be set when the source is created.`.trim(),
			},
		},
	}
}

// Puts the saved preserve_toast_values back on every edit form change, since rjsf
// resets it on CDC/Standalone switches. Spec versions without the field get no key.
export const restorePreserveToastValues = (
	formData: any,
	saved: boolean | undefined,
) => {
	const updateMethod = formData?.update_method
	if (
		updateMethod?.type !== UPDATE_METHOD_CDC ||
		(saved === undefined && !("preserve_toast_values" in updateMethod))
	) {
		return formData
	}
	return {
		...formData,
		update_method: { ...updateMethod, preserve_toast_values: saved ?? false },
	}
}

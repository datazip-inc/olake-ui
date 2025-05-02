import { StreamData } from "../types"

export const mockStreamData: StreamData[] = [
	{
		sync_mode: "full_refresh",
		destination_sync_mode: "overwrite",
		selected_columns: null,
		sort_key: ["eventn_ctx_event_id"],
		stream: {
			json_schema: {
				properties: {
					"canonical-vid": {
						type: ["null", "integer"],
					},
					"internal-list-id": {
						type: ["null", "integer"],
					},
					"is-member": {
						type: ["null", "boolean"],
					},
					"static-list-id": {
						type: ["null", "integer"],
					},
					timestamp: {
						type: ["null", "integer"],
					},
					vid: {
						type: ["null", "integer"],
					},
				},
			},
			name: "contacts_list_memberships",
			source_defined_cursor: false,
			supported_sync_modes: ["full_refresh"],
		},
	},
	{
		sync_mode: "cdc",
		cursor_field: ["updatedAt"],
		destination_sync_mode: "overwrite",
		selected_columns: null,
		sort_key: null,
		stream: {
			default_cursor_field: ["updatedAt"],
			json_schema: {
				properties: {
					archived: {
						type: ["null", "boolean"],
					},
					companies: {
						type: ["null", "array"],
					},
					contacts: {
						type: ["null", "array"],
					},
					createdAt: {
						format: "date-time",
						type: ["null", "string"],
					},
					id: {
						type: ["null", "string"],
					},
					line_items: {
						type: ["null", "array"],
					},
					properties: {
						properties: {
							amount: {
								type: ["null", "number"],
							},
						},
						type: "object",
					},
					updatedAt: {
						format: "date-time",
						type: ["null", "string"],
					},
				},
			},
			name: "deals",
			source_defined_cursor: true,
			source_defined_primary_key: [["id"]],
			supported_sync_modes: ["full_refresh", "incremental"],
		},
	},
]

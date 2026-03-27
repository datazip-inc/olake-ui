import type { SpecResponse } from "@/common/types"

import type { Catalog, CatalogFormData, GetCatalogsResponse } from "../types"

/** RJSF / spec expect `type` + nested `writer`; GET catalog returns the writer body only. */
const CATALOG_FORM_TYPE_ICEBERG = "ICEBERG"

/**
 * Normalizes GET /fusion/catalog/:name payload for the form.
 * Pass-through when `writer` is already nested; otherwise wraps flat writer fields.
 */
export const mapGetCatalogResponseToFormData = (
	data: CatalogFormData,
): CatalogFormData => {
	const writer = (data as { writer?: unknown }).writer
	if (
		writer !== null &&
		writer !== undefined &&
		typeof writer === "object" &&
		!Array.isArray(writer)
	) {
		return data
	}

	return {
		type: CATALOG_FORM_TYPE_ICEBERG,
		writer: { ...(data as Record<string, unknown>) },
	}
}

// Converts GetCatalogsResponse into Catalog rows, uppercasing the type and attaching a stable row id.
export const mapGetCatalogsResponseToCatalogs = (
	response: GetCatalogsResponse,
): Catalog[] => {
	return response.result.map((catalog, idx) => {
		const props = catalog.catalogProperties ?? {}

		return {
			id: `${catalog.catalogName}-${idx}`,
			name: catalog.catalogName,
			type: catalog.catalogType.toUpperCase(),
			createdOn: props["created-at"] ?? props["created_at"] ?? "",
			olakeCreated: props["olake_created"] === "true",
		}
	})
}

// Parses spec.uischema and (in edit mode) disables writer.catalog_name so the catalog name can't be changed while editing the spec.
export const mapCatalogSpecResponse = (
	data: SpecResponse,
	isEditMode: boolean,
): SpecResponse => {
	if (!data?.spec?.uischema) return data

	try {
		const uiSchema = JSON.parse(data.spec.uischema)
		if (isEditMode) {
			const finalUiSchema = {
				...uiSchema,
				writer: {
					...(uiSchema.writer || {}),
					catalog_name: {
						...(uiSchema.writer?.catalog_name || {}),
						"ui:disabled": true,
					},
				},
			}
			return {
				...data,
				spec: {
					...data.spec,
					uischema: JSON.stringify(finalUiSchema),
				},
			}
		}
	} catch (e) {
		console.error("Failed to parse uiSchema in mapCatalogSpecResponse", e)
		throw new Error("Failed to parse uiSchema")
	}

	return data
}

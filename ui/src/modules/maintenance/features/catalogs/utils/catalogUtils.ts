import type { Catalog, GetCatalogsResponse } from "../types"

// Converts GetCatalogsResponse into Catalog rows, uppercasing the type and attaching a stable row id.
export const mapGetCatalogsResponseToCatalogs = (
	response: GetCatalogsResponse,
): Catalog[] => {
	return response.catalogs.map((catalog, idx) => ({
		id: `${catalog.name}-${idx}`,
		name: catalog.name,
		type: catalog.type.toUpperCase(),
		databases: catalog.databases,
		createdOn: catalog.created_on,
	}))
}

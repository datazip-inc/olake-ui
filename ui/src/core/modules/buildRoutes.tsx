import { createElement } from "react"
import type { RouteObject } from "react-router-dom"

import type { AppModule } from "./types"

/**
 * Converts a list of AppModule descriptors into a flat array of
 * React Router RouteObjects suitable for use as children of the
 * root protected route.
 *
 * Modules with a `gate` component are wrapped:
 *   { element: <GateComponent />, children: module.routes }
 *
 * Modules without a gate have their routes spread directly into
 * the array.
 */
export function buildProtectedChildren(modules: AppModule[]): RouteObject[] {
	return modules.flatMap(({ routes, gate }) => {
		if (gate) {
			return [{ element: createElement(gate), children: routes }]
		}

		return routes
	})
}

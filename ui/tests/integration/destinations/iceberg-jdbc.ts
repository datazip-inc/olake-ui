/** The Iceberg destination with a JDBC catalog the Go test starts (server/tests/destinations/jdbc). */

import { Page } from "@playwright/test"

import { createDestination } from "./destination"
import { CreateDestinationPage } from "../../pages/CreateDestinationPage"
import { DestinationsPage } from "../../pages/DestinationsPage"
import { createIcebergJdbcConfig, ICEBERG_JDBC_CONFIG } from "../../utils"

/** Creates the Iceberg (JDBC catalog) destination the Go test started. */
export const createIcebergJdbcDestination = (
	page: Page,
	destinationsPage: DestinationsPage,
	createDestinationPage: CreateDestinationPage,
	name: string,
) =>
	createDestination(
		page,
		destinationsPage,
		createDestinationPage,
		createIcebergJdbcConfig({ name, ...ICEBERG_JDBC_CONFIG }),
	)

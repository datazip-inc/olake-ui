/**
 * The MongoDB source the Go test starts (server/tests/sources/mongodb): a standalone server, so
 * CDC, which needs a replica set, is off.
 */

import { Page } from "@playwright/test"

import { createSource, setSourceCdc } from "./source"
import { SourceConnector } from "../../enums"
import { CreateSourcePage } from "../../pages/CreateSourcePage"
import { SourcesPage } from "../../pages/SourcesPage"
import { SERVICE_HOST } from "../../utils"

/** Creates the MongoDB source the Go test started, with CDC off. */
export const createMongoDBSource = (
	page: Page,
	sourcesPage: SourcesPage,
	createSourcePage: CreateSourcePage,
	name: string,
) =>
	createSource(
		page,
		sourcesPage,
		createSourcePage,
		{
			name,
			connector: SourceConnector.MongoDB,
			fields: {
				hosts: [`${SERVICE_HOST}:27017`],
				database: "test_db",
				authdb: "admin",
				username: "admin",
				password: "password",
			},
		},
		() => setSourceCdc(page, false),
	)

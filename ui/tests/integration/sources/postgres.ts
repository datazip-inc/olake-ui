/** The Postgres source the Go test starts (server/tests/sources/postgres). */

import { Page } from "@playwright/test"

import { createSource } from "./source"
import { expect } from "../../fixtures"
import { CreateSourcePage } from "../../pages/CreateSourcePage"
import { SourcesPage } from "../../pages/SourcesPage"
import { createPostgresSourceConfig, POSTGRES_SOURCE_CONFIG } from "../../utils"

/** The replication slot and publication the Go test's Postgres creates. */
export const CDC = {
	slot: "olake_slot",
	publication: "olake_publication",
	initialWaitTime: "120", // the minimum; lower values silently become 10800
}

/** Turns the source form's CDC on or off: the "Enable CDC" switch, or the older radios. */
export const setSourceCdc = async (page: Page, on: boolean) => {
	const cdcSwitch = page.getByRole("switch", { name: "Enable CDC" })
	if (await cdcSwitch.isVisible()) {
		if ((await cdcSwitch.getAttribute("aria-checked")) !== String(on)) {
			await cdcSwitch.click()
		}
		await expect(cdcSwitch).toHaveAttribute("aria-checked", String(on))
		return
	}
	await page.getByRole("radio", { name: on ? "CDC" : "Standalone" }).check()
}

/** Creates the Postgres source the Go test started, with CDC on or off. */
export const createPostgresSource = async (
	page: Page,
	sourcesPage: SourcesPage,
	createSourcePage: CreateSourcePage,
	name: string,
	cdc: boolean,
) => {
	await createSource(
		page,
		sourcesPage,
		createSourcePage,
		createPostgresSourceConfig({ name, ...POSTGRES_SOURCE_CONFIG }),
		async () => {
			await setSourceCdc(page, cdc)
			if (cdc) {
				await page
					.locator("#root_update_method_replication_slot")
					.fill(CDC.slot)
				await page
					.locator("#root_update_method_publication")
					.fill(CDC.publication)
				await page
					.locator("#root_update_method_initial_wait_time")
					.fill(CDC.initialWaitTime)
			}
		},
	)
}

import { execFileSync } from "node:child_process"
import http from "node:http"

import { Page } from "@playwright/test"

import { SourceConnector, DestinationConnector } from "../enums"
import { testAuthenticated as test, expect } from "../fixtures"
import { CreateJobPage } from "../pages/CreateJobPage"
import {
	createPostgresSourceConfig,
	POSTGRES_SOURCE_CONFIG,
	createIcebergJdbcConfig,
	ICEBERG_JDBC_CONFIG,
	verifyEntityCreationSuccessModal,
	E2E_HOST,
} from "../utils"

// Names are shared by the serial tests below. Set E2E_RUN_ID to rerun against an OLake that
// already holds the entities of an earlier run.
const RUN = process.env.E2E_RUN_ID ?? "ci"
const SOURCE = `e2e_pg_source_${RUN}`
const SOURCE_NO_CDC = `e2e_pg_nocdc_source_${RUN}`
const DESTINATION = `e2e_iceberg_dest_${RUN}`
const FEATURES_JOB = `e2e_features_job_${RUN}`
// the Go checks read the tables from the database with this prefix
const DATABASE_PREFIX = "e2e_features"
const FAILED_JOB = `e2e_failed_job_${RUN}`
const WEBHOOK_PORT = 18099

// Runs SQL in the source container, which both the DinD box and a local setup name this way.
const psql = (sql: string) =>
	execFileSync("docker", [
		"exec",
		"olake_postgres-test",
		"psql",
		"-U",
		"postgres",
		"-v",
		"ON_ERROR_STOP=1",
		"-c",
		sql,
	])

// Collects the bodies of the webhook alerts that OLake posts.
const startWebhookReceiver = async () => {
	const bodies: string[] = []
	const server = http.createServer((req, res) => {
		let body = ""
		req.on("data", chunk => (body += chunk))
		req.on("end", () => {
			bodies.push(body)
			res.end("ok")
		})
	})
	await new Promise<void>(resolve =>
		server.listen(WEBHOOK_PORT, "0.0.0.0", resolve),
	)
	return { bodies, close: () => server.close() }
}

// Creates a job over one stream on the CDC-less path: a full refresh that appends.
const createFullRefreshJob = async (
	createJobPage: CreateJobPage,
	jobName: string,
	stream: string,
) => {
	await openStreamsStep(createJobPage, jobName, SOURCE)
	await createJobPage.selectOnlyStreams([stream])
	await createJobPage.openStream(stream)
	await createJobPage.selectSyncMode("Full Refresh")
	await createJobPage.selectIngestionMode("Append")
	await createJobPage.setDestinationDatabase(DATABASE_PREFIX)
	await createJobPage.createJob()
	await createJobPage.goToJobsPage()
}

// Opens the job wizard on its streams step for the source, optionally with target query engines.
const openStreamsStep = (
	createJobPage: CreateJobPage,
	jobName: string,
	source: string,
	queryEngine?: string,
) =>
	createJobPage.openStreamsStep({
		jobName,
		source,
		sourceConnector: SourceConnector.Postgres,
		destination: DESTINATION,
		destinationConnector: DestinationConnector.ApacheIceberg,
		queryEngine,
	})

const upsertTypes = async (page: Page, createJobPage: CreateJobPage) => {
	await createJobPage.upsertTypeSelect().click()
	const options = await page
		.locator(".ant-select-dropdown:visible .ant-select-item-option")
		.allInnerTexts()
	await page.keyboard.press("Escape")
	return options
}

test.describe("Postgres to Iceberg JDBC end-to-end", () => {
	test.describe.configure({ mode: "serial" })

	test("create a CDC source, a source without CDC and the destination", async ({
		sourcesPage,
		createSourcePage,
		destinationsPage,
		createDestinationPage,
		page,
	}) => {
		for (const [name, cdc] of [
			[SOURCE, true],
			[SOURCE_NO_CDC, false],
		] as const) {
			await sourcesPage.goto()
			await sourcesPage.clickCreateSource()
			await createSourcePage.expectCreateSourcePageVisible()
			// new sources start with CDC on and its replication slot and publication required
			const config = createPostgresSourceConfig({
				name,
				...POSTGRES_SOURCE_CONFIG,
			})
			if (cdc) {
				config.fields.update_method = {
					replication_slot: "olake_slot",
					publication: "olake_publication",
				}
			}
			await createSourcePage.fillSourceForm(config)
			if (!cdc) await page.getByRole("switch", { name: "Enable CDC" }).click()
			await createSourcePage.clickCreate()
			await verifyEntityCreationSuccessModal(createSourcePage)
			await sourcesPage.expectSourceExists(name)
		}

		await destinationsPage.goto()
		await destinationsPage.clickCreateDestination()
		await createDestinationPage.expectCreateDestinationPageVisible()
		await createDestinationPage.fillDestinationForm(
			createIcebergJdbcConfig({ name: DESTINATION, ...ICEBERG_JDBC_CONFIG }),
		)
		await createDestinationPage.clickCreate()
		await verifyEntityCreationSuccessModal(createDestinationPage)
		await destinationsPage.expectDestinationExists(DESTINATION)
	})

	test("stream rules: CDC modes follow the source", async ({
		page,
		createJobPage,
	}) => {
		await openStreamsStep(createJobPage, "e2e_rules_job", SOURCE_NO_CDC)
		await createJobPage.openStream("e2e_cdc")
		for (const mode of ["Full Refresh + CDC", "CDC Only"]) {
			await expect(
				page.getByRole("radio", { name: mode, exact: true }),
			).toBeDisabled()
		}
		await expect(
			page.getByRole("radio", { name: "Full Refresh", exact: true }),
		).toBeEnabled()
	})

	test("stream rules: target query engines decide the upsert types", async ({
		page,
		createJobPage,
	}) => {
		await openStreamsStep(createJobPage, "e2e_rules_job", SOURCE, "Snowflake")
		await createJobPage.openStream("e2e_upsert_eq")
		expect(await upsertTypes(page, createJobPage)).toEqual([
			"Positional Deletes",
			"Deletion Vector",
		])

		await openStreamsStep(
			createJobPage,
			"e2e_rules_job",
			SOURCE,
			"Apache Spark",
		)
		await createJobPage.openStream("e2e_upsert_eq")
		expect(await upsertTypes(page, createJobPage)).toEqual([
			"Equality Deletes",
			"Positional Deletes",
			"Deletion Vector",
		])
	})

	test("stream rules: global ingestion mode, bulk configure and the index warning", async ({
		page,
		createJobPage,
	}) => {
		await openStreamsStep(createJobPage, "e2e_rules_job", SOURCE)
		await createJobPage.openStream("e2e_upsert_eq")

		// All Append asks first, then switches every stream
		await page.getByText("All Append", { exact: true }).click()
		const modal = page.getByRole("dialog")
		await expect(modal).toContainText("for all tables")
		await modal.getByRole("button", { name: "Confirm" }).click()
		await expect(
			page.getByRole("radio", { name: "Append", exact: true }),
		).toBeChecked()
		await expect(createJobPage.upsertTypeSelect()).toHaveCount(0)
		await page.getByText("All Upsert", { exact: true }).click()
		await modal.getByRole("button", { name: "Confirm" }).click()
		await expect(
			page.getByRole("radio", { name: "Upsert", exact: true }),
		).toBeChecked()

		// Bulk configure only enables Review Changes once a property is changed, then applies it to the streams
		await page.getByRole("button", { name: "Bulk Configure" }).click()
		const bulk = page
			.getByRole("dialog")
			.filter({ hasText: "Bulk Streams Configure" })
		await bulk.getByText("e2e_cdc", { exact: true }).click()
		await bulk.getByText("e2e_strict_cdc", { exact: true }).click()
		await bulk.getByRole("button", { name: "Configure Streams" }).click()
		await expect(bulk).toContainText("Streams Selected (2)")
		const review = bulk.getByRole("button", { name: "Review Changes" })
		await expect(review).toBeDisabled()
		await createJobPage.sectionSwitch("Normalization", bulk).click()
		await expect(review).toBeEnabled()
		await review.click()
		await bulk.getByRole("button", { name: "Apply Changes" }).click()
		await expect(bulk).toBeHidden()
		for (const [stream, normalized] of [
			["e2e_cdc", false],
			["e2e_strict_cdc", false],
			["e2e_upsert_eq", true],
		] as const) {
			await createJobPage.openStream(stream)
			const normalization = createJobPage.sectionSwitch("Normalization")
			if (normalized) await expect(normalization).toBeChecked()
			else await expect(normalization).not.toBeChecked()
		}

		// A deletion vector stream builds an index, so creating the job asks to confirm
		await createJobPage.openStream("e2e_upsert_dv")
		await createJobPage.selectUpsertType("Deletion Vector")
		await createJobPage.createJob()
		const warning = page
			.getByRole("dialog")
			.filter({ hasText: "Please Confirm" })
		await expect(warning).toContainText("Indexes might get built")
		await warning.getByRole("button", { name: "Cancel" }).click()
		await expect(warning).toBeHidden()
	})

	test("create the features job", async ({ page, createJobPage, jobsPage }) => {
		await openStreamsStep(createJobPage, FEATURES_JOB, SOURCE)
		await createJobPage.selectOnlyStreams([
			"e2e_cdc",
			"e2e_strict_cdc",
			"e2e_upsert_eq",
			"e2e_upsert_dv",
			"e2e_filter",
			"e2e_new_columns",
		])

		await createJobPage.openStream("e2e_cdc")
		await createJobPage.selectSyncMode("Full Refresh + CDC")
		await createJobPage.selectIngestionMode("Append")

		await createJobPage.openStream("e2e_strict_cdc")
		await createJobPage.selectSyncMode("CDC Only")
		await createJobPage.selectIngestionMode("Append")

		await createJobPage.openStream("e2e_upsert_eq")
		await createJobPage.selectSyncMode("Full Refresh + CDC")
		await createJobPage.selectIngestionMode("Upsert")
		await createJobPage.selectUpsertType("Equality Deletes")

		await createJobPage.openStream("e2e_upsert_dv")
		await createJobPage.selectSyncMode("Full Refresh + CDC")
		await createJobPage.selectIngestionMode("Upsert")
		await createJobPage.selectUpsertType("Deletion Vector")

		// a full refresh re-reads the source, so a row that starts matching the filter shows up
		await createJobPage.openStream("e2e_filter")
		await createJobPage.selectSyncMode("Full Refresh")
		await createJobPage.selectIngestionMode("Upsert")
		await createJobPage.setDataFilter([
			{ column: "amount", operator: ">", value: "20" },
			{ column: "id", operator: "<", value: "5" },
		])

		await createJobPage.openStream("e2e_new_columns")
		await createJobPage.selectSyncMode("Full Refresh")
		await createJobPage.selectIngestionMode("Upsert")
		await createJobPage.openSchemaTab()
		await expect(createJobPage.syncNewColumnsSwitch()).toBeChecked()

		await createJobPage.setDestinationDatabase(DATABASE_PREFIX)
		await expect(page.getByText(`${DATABASE_PREFIX}_`).first()).toBeVisible()

		await createJobPage.createJob()
		await createJobPage.confirmIndexBuildWarning()
		await createJobPage.goToJobsPage()
		await jobsPage.expectJobExists(FEATURES_JOB)
	})

	test("first sync of the features job", async ({ jobsPage }) => {
		await jobsPage.syncAndWait(FEATURES_JOB)
	})

	test("task logs: run summary, search, refresh and the errors filter", async ({
		jobsPage,
	}) => {
		await jobsPage.goto()
		await jobsPage.openHistory(FEATURES_JOB)
		await jobsPage.viewJobLogs()
		await jobsPage.expectLogSummary()

		// the search keeps only the lines that contain the text
		await jobsPage.searchLogs("Total records read")
		await jobsPage.expectLogSummary()
		await jobsPage.searchLogs("no log line contains this")
		await expect(jobsPage.logSummary).toHaveCount(0)
		await jobsPage.searchLogs("")

		// refetching keeps the logs, and a clean run has no error lines
		await jobsPage.refreshLogs()
		await jobsPage.expectLogSummary()
		await jobsPage.toggleErrorsOnly()
		await expect(jobsPage.logSummary).toHaveCount(0)
		await jobsPage.toggleErrorsOnly()
		await jobsPage.expectLogSummary()
	})

	test("a second sync picks up the source changes", async ({ jobsPage }) => {
		// insert, update and delete, rows arriving later, a row that starts matching the filter, a new column
		psql(`
			INSERT INTO e2e_cdc VALUES (6,'f',60);
			UPDATE e2e_cdc SET amount = 99 WHERE id = 1;
			DELETE FROM e2e_cdc WHERE id = 2;
			INSERT INTO e2e_strict_cdc VALUES (1,'a',10),(2,'b',20);
			UPDATE e2e_upsert_eq SET amount = 99 WHERE id = 1;
			DELETE FROM e2e_upsert_eq WHERE id = 2;
			UPDATE e2e_upsert_dv SET amount = 99 WHERE id = 1;
			DELETE FROM e2e_upsert_dv WHERE id = 2;
			UPDATE e2e_filter SET amount = 25 WHERE id = 2;
			INSERT INTO e2e_filter VALUES (7,'g',70);
			ALTER TABLE e2e_new_columns ADD COLUMN extra TEXT;
			INSERT INTO e2e_new_columns VALUES (3,'c',30,'new');
		`)
		await jobsPage.syncAndWait(FEATURES_JOB)
	})

	test("clear destination drops every table the job wrote", async ({
		jobsPage,
	}) => {
		await jobsPage.goto()
		await jobsPage.clearDestination(FEATURES_JOB)
		await jobsPage.waitForRun("Clear Destination", "Completed")
	})

	test("the next sync rebuilds every table in full", async ({ jobsPage }) => {
		await jobsPage.syncAndWait(FEATURES_JOB)
	})

	test("webhook alerts can be saved, cleared and saved again", async ({
		settingsPage,
	}) => {
		const url = `http://${E2E_HOST}:${WEBHOOK_PORT}`
		await settingsPage.goto()
		await settingsPage.saveWebhook(url)
		await settingsPage.clearWebhook()
		await settingsPage.saveWebhook(url)
	})

	test("a failed sync is listed under failed jobs and alerts the webhook", async ({
		createJobPage,
		jobsPage,
		settingsPage,
	}) => {
		await createFullRefreshJob(createJobPage, FAILED_JOB, "e2e_failed")

		// the stream's table is gone at the source, so the sync cannot read it
		psql("DROP TABLE e2e_failed")
		const receiver = await startWebhookReceiver()
		try {
			await settingsPage.goto()
			await settingsPage.saveWebhook(`http://${E2E_HOST}:${WEBHOOK_PORT}`)

			await jobsPage.goto()
			await jobsPage.syncJob(FAILED_JOB)
			await jobsPage.waitForRun("Sync", "Failed")

			await jobsPage.goto()
			await jobsPage.switchToFailedTab()
			await jobsPage.expectJobInTab("Failed jobs", FAILED_JOB)
			await expect
				.poll(() => receiver.bodies.join("\n"), { timeout: 60_000 })
				.toContain("Sync Failure Detected!")
		} finally {
			receiver.close()
			await settingsPage.goto()
			await settingsPage.clearWebhook()
		}
	})
})

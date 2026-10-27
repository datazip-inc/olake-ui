/**
 * Postgres → Apache Iceberg (JDBC catalog) integration workflow.
 *
 * Run by TestPostgres in server/tests, one tagged test at a time and in this order, between the
 * Go steps that seed and change the source and check the destination through Spark:
 *
 *   @create-job        CDC source, destination, and one job with a stream per feature
 *   @first-sync        sync 1                                  (Go checks, then changes the source)
 *   @second-sync       sync 2                                  (Go checks)
 *   @clear-destination Clear Destination from Job settings     (Go checks every table is gone)
 *   @rebuild-sync      sync 3 rebuilds the tables              (Go checks)
 *   @history           history, Job Type, search, refresh, task logs, errors filter, download
 *   @stream-rules      CDC modes off without CDC, Target Query Engines, "All Append",
 *                      Bulk Configure, index-build warning Cancel, wizard Cancel
 *   @alerts            webhook URL: save, edit + cancel, clear, save again
 *   @failed-sync       sync 4 with the source stopped by Go: Failed, "Failed jobs" tab
 *                                                             (Go waits for the webhook)
 *
 * Without E2E_JOB_NAME every test skips. The steps come from the modules next to this file; this
 * file holds what is particular to Postgres → Iceberg: its names, streams and their settings.
 */

import { DestinationConnector, SourceConnector } from "../enums"
import { testAuthenticated as test, expect } from "../fixtures"
import { createIcebergJdbcDestination } from "./destinations/iceberg-jdbc"
import { E2E, SYNC_TIMEOUT } from "./e2e"
import { expectLogTools, expectNewestRunLogs } from "./job/logs"
import {
	expectHistory,
	expectInFailedJobs,
	jobMenu,
	syncAndWait,
} from "./job/runs"
import { clearDestination } from "./job/settings"
import {
	bulkConfigureSyncMode,
	deselectAllStreams,
	offeredUpsertTypes,
	selectStream,
	setDataFilter,
	setIngestionMode,
	setNormalization,
	setSyncMode,
	setSyncNewColumns,
	setUpsertType,
	streamTab,
	switchAllIngestion,
	syncModeRadio,
} from "./job/streams"
import {
	JobConfig,
	cancelAtIndexBuildWarning,
	cancelWizard,
	createJobFromStreams,
	expectDynamicDatabase,
	pickExistingSource,
	startJob,
	toggleQueryEngine,
	wizardBack,
	wizardNext,
} from "./job/wizard"
import { expectWebhookAlerts } from "./settings/alerts"
import { createPostgresSource } from "./sources/postgres"

const job = E2E.job

/** Names of everything the workflow creates, derived from the job name. */
const n = {
	source: `${job}_source`,
	// Unique per run: the source is created in @stream-rules, which can be run again on its own.
	sourceNoCdc: `${job}_source_nocdc_${Date.now()}`,
	destination: `${job}_dest`,
	rulesJob: `${job}_rules`,
	database: E2E.database,
	// The UI shows a Dynamic name as <prefix>_${source_namespace}: the database without its
	// source schema (public).
	databasePrefix: E2E.database.replace(/_public$/, ""),
}

/** The e2e tables the Go test seeds, one stream per feature. */
const STREAMS = {
	cdc: "e2e_cdc",
	strictCdc: "e2e_strict_cdc",
	upsertEq: "e2e_upsert_eq",
	upsertDv: "e2e_upsert_dv",
	filter: "e2e_filter",
	newColumns: "e2e_new_columns",
} as const

/** The job's config step; @stream-rules uses it with its own job and sources. */
const jobConfig: JobConfig = {
	job,
	source: n.source,
	sourceConnector: SourceConnector.Postgres,
	destination: n.destination,
	destinationConnector: DestinationConnector.ApacheIceberg,
}

test.describe("Postgres → Iceberg (JDBC) integration workflow", () => {
	// Go owns retries: a retry would create the job a second time.
	test.describe.configure({ retries: 0 })
	test.skip(
		!job,
		"run by the Go integration test (server/tests), which sets E2E_JOB_NAME",
	)

	// Every test starts on the jobs page, where the sidebar links the steps click are rendered.
	test.beforeEach(async ({ page }) => {
		await page.goto("/jobs")
		await expect(page).toHaveURL(/\/jobs$/)
	})

	test("@create-job: CDC source, destination, and a job with one stream per feature", async ({
		page,
		sourcesPage,
		createSourcePage,
		destinationsPage,
		createDestinationPage,
		jobsPage,
		createJobPage,
	}) => {
		test.setTimeout(20 * 60 * 1000)
		expect(
			E2E.database,
			"E2E_DATABASE: the database the Go checks read",
		).not.toBe("")

		await createPostgresSource(
			page,
			sourcesPage,
			createSourcePage,
			n.source,
			true,
		)
		await createIcebergJdbcDestination(
			page,
			destinationsPage,
			createDestinationPage,
			n.destination,
		)
		await startJob(page, jobsPage, createJobPage, jobConfig)

		// Dynamic database name: every stream's destination table is <job>_postgres_public/<table>.
		await expectDynamicDatabase(page, n.databasePrefix)
		await deselectAllStreams(page)

		const configure: Record<string, () => Promise<void>> = {
			// Full Refresh + CDC, Append: inserts, updates and deletes all land as rows.
			[STREAMS.cdc]: async () => {
				await setSyncMode(page, "Full Refresh + CDC")
				await setIngestionMode(page, "Append")
			},
			// CDC Only, Append: sync 1 writes nothing, sync 2 only the new rows.
			[STREAMS.strictCdc]: async () => {
				await setSyncMode(page, "CDC Only")
				await setIngestionMode(page, "Append")
			},
			// Upsert with equality deletes, Iceberg v2.
			[STREAMS.upsertEq]: async () => {
				await setSyncMode(page, "Full Refresh + CDC")
				await setIngestionMode(page, "Upsert")
				await setUpsertType(page, "Equality Deletes", "V2")
			},
			// Upsert with deletion vectors, Iceberg v3.
			[STREAMS.upsertDv]: async () => {
				await setSyncMode(page, "Full Refresh + CDC")
				await setIngestionMode(page, "Upsert")
				await setUpsertType(page, "Deletion Vector", "V3")
			},
			// Data Filter amount > 20 AND id < 5 (needs Normalization, set below).
			[STREAMS.filter]: async () => {
				await setSyncMode(page, "Full Refresh + CDC")
				await setIngestionMode(page, "Append")
				await setDataFilter(
					page,
					[
						{ column: "amount", operator: ">", value: "20" },
						{ column: "id", operator: "<", value: "5" },
					],
					"AND",
				)
			},
			// Sync new columns automatically on: a column added at the source reaches Iceberg.
			[STREAMS.newColumns]: async () => {
				await setSyncMode(page, "Full Refresh + CDC")
				await setIngestionMode(page, "Append")
				await setSyncNewColumns(page, true)
			},
		}

		for (const stream of Object.values(STREAMS)) {
			await selectStream(page, stream)
			await streamTab(page, "Config")
			await expect(
				page.getByText(`${n.database}/${stream}`, { exact: true }).first(),
			).toBeVisible()
			// Every stream is normalized: the checks read the source columns.
			await setNormalization(page, true)
			await configure[stream]()
		}

		// Deletion Vector needs a destination index, so Create Job asks first.
		await createJobFromStreams(page, createJobPage, jobsPage, job, true)
	})

	test("@first-sync: sync 1", async ({ page, jobsPage }) => {
		test.setTimeout(SYNC_TIMEOUT + 5 * 60 * 1000)
		await syncAndWait(page, jobsPage, job, 1, "Completed")
		await expectNewestRunLogs(page)
		await jobsPage.navigateToJobs()
		await jobsPage.expectJobStatus(job, "Completed")
	})

	test("@second-sync: sync 2, after the source changes", async ({
		page,
		jobsPage,
	}) => {
		test.setTimeout(SYNC_TIMEOUT + 5 * 60 * 1000)
		await syncAndWait(page, jobsPage, job, 2, "Completed")
		await expectNewestRunLogs(page)
	})

	test("@clear-destination: Clear Destination from Job settings", async ({
		page,
		jobsPage,
	}) => {
		test.setTimeout(SYNC_TIMEOUT + 5 * 60 * 1000)
		await clearDestination(page, jobsPage, job, 3)
	})

	test("@rebuild-sync: sync 3 rebuilds the cleared tables", async ({
		page,
		jobsPage,
	}) => {
		test.setTimeout(SYNC_TIMEOUT + 5 * 60 * 1000)
		await syncAndWait(page, jobsPage, job, 4, "Completed")
		await expectNewestRunLogs(page)
	})

	test("@history: runs, Job Type, search, refresh and task logs", async ({
		page,
		jobsPage,
	}) => {
		test.setTimeout(10 * 60 * 1000)
		await jobMenu(page, jobsPage, job, "Job Logs & History")
		await expectHistory(page, ["Sync", "Clear Destination", "Sync", "Sync"])
		await expectNewestRunLogs(page)
		await expectLogTools(page)
	})

	test("@stream-rules: CDC modes, Target Query Engines, All Append, Bulk Configure", async ({
		page,
		sourcesPage,
		createSourcePage,
		jobsPage,
		createJobPage,
	}) => {
		test.setTimeout(25 * 60 * 1000)

		const upsertTypesOf = async (stream: string) => {
			await selectStream(page, stream)
			await streamTab(page, "Config")
			await setSyncMode(page, "Full Refresh + CDC")
			await setIngestionMode(page, "Upsert")
			return offeredUpsertTypes(page)
		}

		// A source with CDC off offers no CDC sync modes.
		await createPostgresSource(
			page,
			sourcesPage,
			createSourcePage,
			n.sourceNoCdc,
			false,
		)
		await startJob(page, jobsPage, createJobPage, {
			...jobConfig,
			job: n.rulesJob,
			source: n.sourceNoCdc,
		})
		await selectStream(page, STREAMS.cdc)
		await streamTab(page, "Config")
		await expect(syncModeRadio(page, "Full Refresh + CDC")).toBeDisabled()
		await expect(syncModeRadio(page, "CDC Only")).toBeDisabled()
		await expect(syncModeRadio(page, "Full Refresh")).toBeEnabled()
		await expect(
			syncModeRadio(page, "Full Refresh + Incremental"),
		).toBeEnabled()

		// With the CDC source the CDC modes are offered.
		await wizardBack(page)
		await pickExistingSource(page, n.source)
		await wizardNext(page, STREAMS.upsertEq)
		await selectStream(page, STREAMS.cdc)
		await streamTab(page, "Config")
		await expect(syncModeRadio(page, "Full Refresh + CDC")).toBeEnabled()
		await expect(syncModeRadio(page, "CDC Only")).toBeEnabled()

		// Target Query Engines decide the upsert types: none → all three, Dremio → Positional,
		// Databricks → Deletion Vector.
		expect(await upsertTypesOf(STREAMS.upsertEq)).toEqual([
			"Equality Deletes",
			"Positional Deletes",
			"Deletion Vector",
		])
		await wizardBack(page)
		await toggleQueryEngine(page, "Dremio")
		await wizardNext(page, STREAMS.upsertEq)
		expect(await upsertTypesOf(STREAMS.upsertEq)).toEqual([
			"Positional Deletes",
		])
		await wizardBack(page)
		await toggleQueryEngine(page, "Dremio")
		await toggleQueryEngine(page, "Databricks")
		await wizardNext(page, STREAMS.upsertEq)
		expect(await upsertTypesOf(STREAMS.upsertEq)).toEqual(["Deletion Vector"])
		await wizardBack(page)
		await toggleQueryEngine(page, "Databricks")
		await wizardNext(page, STREAMS.upsertEq)

		// "All Append" switches every stream to Append; "All Upsert" switches them back.
		await selectStream(page, STREAMS.cdc)
		await selectStream(page, STREAMS.upsertEq)
		await switchAllIngestion(page, "Append", [STREAMS.cdc, STREAMS.upsertEq])
		await switchAllIngestion(page, "Upsert", [STREAMS.cdc, STREAMS.upsertEq])

		// Bulk Configure: Full Refresh for two streams.
		await bulkConfigureSyncMode(
			page,
			[STREAMS.cdc, STREAMS.filter],
			"Full Refresh",
		)

		// The index-build warning: Cancel keeps the wizard open, nothing is created.
		await selectStream(page, STREAMS.upsertDv)
		await streamTab(page, "Config")
		await setSyncMode(page, "Full Refresh + CDC")
		await setIngestionMode(page, "Upsert")
		await setUpsertType(page, "Deletion Vector", "V3")
		await cancelAtIndexBuildWarning(page)

		// Cancel the wizard: the job is never created.
		await cancelWizard(page, n.rulesJob)
	})

	test("@alerts: webhook URL save, edit and cancel, clear, save again", async ({
		page,
	}) => {
		test.skip(!E2E.webhookUrl, "needs E2E_WEBHOOK_URL")
		await expectWebhookAlerts(page, E2E.webhookUrl)
	})

	test("@failed-sync: a sync with the source down fails and shows in Failed jobs", async ({
		page,
		jobsPage,
	}) => {
		test.setTimeout(SYNC_TIMEOUT + 5 * 60 * 1000)
		await syncAndWait(page, jobsPage, job, 5, "Failed")
		await expectInFailedJobs(page, jobsPage, job)
	})
})

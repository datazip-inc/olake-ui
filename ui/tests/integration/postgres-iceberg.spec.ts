/**
 * Postgres → Apache Iceberg (JDBC catalog) integration workflow.
 *
 * Run by TestPostgres in server/tests, one tagged test at a time and in this order, between the
 * Go steps that seed and change the source and check the destination through Spark. Go counts the
 * job's runs and passes E2E_RUNS to the steps that start one, how many runs the job's history
 * lists once it has, and to @history, how many runs there are.
 *
 *   @create-job          CDC source, destination, and one job with a stream per feature
 *   @first-sync          sync 1                                   (Go checks, then changes the source)
 *   @sync                sync 2                                   (Go checks)
 *   @clear-destination   Clear Destination: Cancel, then Confirm; "Editing Disabled" while it runs
 *   @sync                the next sync rebuilds the tables        (Go checks)
 *   @edit-query-engines  Edit Streams: Dremio re-resolves the upsert types; left unsaved
 *   @edit-upsert-type    Edit Streams: e2e_upsert_dv to Positional Deletes clears its table
 *   @sync, @sync         its full load, then CDC with position deletes (Go checks)
 *   @dropped-table       the dropped e2e_dropped fails the sync; Edit Streams removes it
 *   @sync                the next sync completes                  (Go checks)
 *   @stream-rules        CDC modes off without CDC, Target Query Engines, "All Append",
 *                        Bulk Configure, index-build warning Cancel, wizard Cancel
 *   @alerts              webhook URL: validation, save, edit + cancel, clear, save again
 *   @failed-sync         with the source stopped by Go: Failed, its error log, "Failed jobs" tab
 *   @discovery-failure   Edit Streams cannot discover the streams; Retry fails again
 *   @recovery            with the source back: discovery works, the next sync completes (Go checks)
 *   @history             every run over two pages, Job Type, search, refresh, task logs, download
 *
 * Without E2E_JOB_NAME every test skips. The steps come from the modules next to this file; this
 * file holds what is particular to Postgres → Iceberg: its names, streams and their settings.
 */

import { DestinationConnector, SourceConnector } from "../enums"
import { testAuthenticated as test, expect } from "../fixtures"
import { createIcebergJdbcDestination } from "./destinations/iceberg-jdbc"
import { E2E, SYNC_TIMEOUT } from "./e2e"
import {
	editBack,
	expectDiscoveryFailure,
	finishEdit,
	finishWithStreamChange,
	openEditStreams,
	proceedWithChangedQueryEngines,
} from "./job/edit"
import {
	expectErrorLogs,
	expectLogTools,
	expectNewestRunLogs,
	openNewestRunLogs,
} from "./job/logs"
import {
	HistoryRun,
	expectHistory,
	expectInFailedJobs,
	jobMenu,
	syncAndWait,
	waitForNewestRun,
} from "./job/runs"
import { clearDestination } from "./job/settings"
import {
	bulkConfigureSyncMode,
	deselectAllStreams,
	offeredUpsertTypes,
	openStream,
	selectStream,
	setDataFilter,
	setIngestionMode,
	setNormalization,
	setSyncMode,
	setSyncNewColumns,
	setUpsertType,
	streamRow,
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
	frAppend: "e2e_fr_append",
	frUpsert: "e2e_fr_upsert",
	types: "e2e_types",
	noPk: "e2e_no_pk",
	schemaChange: "e2e_schema_change",
	dropped: "e2e_dropped",
} as const

/** The job's runs, newest first, once TestPostgres has run every step before @history. */
const HISTORY: HistoryRun[] = [
	{ type: "Sync", status: "Completed" }, // @recovery
	{ type: "Sync", status: "Failed" }, // @failed-sync, with the source down
	{ type: "Sync", status: "Completed" }, // after the dropped table's stream is gone
	{ type: "Sync", status: "Failed" }, // @dropped-table
	{ type: "Sync", status: "Completed" }, // position deletes
	{ type: "Sync", status: "Completed" }, // e2e_upsert_dv's full load again
	{ type: "Clear Destination", status: "Completed" }, // @edit-upsert-type
	{ type: "Sync", status: "Completed" }, // the rebuild
	{ type: "Clear Destination", status: "Completed" }, // @clear-destination
	{ type: "Sync", status: "Completed" }, // sync 2
	{ type: "Sync", status: "Completed" }, // @first-sync
]

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
			// Full Refresh alone: every sync drops the table and reads the source again.
			[STREAMS.frAppend]: async () => {
				await setSyncMode(page, "Full Refresh")
				await setIngestionMode(page, "Append")
			},
			[STREAMS.frUpsert]: async () => {
				await setSyncMode(page, "Full Refresh")
				await setIngestionMode(page, "Upsert")
			},
			// A column per type, compared row by row with the source.
			[STREAMS.types]: async () => {
				await setSyncMode(page, "Full Refresh + CDC")
				await setIngestionMode(page, "Append")
			},
			// No primary key; its replica identity is FULL.
			[STREAMS.noPk]: async () => {
				await setSyncMode(page, "Full Refresh + CDC")
				await setIngestionMode(page, "Append")
			},
			// amount becomes BIGINT and name is dropped at the source before sync 2.
			[STREAMS.schemaChange]: async () => {
				await setSyncMode(page, "Full Refresh + CDC")
				await setIngestionMode(page, "Append")
			},
			// Dropped at the source before sync 2.
			[STREAMS.dropped]: async () => {
				await setSyncMode(page, "Full Refresh + CDC")
				await setIngestionMode(page, "Append")
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
		await syncAndWait(page, jobsPage, job, E2E.runs, "Completed")
		await expectNewestRunLogs(page)
		await jobsPage.navigateToJobs()
		await jobsPage.expectJobStatus(job, "Completed")
	})

	test("@sync: Sync now, the run completes and logs its end", async ({
		page,
		jobsPage,
	}) => {
		test.setTimeout(SYNC_TIMEOUT + 5 * 60 * 1000)
		await syncAndWait(page, jobsPage, job, E2E.runs, "Completed")
		await expectNewestRunLogs(page)
	})

	test("@clear-destination: Clear Destination from Job settings", async ({
		page,
		jobsPage,
	}) => {
		test.setTimeout(SYNC_TIMEOUT + 5 * 60 * 1000)
		await clearDestination(page, jobsPage, job, E2E.runs)
	})

	test("@edit-query-engines: Edit Streams, Target Query Engines resolve the upsert types again", async ({
		page,
		jobsPage,
	}) => {
		test.setTimeout(15 * 60 * 1000)
		await openEditStreams(page, jobsPage, job, STREAMS.upsertEq)
		await editBack(page)
		await toggleQueryEngine(page, "Dremio")
		await proceedWithChangedQueryEngines(page, STREAMS.upsertEq)
		// Dremio reads position deletes only, so both upsert streams are offered just those now.
		for (const stream of [STREAMS.upsertEq, STREAMS.upsertDv]) {
			await openStream(page, stream)
			await streamTab(page, "Config")
			expect(await offeredUpsertTypes(page)).toEqual(["Positional Deletes"])
		}
		// Leave without saving: the job keeps its streams as they are.
		await page.goto("/jobs")
	})

	test("@edit-upsert-type: Edit Streams, an upsert type change clears the stream's table", async ({
		page,
		jobsPage,
	}) => {
		test.setTimeout(SYNC_TIMEOUT + 5 * 60 * 1000)
		// Deletion Vector (Iceberg v3) to Positional Deletes (v2): the table has to be written anew.
		await openEditStreams(page, jobsPage, job, STREAMS.upsertDv)
		await openStream(page, STREAMS.upsertDv)
		await streamTab(page, "Config")
		await setUpsertType(page, "Positional Deletes", "V2")
		await finishWithStreamChange(page, STREAMS.upsertDv)
		await jobMenu(page, jobsPage, job, "Job Logs & History")
		await waitForNewestRun(page, E2E.runs, "Clear Destination", "Completed")
	})

	test("@dropped-table: a table dropped at the source fails the sync until its stream is gone", async ({
		page,
		jobsPage,
	}) => {
		test.setTimeout(SYNC_TIMEOUT + 10 * 60 * 1000)
		// The publication no longer covers the dropped table, and the log says so.
		await syncAndWait(page, jobsPage, job, E2E.runs, "Failed")
		await expectNewestRunLogs(page, "NOT tracked by publication")
		// Edit Streams discovers the source again, which has no such table, and Finish saves that.
		await openEditStreams(page, jobsPage, job, STREAMS.cdc)
		await expect(streamRow(page, STREAMS.dropped)).toHaveCount(0)
		await finishEdit(page)
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

	test("@alerts: webhook URL validation, save, edit and cancel, clear, save again", async ({
		page,
	}) => {
		test.skip(!E2E.webhookUrl, "needs E2E_WEBHOOK_URL")
		await expectWebhookAlerts(page, E2E.webhookUrl)
	})

	test("@failed-sync: a sync with the source down fails, logs why, and shows in Failed jobs", async ({
		page,
		jobsPage,
	}) => {
		test.setTimeout(SYNC_TIMEOUT + 5 * 60 * 1000)
		await syncAndWait(page, jobsPage, job, E2E.runs, "Failed")
		await openNewestRunLogs(page)
		await expectErrorLogs(page)
		await expectInFailedJobs(page, jobsPage, job)
	})

	test("@discovery-failure: Edit Streams with the source down cannot discover the streams", async ({
		page,
		jobsPage,
	}) => {
		test.setTimeout(15 * 60 * 1000)
		await expectDiscoveryFailure(page, jobsPage, job)
		await page.goto("/jobs")
	})

	test("@recovery: with the source back, discovery works and the next sync completes", async ({
		page,
		jobsPage,
	}) => {
		test.setTimeout(SYNC_TIMEOUT + 10 * 60 * 1000)
		await openEditStreams(page, jobsPage, job, STREAMS.cdc)
		await page.goto("/jobs")
		await syncAndWait(page, jobsPage, job, E2E.runs, "Completed")
		await expectNewestRunLogs(page)
	})

	test("@history: runs over two pages, Job Type, search, refresh and task logs", async ({
		page,
		jobsPage,
	}) => {
		test.setTimeout(10 * 60 * 1000)
		expect(
			HISTORY,
			"HISTORY lists every run TestPostgres started",
		).toHaveLength(E2E.runs)
		await jobMenu(page, jobsPage, job, "Job Logs & History")
		await expectHistory(page, HISTORY)
		await expectNewestRunLogs(page)
		await expectLogTools(page)
	})
})

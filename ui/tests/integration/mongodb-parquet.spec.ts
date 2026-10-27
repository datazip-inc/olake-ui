/**
 * MongoDB → Amazon S3 (parquet) integration workflow.
 *
 * Run by TestMongoDB in server/tests, one tagged test at a time and in this order, between the
 * Go steps that change the source and check the parquet files through Spark. Go passes E2E_RUNS
 * to the steps that start a run: how many runs the job's history lists once it has.
 *
 *   @create-job          MongoDB source (no CDC), S3 destination, a job with one stream per feature
 *                        into the custom S3 folder E2E_DATABASE
 *   @stream-rules        in a wizard that is cancelled: S3 rules, Search Streams and its filters,
 *                        select-all per namespace, a single data filter's conditions, Bulk Configure
 *                        without a common cursor, and its data filter
 *   @first-sync          sync 1                                         (Go checks, then changes the source)
 *   @edit-streams        Edit Streams: Back and stay, a new frequency, re-discovery, add a stream,
 *                        deselect a column, Confirm and Finish
 *   @sync                sync 2                                         (Go checks)
 *   @edit-stream-change  Edit Streams: Normalization on for e2e_norm_off clears only its folder
 *   @sync                its full load again                            (Go checks)
 *   @deselect-stream     Edit Streams: e2e_full_refresh is deselected
 *   @sync                the next sync leaves it alone                  (Go checks)
 *   @schedule            every minute: a scheduled run, Pause (Inactive jobs, Clear Destination
 *                        disabled, no run), Resume (Active jobs, a run again)
 *
 * Without E2E_JOB_NAME every test skips.
 */

import { DestinationConnector, SourceConnector } from "../enums"
import { testAuthenticated as test, expect } from "../fixtures"
import { createS3Destination } from "./destinations/s3"
import { E2E, SYNC_TIMEOUT } from "./e2e"
import {
	editBack,
	editBackAndStay,
	editNext,
	finishEdit,
	finishWithStreamChange,
	openEditStreams,
} from "./job/edit"
import { expectNewestRunLogs } from "./job/logs"
import {
	HISTORY_PAGE_SIZE,
	expectJobInTab,
	historyRows,
	jobMenu,
	setJobPaused,
	syncAndWait,
	waitForNewestRun,
} from "./job/runs"
import { expectClearDestinationDisabled, setJobFrequency } from "./job/settings"
import {
	deselectAllStreams,
	expectAppendOnlyStream,
	expectBulkRules,
	expectFilterConditionRules,
	filterStreams,
	namespaceCheckbox,
	openStream,
	searchStreams,
	streamRow,
	selectStream,
	setColumnSelected,
	setCursor,
	setDataFilter,
	setNormalization,
	setPartitionRegex,
	setSyncMode,
	streamTab,
} from "./job/streams"
import {
	JobConfig,
	cancelWizard,
	createJobFromStreams,
	setCustomDatabase,
	setFrequency,
	startJob,
} from "./job/wizard"
import { createMongoDBSource } from "./sources/mongodb"

const job = E2E.job

/** Names of everything the workflow creates, derived from the job name. */
const n = {
	source: `${job}_source`,
	destination: `${job}_dest`,
	rulesJob: `${job}_rules`,
	// The custom S3 folder every stream is written under; the Go checks read it.
	folder: E2E.database,
}

/** The collections init.js seeds (server/tests/sources/mongodb), one per feature. */
const STREAMS = {
	fullRefresh: "e2e_full_refresh",
	normOff: "e2e_norm_off",
	incremental: "e2e_incremental",
	partition: "e2e_partition",
	columns: "e2e_columns",
	filterOr: "e2e_filter_or",
	filterAnd: "e2e_filter_and",
	partitionTs: "e2e_partition_ts",
	added: "e2e_added",
	intIds: "e2e_int_ids",
} as const

/** The job's config step; @stream-rules uses it with its own job. */
const jobConfig: JobConfig = {
	job,
	source: n.source,
	sourceConnector: SourceConnector.MongoDB,
	destination: n.destination,
	destinationConnector: DestinationConnector.AmazonS3,
}

test.describe("MongoDB → S3 (parquet) integration workflow", () => {
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

	test("@create-job: MongoDB source, S3 destination, and a job with one stream per feature", async ({
		page,
		sourcesPage,
		createSourcePage,
		destinationsPage,
		createDestinationPage,
		jobsPage,
		createJobPage,
	}) => {
		test.setTimeout(20 * 60 * 1000)
		expect(n.folder, "E2E_DATABASE: the S3 folder the Go checks read").not.toBe(
			"",
		)

		await createMongoDBSource(page, sourcesPage, createSourcePage, n.source)
		await createS3Destination(
			page,
			destinationsPage,
			createDestinationPage,
			n.destination,
		)
		await startJob(page, jobsPage, createJobPage, jobConfig)

		// Custom S3 folder: every stream is written to <folder>/<collection>.
		await setCustomDatabase(page, n.folder, "S3 Folder")
		await deselectAllStreams(page)

		const configure: Record<string, () => Promise<void>> = {
			// Full Refresh, Normalization on: the document's fields become columns.
			[STREAMS.fullRefresh]: async () => {
				await setSyncMode(page, "Full Refresh")
				await setNormalization(page, true)
			},
			// Normalization off: the whole document in one data column.
			[STREAMS.normOff]: async () => {
				await setSyncMode(page, "Full Refresh")
				await setNormalization(page, false)
			},
			// Incremental on updated_at, and on created_at where updated_at is null.
			[STREAMS.incremental]: async () => {
				await setSyncMode(page, "Full Refresh + Incremental")
				await setNormalization(page, true)
				await setCursor(page, "updated_at", "created_at")
			},
			// Folders category=a/ and category=b/.
			[STREAMS.partition]: async () => {
				await setSyncMode(page, "Full Refresh")
				await setNormalization(page, true)
				await setPartitionRegex(page, "/category={category,default_category,}")
			},
			// secret is deselected now; note later, in Edit Streams.
			[STREAMS.columns]: async () => {
				await setSyncMode(page, "Full Refresh")
				await setNormalization(page, true)
				await setColumnSelected(page, "secret", false)
			},
			// category = "a" or amount <= 10: = and <= with a quoted string, joined by or.
			[STREAMS.filterOr]: async () => {
				await setSyncMode(page, "Full Refresh + Incremental")
				await setNormalization(page, true)
				await setCursor(page, "updated_at")
				await setDataFilter(
					page,
					[
						{ column: "category", operator: "=", value: "a" },
						{ column: "amount", operator: "<=", value: "10" },
					],
					"OR",
				)
			},
			// amount != 30 and amount >= 20.
			[STREAMS.filterAnd]: async () => {
				await setSyncMode(page, "Full Refresh + Incremental")
				await setNormalization(page, true)
				await setCursor(page, "updated_at")
				await setDataFilter(
					page,
					[
						{ column: "amount", operator: "!=", value: "30" },
						{ column: "amount", operator: ">=", value: "20" },
					],
					"AND",
				)
			},
			// Folders by the year and month of the timestamp event_at.
			[STREAMS.partitionTs]: async () => {
				await setSyncMode(page, "Full Refresh")
				await setNormalization(page, true)
				await setPartitionRegex(
					page,
					"/year={event_at,,YYYY}/month={event_at,,MM}",
				)
			},
		}

		for (const [stream, setUp] of Object.entries(configure)) {
			await selectStream(page, stream)
			await streamTab(page, "Config")
			await expect(
				page.getByText(`${n.folder}/${stream}`, { exact: true }).first(),
			).toBeVisible()
			// S3 has no Ingestion Mode or Upsert Type: parquet files are only appended to.
			await expectAppendOnlyStream(page)
			await setUp()
		}

		await createJobFromStreams(page, createJobPage, jobsPage, job, false)
	})

	test("@first-sync: sync 1", async ({ page, jobsPage }) => {
		test.setTimeout(SYNC_TIMEOUT + 5 * 60 * 1000)
		await syncAndWait(page, jobsPage, job, E2E.runs, "Completed")
		await expectNewestRunLogs(page)
		await jobsPage.navigateToJobs()
		await jobsPage.expectJobStatus(job, "Completed")
	})
	test("@stream-rules: S3 rules, stream search and filters, data filter conditions, Bulk Configure", async ({
		page,
		jobsPage,
		createJobPage,
	}) => {
		test.setTimeout(15 * 60 * 1000)
		await startJob(page, jobsPage, createJobPage, {
			...jobConfig,
			job: n.rulesJob,
		})
		const allStreams = Object.values(STREAMS)

		// Search Streams keeps the streams whose name holds the text.
		await searchStreams(page, "filter")
		for (const stream of allStreams) {
			await expect(streamRow(page, stream)).toHaveCount(
				stream.includes("filter") ? 1 : 0,
			)
		}
		await searchStreams(page, "")

		// The namespace's checkbox selects and deselects every stream at once.
		const namespace = namespaceCheckbox(page, "test_db")
		await namespace.uncheck()
		for (const stream of allStreams) {
			await expect(
				streamRow(page, stream).getByRole("checkbox"),
			).not.toBeChecked()
		}
		await namespace.check()
		for (const stream of allStreams) {
			await expect(streamRow(page, stream).getByRole("checkbox")).toBeChecked()
		}

		// Selected and Not Selected list each side of the selection; All tables lists both.
		await deselectAllStreams(page)
		await selectStream(page, STREAMS.partition)
		await selectStream(page, STREAMS.columns)
		const picked: string[] = [STREAMS.partition, STREAMS.columns]
		await filterStreams(page, "Selected")
		for (const stream of allStreams) {
			await expect(streamRow(page, stream)).toHaveCount(
				picked.includes(stream) ? 1 : 0,
			)
		}
		await filterStreams(page, "Not Selected")
		for (const stream of allStreams) {
			await expect(streamRow(page, stream)).toHaveCount(
				picked.includes(stream) ? 0 : 1,
			)
		}
		await filterStreams(page, "All tables")
		for (const stream of allStreams) {
			await expect(streamRow(page, stream)).toHaveCount(1)
		}

		// A single stream's Data Filter: the second condition waits for the first, and Remove takes it away.
		await openStream(page, STREAMS.partition)
		await streamTab(page, "Config")
		await expectFilterConditionRules(page, {
			column: "category",
			operator: "=",
			value: "a",
		})

		// Bulk Configure over collections with no cursor field of one type in common (an ObjectId
		// _id and an integer _id): Incremental and the data filter are disabled and say why. Over
		// two with an ObjectId _id, the data filter waits for a first condition too.
		await expectBulkRules(
			page,
			[STREAMS.partition, STREAMS.intIds],
			[STREAMS.partition, STREAMS.columns],
		)

		// Cancel the wizard: the job is never created.
		await cancelWizard(page, n.rulesJob)
	})

	test("@edit-streams: Back and stay, a new frequency, re-discovery, a new stream, a deselected column", async ({
		page,
		jobsPage,
	}) => {
		test.setTimeout(SYNC_TIMEOUT + 10 * 60 * 1000)
		await openEditStreams(page, jobsPage, job, STREAMS.columns)
		// The changes of the streams step are lost on Back, so it asks first; No stays.
		await editBackAndStay(page, STREAMS.columns)
		await editBack(page)
		await setFrequency(page, "Every Day")
		// Next discovers the source again.
		await editNext(page, STREAMS.added)

		// e2e_added joins the job, which clears its folder first.
		await selectStream(page, STREAMS.added)
		await streamTab(page, "Config")
		await setSyncMode(page, "Full Refresh")
		await setNormalization(page, true)
		// note leaves e2e_columns; a column change clears nothing.
		await openStream(page, STREAMS.columns)
		await setColumnSelected(page, "note", false)

		await finishWithStreamChange(page, STREAMS.added)
		await jobMenu(page, jobsPage, job, "Job Logs & History")
		await waitForNewestRun(page, E2E.runs, "Clear Destination", "Completed")
	})

	test("@sync: Sync now, the run completes and logs its end", async ({
		page,
		jobsPage,
	}) => {
		test.setTimeout(SYNC_TIMEOUT + 5 * 60 * 1000)
		await syncAndWait(page, jobsPage, job, E2E.runs, "Completed")
		await expectNewestRunLogs(page)
	})

	test("@edit-stream-change: Edit Streams, Normalization on clears only that stream's folder", async ({
		page,
		jobsPage,
	}) => {
		test.setTimeout(SYNC_TIMEOUT + 5 * 60 * 1000)
		await openEditStreams(page, jobsPage, job, STREAMS.normOff)
		await openStream(page, STREAMS.normOff)
		await streamTab(page, "Config")
		await setNormalization(page, true)
		await finishWithStreamChange(page, STREAMS.normOff)
		await jobMenu(page, jobsPage, job, "Job Logs & History")
		await waitForNewestRun(page, E2E.runs, "Clear Destination", "Completed")
	})

	test("@deselect-stream: Edit Streams, a deselected stream leaves the job", async ({
		page,
		jobsPage,
	}) => {
		test.setTimeout(10 * 60 * 1000)
		await openEditStreams(page, jobsPage, job, STREAMS.fullRefresh)
		await streamRow(page, STREAMS.fullRefresh).getByRole("checkbox").uncheck()
		await finishEdit(page)
	})

	test("@schedule: every minute, a paused job runs nothing until it is resumed", async ({
		page,
		jobsPage,
	}) => {
		test.setTimeout(20 * 60 * 1000)
		// E2E_RUNS is how many runs the job has before this step; the schedule adds the others.
		const runs = E2E.runs
		const wait = 4 * 60 * 1000
		await setJobFrequency(page, jobsPage, job, "Every Day", "Every Minute")
		await jobMenu(page, jobsPage, job, "Job Logs & History")
		await waitForNewestRun(page, runs + 1, "Sync", "Completed", wait)

		// Paused: Inactive jobs lists it, Clear Destination is disabled, and no run starts.
		await setJobPaused(page, jobsPage, job, true)
		await expectJobInTab(page, jobsPage, "Inactive jobs", job, true)
		await expectJobInTab(page, jobsPage, "Active jobs", job, false)
		await expectClearDestinationDisabled(page, jobsPage, job)
		await jobMenu(page, jobsPage, job, "Job Logs & History")
		await page.waitForTimeout(90_000)
		await page.getByTestId("refresh-tasks-button").click()
		await expect(historyRows(page)).toHaveCount(
			Math.min(runs + 1, HISTORY_PAGE_SIZE),
		)

		// Resumed: Active jobs lists it again, and the schedule starts a run again.
		await setJobPaused(page, jobsPage, job, false)
		await expectJobInTab(page, jobsPage, "Active jobs", job, true)
		await jobMenu(page, jobsPage, job, "Job Logs & History")
		await waitForNewestRun(page, runs + 2, "Sync", "Completed", wait)
		// Pause it again, so no run starts while the stack is taken down.
		await setJobPaused(page, jobsPage, job, true)
	})
})

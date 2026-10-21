import { execFileSync } from "node:child_process"

import { SourceConnector, DestinationConnector } from "../enums"
import { testAuthenticated as test, expect } from "../fixtures"
import { CreateJobPage } from "../pages/CreateJobPage"
import {
	createMongoDBSourceConfig,
	MONGODB_SOURCE_CONFIG,
	createS3DestinationConfig,
	S3_DESTINATION_CONFIG,
	verifyEntityCreationSuccessModal,
} from "../utils"

// Names are shared by the serial tests below. Set E2E_RUN_ID to rerun against an OLake that
// already holds the entities of an earlier run.
const RUN = process.env.E2E_RUN_ID ?? "ci"
const SOURCE = `e2e_mongo_source_${RUN}`
const DESTINATION = `e2e_s3_dest_${RUN}`
const JOB = `e2e_mongo_parquet_job_${RUN}`
// every table is written to this one S3 folder (custom format); the Go checks read it
const S3_FOLDER = "e2e_mongo_folder"
const PARTITION_REGEX = "/category={category,default_category,}"

// Runs JavaScript in the source container's mongo shell, which both the DinD box and a local setup name this way.
const mongosh = (script: string) =>
	execFileSync("docker", [
		"exec",
		"olake_mongodb-test",
		"mongosh",
		"-u",
		"admin",
		"-p",
		"password",
		"--authenticationDatabase",
		"admin",
		"--quiet",
		"--eval",
		`db = db.getSiblingDB("test_db"); ${script}`,
	])

// Opens the job wizard on its streams step for the MongoDB source and the S3 destination.
const openStreamsStep = (createJobPage: CreateJobPage, jobName: string) =>
	createJobPage.openStreamsStep({
		jobName,
		source: SOURCE,
		sourceConnector: SourceConnector.MongoDB,
		destination: DESTINATION,
		destinationConnector: DestinationConnector.AmazonS3,
	})

test.describe("MongoDB to S3 parquet end-to-end", () => {
	test.describe.configure({ mode: "serial" })

	test("create the MongoDB source and the S3 destination", async ({
		sourcesPage,
		createSourcePage,
		destinationsPage,
		createDestinationPage,
		page,
	}) => {
		await sourcesPage.goto()
		await sourcesPage.clickCreateSource()
		await createSourcePage.expectCreateSourcePageVisible()
		await createSourcePage.fillSourceForm(
			createMongoDBSourceConfig({
				name: SOURCE,
				hosts: [`${MONGODB_SOURCE_CONFIG.host}:${MONGODB_SOURCE_CONFIG.port}`],
				database: MONGODB_SOURCE_CONFIG.database,
				username: MONGODB_SOURCE_CONFIG.username,
				password: MONGODB_SOURCE_CONFIG.password,
			}),
		)
		// CDC needs a replica set, and this job runs full refreshes
		await page.getByRole("switch", { name: "Enable CDC" }).click()
		await createSourcePage.clickCreate()
		await verifyEntityCreationSuccessModal(createSourcePage)
		await sourcesPage.expectSourceExists(SOURCE)

		await destinationsPage.goto()
		await destinationsPage.clickCreateDestination()
		await createDestinationPage.expectCreateDestinationPageVisible()
		await createDestinationPage.fillDestinationForm(
			createS3DestinationConfig({
				name: DESTINATION,
				...S3_DESTINATION_CONFIG,
			}),
		)
		await createDestinationPage.clickCreate()
		await verifyEntityCreationSuccessModal(createDestinationPage)
		await destinationsPage.expectDestinationExists(DESTINATION)
	})

	test("S3 destination rules: upsert is not offered, append is the only mode", async ({
		page,
		createJobPage,
	}) => {
		await openStreamsStep(createJobPage, "e2e_rules_job")
		await createJobPage.openStream("e2e_full_refresh")
		await expect(page.getByText("Sync mode:")).toBeVisible()
		// parquet files cannot be updated in place, so there is no ingestion mode to pick
		await expect(page.getByText("Ingestion Mode:")).toHaveCount(0)
		await expect(page.getByText("Upsert Type:")).toHaveCount(0)
		await expect(page.getByText("All Upsert", { exact: true })).toHaveCount(0)
		await expect(page.getByText("All Append", { exact: true })).toHaveCount(0)
	})

	test("create the job: one collection per behaviour", async ({
		page,
		createJobPage,
		jobsPage,
	}) => {
		await openStreamsStep(createJobPage, JOB)
		await createJobPage.selectOnlyStreams([
			"e2e_full_refresh",
			"e2e_norm_off",
			"e2e_incremental",
			"e2e_partition",
		])

		// normalization on: the document fields become columns; off: one data column.
		// Incremental cursors and partitions need real columns, so those streams have it on.
		for (const [stream, normalized] of [
			["e2e_full_refresh", true],
			["e2e_norm_off", false],
			["e2e_incremental", true],
			["e2e_partition", true],
		] as const) {
			await createJobPage.openStream(stream)
			const normalization = createJobPage.sectionSwitch("Normalization")
			if ((await normalization.isChecked()) !== normalized) {
				await normalization.click()
			}
			if (normalized) await expect(normalization).toBeChecked()
			else await expect(normalization).not.toBeChecked()
		}
		// each stream keeps its own setting
		await createJobPage.openStream("e2e_norm_off")
		await expect(createJobPage.sectionSwitch("Normalization")).not.toBeChecked()

		await createJobPage.openStream("e2e_full_refresh")
		await createJobPage.selectSyncMode("Full Refresh")
		await createJobPage.openStream("e2e_norm_off")
		await createJobPage.selectSyncMode("Full Refresh")

		// rows with no updated_at are picked up by the fallback cursor created_at
		await createJobPage.openStream("e2e_incremental")
		await createJobPage.selectSyncMode("Full Refresh + Incremental")
		await createJobPage.setCursor("updated_at", "created_at")

		await createJobPage.openStream("e2e_partition")
		await createJobPage.selectSyncMode("Full Refresh")
		await createJobPage.setPartitionRegex(PARTITION_REGEX)

		await createJobPage.setDestinationDatabase(S3_FOLDER, "Custom")
		await expect(
			page.getByText(S3_FOLDER, { exact: true }).first(),
		).toBeVisible()
		await createJobPage.createJob()
		await createJobPage.goToJobsPage()
		await jobsPage.expectJobExists(JOB)
	})

	test("first sync", async ({ jobsPage }) => {
		await jobsPage.syncAndWait(JOB)
	})

	test("edit streams: new frequency, re-discover, add a collection, sync", async ({
		createJobPage,
		jobsPage,
	}) => {
		// i4 is picked up by the primary cursor, i5 has no updated_at and is picked up by the fallback
		mongosh(`
			db.e2e_incremental.insertOne({ name: "i4", updated_at: new Date("2025-01-01T00:00:00Z"), created_at: new Date("2025-01-01T00:00:00Z") });
			db.e2e_incremental.insertOne({ name: "i5", updated_at: null, created_at: new Date("2025-01-02T00:00:00Z") });
		`)

		await jobsPage.goto()
		await jobsPage.editStreams(JOB)
		await createJobPage.backToJobConfig()
		await createJobPage.changeFrequency("Every Week", "Every Day")
		// Next discovers the source again, so the streams step lists the current collections
		await createJobPage.nextButton.click()
		await expect(createJobPage.syncAllCheckbox).toBeVisible()

		await createJobPage.streamRow("e2e_added").getByRole("checkbox").check()
		await createJobPage.openStream("e2e_added")
		await createJobPage.selectSyncMode("Full Refresh")
		const normalization = createJobPage.sectionSwitch("Normalization")
		if (!(await normalization.isChecked())) await normalization.click()
		await createJobPage.finishEdit()

		// saving clears the new collection at the destination first; sync once that run is done
		await jobsPage.goto()
		await jobsPage.openHistory(JOB)
		await jobsPage.waitForRun("Clear Destination", "Completed")
		await jobsPage.syncAndWait(JOB)
	})

	test("pause and resume the job", async ({ jobsPage }) => {
		await jobsPage.goto()
		await jobsPage.pauseJob(JOB)
		await jobsPage.switchToInactiveTab()
		await jobsPage.expectJobInTab("Inactive jobs", JOB)
		// a tab lists its jobs when it is opened
		await jobsPage.switchToActiveTab()
		await jobsPage.expectJobNotInTab("Active jobs", JOB)

		await jobsPage.resumeJob(JOB)
		await jobsPage.switchToActiveTab()
		await jobsPage.expectJobInTab("Active jobs", JOB)
	})
})

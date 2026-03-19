import { Buffer } from "node:buffer"
import path from "node:path"
import { fileURLToPath } from "node:url"

import { testAuthenticated as test, expect } from "../fixtures"

const currentDir = path.dirname(fileURLToPath(import.meta.url))
const bundleDataDir = path.resolve(currentDir, "../data/cli-bundles")

const fakeJobs = [
	{
		id: 7,
		name: "orders-zeus-cdc",
		source: {
			id: 11,
			name: "mongodb-orders-zeus",
			type: "mongodb",
			version: "0.1.30",
			config: "{}",
		},
		destination: {
			id: 21,
			name: "parquet-minio-local",
			type: "local",
			version: "0.1.20",
			config: "{}",
		},
		streams_config: "[]",
		frequency: "0 * * * *",
		last_run_type: "sync",
		last_run_state: "success",
		last_run_time: "2026-03-19T18:00:00Z",
		created_at: "2026-03-19T18:00:00Z",
		updated_at: "2026-03-19T18:00:00Z",
		created_by: "1",
		updated_by: "1",
		activate: true,
	},
]

const makeApplyResponse = (bundleName: string) => ({
	dry_run: true,
	prune: false,
	bundle: bundleName,
	effective: {
		apply_identity: bundleName,
		job_name: bundleName,
		source_name: `source-${bundleName}`,
		source_type: "mongodb",
		source_version: "0.1.30",
		destination_name: "parquet-minio-local",
		destination_type: "local",
		destination_version: "0.1.20",
		frequency: "@hourly",
		activate: true,
	},
	source: {
		action: "created",
		name: `source-${bundleName}`,
		fields: ["config"],
	},
	destination: {
		action: "unchanged",
		name: "parquet-minio-local",
		fields: [],
	},
	job: {
		action: "created",
		name: bundleName,
		fields: ["streams_config"],
	},
	state: {
		action: "preserved",
		fields: [],
	},
})

test.describe("CLI bundle jobs UI", () => {
	test("stages multiple folders and applies them in batch", async ({
		jobsPage,
		page,
	}) => {
		let previewCalls = 0
		let applyCalls = 0

		await page.route("**/api/v1/project/123/jobs", async route => {
			if (route.request().method() !== "GET") {
				await route.fallback()
				return
			}

			await route.fulfill({
				status: 200,
				json: {
					data: fakeJobs,
				},
			})
		})

		await page.route("**/api/v1/project/123/jobs/apply-cli-bundle**", async route => {
			if (route.request().method() !== "POST") {
				await route.fallback()
				return
			}

			const isDryRun = route.request().url().includes("dry_run=true")
			if (isDryRun) {
				previewCalls += 1
				const bundleName =
					previewCalls === 1 ? "orders-zeus-cdc" : "store-bodega-cdc"
				await route.fulfill({
					status: 200,
					json: {
						data: makeApplyResponse(bundleName),
					},
				})
				return
			}

			applyCalls += 1
			const bundleName = applyCalls === 1 ? "orders-zeus-cdc" : "store-bodega-cdc"
			await route.fulfill({
				status: 200,
				json: {
					message: "Bundle applied",
					data: {
						...makeApplyResponse(bundleName),
						dry_run: false,
					},
				},
			})
		})

		await jobsPage.goto()
		await jobsPage.expectJobsPageVisible()

		await page.getByRole("button", { name: "Import Bundle" }).click()
		await expect(page.getByText("Queued Imports (0)")).toBeVisible()

		const folderInput = page.getByTestId("cli-bundle-folder-input")
		await folderInput.setInputFiles(
			path.join(bundleDataDir, "orders-zeus-cdc"),
		)
		await expect(
			page.getByTestId("cli-bundle-stage-orders-zeus-cdc"),
		).toBeVisible()

		await folderInput.setInputFiles(
			path.join(bundleDataDir, "store-bodega-cdc"),
		)
		await expect(
			page.getByTestId("cli-bundle-stage-store-bodega-cdc"),
		).toBeVisible()
		await expect(page.getByText("Queued Imports (2)")).toBeVisible()

		await page.getByRole("button", { name: "Preview Import" }).click()
		await expect.poll(() => previewCalls).toBe(2)
		await expect(
			page.getByTestId("cli-bundle-preview-orders-zeus-cdc"),
		).toBeVisible()
		await expect(
			page.getByTestId("cli-bundle-preview-store-bodega-cdc"),
		).toBeVisible()

		await page.getByRole("button", { name: "Import Bundles" }).click()
		await expect.poll(() => applyCalls).toBe(2)
		await expect(page.getByText("Imported 2 bundles")).toBeVisible()
	})

	test("flags incomplete folder imports before any API call", async ({
		jobsPage,
		page,
	}) => {
		await page.route("**/api/v1/project/123/jobs", async route => {
			if (route.request().method() !== "GET") {
				await route.fallback()
				return
			}

			await route.fulfill({
				status: 200,
				json: {
					data: fakeJobs,
				},
			})
		})

		await jobsPage.goto()
		await jobsPage.expectJobsPageVisible()

		await page.getByRole("button", { name: "Import Bundle" }).click()

		await page
			.getByTestId("cli-bundle-folder-input")
			.setInputFiles(path.join(bundleDataDir, "incomplete-bundle"))

		await expect(
			page.getByTestId("cli-bundle-stage-incomplete-bundle"),
		).toBeVisible()
		await expect(
			page.getByText("Missing required files: streams.json"),
		).toBeVisible()
		await expect(
			page.getByRole("button", { name: "Preview Bundle" }),
		).toBeDisabled()
		await expect(
			page.getByRole("button", { name: "Apply Bundle" }),
		).toBeDisabled()
	})

	test("exports a CLI bundle from the job actions menu", async ({
		jobsPage,
		page,
	}) => {
		await page.route("**/api/v1/project/123/jobs", async route => {
			if (route.request().method() !== "GET") {
				await route.fallback()
				return
			}

			await route.fulfill({
				status: 200,
				json: {
					data: fakeJobs,
				},
			})
		})

		await page.route(
			"**/api/v1/project/123/jobs/7/export-cli-bundle**",
			async route => {
				await route.fulfill({
					status: 200,
					headers: {
						"content-type": "application/zip",
						"content-disposition":
							'attachment; filename="orders-zeus-cdc.zip"',
					},
					body: Buffer.from("fake-cli-bundle"),
				})
			},
		)

		await jobsPage.goto()
		await jobsPage.expectJobsPageVisible()

		await page.getByTestId("job-orders-zeus-cdc").click()
		const downloadPromise = page.waitForEvent("download")
		await page.getByText("Export Bundle").click()
		const download = await downloadPromise

		expect(download.suggestedFilename()).toBe("job-7-cli-bundle.zip")
	})
})

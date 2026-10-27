/**
 * Inputs and timeouts shared by the integration workflows in this folder.
 *
 * The Go integration test (server/tests) starts the product, then runs one tagged test of a spec
 * here at a time, with the inputs in E2E_* variables. Selectors follow the components under
 * ui/src; field ids follow RJSF (#root_<field>, nested #root_<parent>_<field>).
 *
 * The UI steps live next to this file by what they act on: common/ (antd controls), sources/,
 * destinations/, job/ (wizard, streams, runs, logs, settings) and settings/ (system settings).
 */

export { TIMEOUTS } from "../../playwright.config"

/**
 * Inputs from the Go integration test (server/tests, Run.Playwright), which runs Playwright as a
 * separate process and passes them as environment variables. A spec skips itself when
 * E2E_JOB_NAME is unset.
 */
export const E2E = {
	job: process.env.E2E_JOB_NAME?.trim() || "",
	// The destination database the Go checks read: <job>_<source database>_<schema>.
	database: process.env.E2E_DATABASE?.trim() || "",
	webhookUrl: process.env.E2E_WEBHOOK_URL?.trim() || "",
}

/** Long waits: a CDC sync waits initial_wait_time for changes after its full load. */
export const SYNC_TIMEOUT = 15 * 60 * 1000

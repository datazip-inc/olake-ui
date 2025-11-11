import { test as baseTest } from "./base.fixture"

const AUTH_STATE_PATH = "./tests/.auth/user.json"

/**
 * Authenticated test - automatically logs in before each test
 */
export const test = baseTest

test.use({ storageState: AUTH_STATE_PATH })

export { expect } from "@playwright/test"

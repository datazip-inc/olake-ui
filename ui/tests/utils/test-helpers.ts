import { Page } from "@playwright/test"
import { LoginPage } from "../pages/LoginPage"
import { TIMEOUTS } from "../../playwright.config"
import { MONGODB_TEST_CONFIG, S3_TEST_CONFIG } from "../setup/test-env"

export class TestHelpers {
	static async loginAsAdmin(page: Page): Promise<void> {
		const loginPage = new LoginPage(page)
		await loginPage.goto()
		await loginPage.login("admin", "password")
		await loginPage.waitForLogin()
	}

	static generateTestSourceData(suffix: string = "") {
		const timestamp = Date.now()
		return {
			name: `test-source-${timestamp}${suffix}`,
			...MONGODB_TEST_CONFIG,
		}
	}

	static generateTestDestinationData(suffix: string = "") {
		const timestamp = Date.now()
		return {
			name: `test-destination-${timestamp}${suffix}`,
			...S3_TEST_CONFIG,
		}
	}

	static generateTestJobData(suffix: string = "") {
		const timestamp = Date.now()
		return {
			name: `test-job-${timestamp}${suffix}`,
			streamName: "posts",
			frequency: "Every Week",
		}
	}

	static async waitForModalToClose(
		page: Page,
		modalSelector: string = ".ant-modal",
	): Promise<void> {
		await page.waitForSelector(modalSelector, {
			state: "hidden",
			timeout: TIMEOUTS.SHORT,
		})
	}

	static async waitForSuccessMessage(page: Page): Promise<void> {
		await page.waitForSelector(".ant-message-success", {
			timeout: TIMEOUTS.SHORT,
		})
	}

	static async waitForErrorMessage(page: Page): Promise<void> {
		await page.waitForSelector(".ant-message-error", {
			timeout: TIMEOUTS.SHORT,
		})
	}

	static async clearAndType(
		page: Page,
		selector: string,
		text: string,
	): Promise<void> {
		await page.locator(selector).clear()
		await page.locator(selector).type(text)
	}

	static async selectDropdownOption(
		page: Page,
		dropdownSelector: string,
		optionText: string,
	): Promise<void> {
		await page.locator(dropdownSelector).click()
		await page.getByText(optionText).click()
	}
}

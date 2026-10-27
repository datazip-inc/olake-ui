/** The Amazon S3 destination the Go test starts (server/tests/destinations/s3): MinIO, parquet. */

import { Page } from "@playwright/test"

import { createDestination } from "./destination"
import { DestinationConnector } from "../../enums"
import { CreateDestinationPage } from "../../pages/CreateDestinationPage"
import { DestinationsPage } from "../../pages/DestinationsPage"
import { SERVICE_HOST } from "../../utils"

/** Creates the S3 destination the Go test started: s3://warehouse/e2e on MinIO. */
export const createS3Destination = (
	page: Page,
	destinationsPage: DestinationsPage,
	createDestinationPage: CreateDestinationPage,
	name: string,
) =>
	createDestination(page, destinationsPage, createDestinationPage, {
		name,
		connector: DestinationConnector.AmazonS3,
		fields: {
			s3_bucket: "warehouse",
			s3_region: "us-east-1",
			s3_path: "e2e",
			s3_endpoint: `http://${SERVICE_HOST}:9000`,
			s3_access_key: "admin",
			s3_secret_key: "password",
		},
	})

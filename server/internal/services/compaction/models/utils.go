package models

const APIBase = "/api/ams/v1/"

// given the fact, we have only one optimizer group for fusion in V0
const OptimizerGroup = "spark-cluster"

// hard-coding to S3 now, as the other options are "hadoop" & "OSS" for Compaction
// GCS & ADLS are supported, given the catalog manages the sdk (eg, Lakekeeper with GCS flavour)
const DefaultStroageType = "S3"

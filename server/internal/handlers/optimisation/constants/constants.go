package constants

// given the fact, we have only one optimizer group for fusion in V0
const OptimizerGroup = "spark-container"

// hard-coding to S3 now, as the other options are "hadoop" & "OSS" for optimisation
// GCS & ADLS are supported, given the catalog manages the sdk (eg, Lakekeeper with GCS flavour)
const DefaultStroageType = "S3"

// TableFormatList defines supported table formats for catalogs
var TableFormatList = []string{"ICEBERG"}

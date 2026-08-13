export enum AnalyticsEvent {
	// Test Connection
	TestConnectionSource = "test_connection_source",
	TestConnectionDestination = "test_connection_destination",
	// Creation flows
	CreateJobClicked = "create_job_clicked",
	CreateSourceClicked = "create_source_clicked",
	CreateDestinationClicked = "create_destination_clicked",
	// Catalog
	AddCatalogClicked = "add_catalog_clicked",
	CatalogConnectClicked = "catalog_connect_clicked",
	CatalogCreated = "catalog_created",
	CatalogUpdated = "catalog_updated",
	// Metrics
	ViewTableMetricsClicked = "view_table_metrics_clicked",
	ViewMetricsClicked = "view_metrics_clicked",
	ViewRunsAndLogsClicked = "view_runs_and_logs_clicked",
	// Job configuration
	ConfigureButtonClicked = "configure_button_clicked",
	// Job status
	ViewLogsClicked = "view_logs_clicked",
	// Maintenance
	MaintenanceModuleOpened = "maintenance_module_opened",
}

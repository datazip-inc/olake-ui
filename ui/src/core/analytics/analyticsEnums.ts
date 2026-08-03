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
	ImportCatalogFromDestinationClicked = "import_catalog_from_destination_clicked",
	CatalogConnectClicked = "catalog_connect_clicked",
	// Metrics
	ViewTableMetricsClicked = "view_table_metrics_clicked",
	ViewMetricsClicked = "view_metrics_clicked",
	HealthScoreTooltipViewed = "health_score_tooltip_viewed",
	// Job configuration
	ConfigureButtonClicked = "configure_button_clicked",
	LiteFrequencySelected = "lite_frequency_selected",
	MediumFrequencySelected = "medium_frequency_selected",
	FullFrequencySelected = "full_frequency_selected",
	FrequencyOptionClicked = "frequency_option_clicked",
	TargetFileSizeCustomized = "target_file_size_customized",
	ConfigurationSaveClicked = "configuration_save_clicked",
	ConfigurationSaveSuccessful = "configuration_save_successful",
	ConfigurationSaveFailed = "configuration_save_failed",
	// Job status
	StatusToggleOnClicked = "status_toggle_on_clicked",
	StatusToggleOffClicked = "status_toggle_off_clicked",
	StatusEnableSuccessful = "status_enable_successful",
	StatusEnableFailed = "status_enable_failed",
	StatusRetryClicked = "status_retry_clicked",
	ViewLogsClicked = "view_logs_clicked",
	// Maintenance
	MaintenanceModuleOpened = "maintenance_module_opened",
}
import { useIsFetching } from "@tanstack/react-query"
import { Button, message } from "antd"
import { useEffect, useMemo, useState } from "react"
import { useLocation, useNavigate } from "react-router-dom"

import { DataTable, PageErrorState } from "@/common/components"
import { ErrorLogsModal } from "@/common/components/modals"
import { usePaginatedSearch } from "@/common/hooks"
import { trackEvent, AnalyticsEvent } from "@/core/analytics"
import { catalogKeys } from "@/modules/maintenance/features/catalogs/constants"
import { useCatalogs } from "@/modules/maintenance/features/catalogs/hooks"

import {
	CatalogNotAvailableModal,
	ConfigureOptimizationModalBulk,
	ConfigureOptimizationModalSingle,
	DatabaseNotAvailableModal,
	TableEmptyState,
	TableFilterBar,
	TableMetricsModal,
	TablePageHeader,
} from "../components"
import { PAGE_SIZE } from "../constants"
import {
	useCancelTableRun,
	useCatalogDatabaseSelection,
	useTables,
	useToggleTableOptimizing,
} from "../hooks"
import type { Table, ToggleTableOptimizingRequest } from "../types"
import type { TableActions } from "../utils"
import { getCancelRunID, getTableColumns } from "../utils"

const EMPTY_TABLES: Table[] = []

const PAGE_STATE = {
	CATALOG_ERROR: "catalog-error",
	CATALOG_EMPTY: "catalog-empty",
	DATABASE_ERROR: "database-error",
	DATABASE_EMPTY: "database-empty",
	TABLES_ERROR: "tables-error",
	READY: "ready",
} as const

const tableSearchFn = (row: Table, term: string): boolean =>
	row.name.toLowerCase().includes(term)

const tableFilterFn = (
	row: Table,
	filter: "all" | "olake" | "external",
): boolean => {
	if (filter === "all") return true
	if (filter === "olake") return row.olakeCreated
	return !row.olakeCreated
}

const Tables: React.FC = () => {
	const navigate = useNavigate()
	const location = useLocation()

	const [openActionRow, setOpenActionRow] = useState<string | null>(null)
	const [configureModalOpen, setConfigureModalOpen] = useState(false)
	const [configureTable, setConfigureTable] = useState<Table | null>(null)
	const [metricsModalOpen, setMetricsModalOpen] = useState(false)
	const [metricsTableName, setMetricsTableName] = useState("")
	const [optimizationErrorOpen, setOptimizationErrorOpen] = useState(false)
	const [optimizationErrorLogs, setOptimizationErrorLogs] = useState<string[]>(
		[],
	)
	const [lastToggleRequest, setLastToggleRequest] =
		useState<ToggleTableOptimizingRequest | null>(null)
	const [selectedTables, setSelectedTables] = useState<string[]>([])
	const [bulkModalOpen, setBulkModalOpen] = useState(false)

	const {
		data: catalogs = [],
		isPending: isCatalogsPending,
		isError: isCatalogsError,
		refetch: refetchCatalogs,
	} = useCatalogs()
	const {
		selectedCatalog,
		selectedDatabase,
		databaseOptions,
		isDatabasesError,
		refetchDatabases,
		handleCatalogChange,
		handleDatabaseChange,
		catalogParam,
		databaseParam,
		catalogNotAvailableOpen,
		setCatalogNotAvailableOpen,
		databaseNotAvailableOpen,
		setDatabaseNotAvailableOpen,
	} = useCatalogDatabaseSelection(catalogs)

	const catalogName = selectedCatalog ?? catalogs[0]?.name ?? ""
	const isDatabasesPending =
		useIsFetching({
			queryKey: catalogKeys.databases(catalogName),
		}) > 0

	// Fetch tables only after databases load and a valid catalog/database pair is resolved.
	const canFetchTables =
		!!selectedCatalog &&
		!isDatabasesPending &&
		!!selectedDatabase &&
		databaseOptions.includes(selectedDatabase)

	const {
		data: tables = EMPTY_TABLES,
		isFetching: isTablesFetching,
		isError: isTablesError,
		refetch: refetchTables,
	} = useTables(selectedCatalog ?? "", selectedDatabase ?? "", canFetchTables)

	const {
		searchTerm,
		setSearchTerm,
		activeFilter,
		setActiveFilter,
		currentPage,
		setCurrentPage,
		filteredRows,
		paginatedRows,
		totalPages,
	} = usePaginatedSearch<Table, "all" | "olake" | "external">({
		rows: tables,
		pageSize: PAGE_SIZE,
		searchFn: tableSearchFn,
		filterFn: tableFilterFn,
		initialFilter: "all",
	})

	useEffect(() => {
		if (isTablesFetching && !bulkModalOpen) {
			setSelectedTables([])
		}
	}, [isTablesFetching, bulkModalOpen])

	const handleBulkConfigure = () => {
		if (selectedTables.length < 2) {
			message.destroy()
			message.info("Please select more than 1 table to bulk configure.")
			return
		}
		trackEvent(AnalyticsEvent.ConfigureButtonClicked)
		setBulkModalOpen(true)
	}

	const {
		mutate: toggleTableOptimizing,
		isPending: isToggleTableOptimizingPending,
		variables: toggleTableOptimizingVariables,
	} = useToggleTableOptimizing()
	const {
		mutate: cancelTableRun,
		isPending: isCancelTableRunPending,
		variables: cancelTableRunVariables,
	} = useCancelTableRun()

	const loading = isCatalogsPending || isDatabasesPending || isTablesFetching

	// Get Page State according to the error
	const pageState = (() => {
		if (isCatalogsError) return PAGE_STATE.CATALOG_ERROR
		if (!isCatalogsPending && catalogs.length === 0)
			return PAGE_STATE.CATALOG_EMPTY
		if (isDatabasesError) return PAGE_STATE.DATABASE_ERROR
		if (
			!isDatabasesPending &&
			!!selectedCatalog &&
			databaseOptions.length === 0
		)
			return PAGE_STATE.DATABASE_EMPTY
		if (isTablesError) return PAGE_STATE.TABLES_ERROR
		return PAGE_STATE.READY
	})()

	const getTableRunsPath = (tableName: string) =>
		`/maintenance/tables/${encodeURIComponent(selectedCatalog ?? "")}/${encodeURIComponent(selectedDatabase ?? "")}/${encodeURIComponent(tableName)}/runs`

	const columns = useMemo(() => {
		const isTogglePendingFor = (tableName: string) =>
			isToggleTableOptimizingPending &&
			toggleTableOptimizingVariables?.tableName === tableName

		const isCancelPendingFor = (tableName: string) =>
			isCancelTableRunPending &&
			cancelTableRunVariables?.tableName === tableName

		const actions: TableActions = {
			onViewLogs: row => navigate(getTableRunsPath(row.name)),
			onCancelRun: row => {
				const runId = getCancelRunID(row)
				if (!runId) return
				cancelTableRun({
					catalog: selectedCatalog ?? "",
					database: selectedDatabase ?? "",
					tableName: row.name,
					runId,
				})
			},
			onToggleOptimizingStatus: (row, enabled) => {
				const request: ToggleTableOptimizingRequest = {
					catalog: selectedCatalog ?? "",
					database: selectedDatabase ?? "",
					tableName: row.name,
					enabled,
				}
				toggleTableOptimizing(request, {
					onSuccess: result => {
						if (!result.success) {
							setLastToggleRequest(request)
							setOptimizationErrorLogs(result.logs ?? [])
							setOptimizationErrorOpen(true)
						}
					},
				})
			},
			onViewMetrics: row => {
				trackEvent(AnalyticsEvent.ViewTableMetricsClicked)
				setMetricsTableName(row.name)
				setMetricsModalOpen(true)
			},
			onConfigure: row => {
				trackEvent(AnalyticsEvent.ConfigureButtonClicked)
				setConfigureTable(row)
				setConfigureModalOpen(true)
			},
		}

		return getTableColumns({
			openActionRow,
			setOpenActionRow,
			isTogglePendingFor,
			isCancelPendingFor,
			actions,
		})
	}, [
		openActionRow,
		selectedCatalog,
		selectedDatabase,
		isToggleTableOptimizingPending,
		toggleTableOptimizingVariables,
		isCancelTableRunPending,
		cancelTableRunVariables,
		navigate,
		cancelTableRun,
		toggleTableOptimizing,
	])

	const handleCatalogUnavailableClose = () => {
		setCatalogNotAvailableOpen(false)
		window.location.assign(location.pathname)
	}

	const handleDatabaseUnavailableClose = () => {
		setDatabaseNotAvailableOpen(false)
		const validCatalog = catalogParam ?? selectedCatalog ?? ""
		window.location.assign(
			`${location.pathname}?catalog=${encodeURIComponent(validCatalog)}`,
		)
	}

	// Retries the appropriate API based on which state is currently failing or empty.
	const handleRetry = () => {
		if (pageState === PAGE_STATE.CATALOG_ERROR) void refetchCatalogs()
		else if (
			pageState === PAGE_STATE.DATABASE_ERROR ||
			pageState === PAGE_STATE.DATABASE_EMPTY
		)
			void refetchDatabases()
		else void refetchTables()
	}

	return (
		<>
			<div className="min-h-full bg-white px-6 pt-6">
				{pageState !== PAGE_STATE.CATALOG_ERROR &&
					pageState !== PAGE_STATE.CATALOG_EMPTY && (
						<TablePageHeader
							catalogs={catalogs}
							isCatalogsPending={isCatalogsPending}
							databaseOptions={databaseOptions}
							selectedCatalog={selectedCatalog}
							selectedDatabase={selectedDatabase}
							loading={loading}
							isRefreshDisabled={
								pageState === PAGE_STATE.DATABASE_EMPTY ||
								pageState === PAGE_STATE.DATABASE_ERROR
							}
							onCatalogChange={handleCatalogChange}
							onDatabaseChange={handleDatabaseChange}
							onRefresh={refetchTables}
						/>
					)}
				<div
					className={
						pageState !== PAGE_STATE.CATALOG_ERROR &&
						pageState !== PAGE_STATE.CATALOG_EMPTY
							? "mt-8 w-full"
							: "w-full"
					}
				>
					{pageState === PAGE_STATE.CATALOG_ERROR ? (
						<PageErrorState
							title="Failed to load catalogs"
							description="Please check your connection and try again."
							onRetry={handleRetry}
						/>
					) : pageState === PAGE_STATE.CATALOG_EMPTY ? (
						<TableEmptyState />
					) : pageState === PAGE_STATE.DATABASE_ERROR ? (
						<PageErrorState
							title="Failed to load databases"
							description="Please check your connection and try again."
							onRetry={handleRetry}
						/>
					) : pageState === PAGE_STATE.DATABASE_EMPTY ? (
						<PageErrorState
							title="No Database Found"
							description="There are no databases in the selected catalog."
							onRetry={handleRetry}
						/>
					) : pageState === PAGE_STATE.TABLES_ERROR ? (
						<PageErrorState
							title="Failed to load tables"
							description="Please check your connection and try again."
							onRetry={handleRetry}
						/>
					) : (
						<div className="flex flex-col gap-6">
							<div className="grid w-full grid-cols-[1fr_auto] items-center gap-4">
								<div className="min-w-0 overflow-x-auto">
									<TableFilterBar
										searchTerm={searchTerm}
										onSearchChange={setSearchTerm}
										activeFilter={activeFilter}
										onFilterChange={setActiveFilter}
									/>
								</div>
								<Button
									type="primary"
									size="middle"
									className="shrink-0"
									onClick={handleBulkConfigure}
								>
									Bulk Configure
								</Button>
							</div>
							<DataTable
								columns={columns}
								rows={paginatedRows}
								rowKey={row => row.name}
								checkboxSelection
								selectedRowKeys={selectedTables}
								allSelectableRows={filteredRows}
								onSelectionChange={setSelectedTables}
								loading={loading}
								emptyStateConfig={{
									title: tables.length === 0 ? "No Tables" : "No Tables Found.",
									subtitle:
										tables.length > 0
											? "Try a different search or filter."
											: "There are no tables in the selected database.",
									onRefetch: handleRetry,
								}}
								pagination={{
									currentPage,
									totalPages,
									onPageChange: setCurrentPage,
								}}
							/>
						</div>
					)}
				</div>
			</div>

			<ConfigureOptimizationModalSingle
				open={configureModalOpen}
				onClose={() => {
					setConfigureModalOpen(false)
					setConfigureTable(null)
				}}
				catalog={selectedCatalog ?? ""}
				database={selectedDatabase ?? ""}
				tableName={configureTable?.name ?? ""}
				tableSize={configureTable?.totalSize ?? ""}
			/>
			<ConfigureOptimizationModalBulk
				open={bulkModalOpen}
				onClose={() => {
					setBulkModalOpen(false)
					setSelectedTables([])
				}}
				catalog={selectedCatalog ?? ""}
				database={selectedDatabase ?? ""}
				tables={selectedTables}
				onRemoveTable={table =>
					setSelectedTables(prev => prev.filter(tab => tab !== table))
				}
			/>
			<TableMetricsModal
				open={metricsModalOpen}
				onClose={() => setMetricsModalOpen(false)}
				catalog={selectedCatalog ?? ""}
				database={selectedDatabase ?? ""}
				tableName={metricsTableName}
			/>
			<CatalogNotAvailableModal
				open={catalogNotAvailableOpen}
				onClose={handleCatalogUnavailableClose}
				catalogName={catalogParam ?? ""}
			/>
			<DatabaseNotAvailableModal
				open={databaseNotAvailableOpen}
				onClose={handleDatabaseUnavailableClose}
				databaseName={databaseParam ?? ""}
			/>

			<ErrorLogsModal
				open={optimizationErrorOpen}
				onClose={() => setOptimizationErrorOpen(false)}
				title="Failed to update optimization configuration"
				error={optimizationErrorLogs.join("\n")}
				onAction={() => {
					setOptimizationErrorOpen(false)
					if (lastToggleRequest) {
						toggleTableOptimizing(lastToggleRequest, {
							onSuccess: result => {
								if (!result.success) {
									setOptimizationErrorLogs(result.logs ?? [])
									setOptimizationErrorOpen(true)
								}
							},
						})
					}
				}}
				actionButtonText="Retry"
			/>
		</>
	)
}

export default Tables

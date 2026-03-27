import { useIsFetching } from "@tanstack/react-query"
import { useMemo, useState } from "react"
import { useLocation, useNavigate } from "react-router-dom"

import { DataTable, PageErrorState } from "@/common/components"
import { usePaginatedSearch } from "@/common/hooks"
import { catalogKeys } from "@/modules/maintenance/features/catalogs/constants"
import { useCatalogs } from "@/modules/maintenance/features/catalogs/hooks"

import {
	CatalogNotAvailableModal,
	ConfigureOptimizationModal,
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
import type { Table } from "../types"
import type { TableActions } from "../utils"
import { getCancelRunID, getTableColumns } from "../utils"

const tableSearchFn = (row: Table, term: string): boolean =>
	row.name.toLowerCase().includes(term)

const tableFilterFn = (
	row: Table,
	filter: "all" | "olake" | "external",
): boolean => {
	if (filter === "all") return true
	if (filter === "olake") return row.byOLake
	return !row.byOLake
}

const Tables: React.FC = () => {
	const navigate = useNavigate()
	const location = useLocation()

	const [openActionRow, setOpenActionRow] = useState<string | null>(null)
	const [configureModalOpen, setConfigureModalOpen] = useState(false)
	const [configureTable, setConfigureTable] = useState<Table | null>(null)
	const [metricsModalOpen, setMetricsModalOpen] = useState(false)
	const [metricsTableName, setMetricsTableName] = useState("")

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
		data: tables = [],
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
		paginatedRows,
		totalPages,
	} = usePaginatedSearch<Table, "all" | "olake" | "external">({
		rows: tables,
		pageSize: PAGE_SIZE,
		searchFn: tableSearchFn,
		filterFn: tableFilterFn,
		initialFilter: "all",
	})

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
	const showPageError = isCatalogsError || isTablesError
	const showCatalogEmptyState =
		!isCatalogsPending && !isCatalogsError && catalogs.length === 0
	const showDatabaseEmptyState =
		!isCatalogsPending &&
		!isCatalogsError &&
		!isDatabasesPending &&
		catalogs.length > 0 &&
		!!selectedCatalog &&
		databaseOptions.length === 0

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
			onToggleOptimizing: (row, enabled) =>
				toggleTableOptimizing({
					catalog: selectedCatalog ?? "",
					database: selectedDatabase ?? "",
					tableName: row.name,
					enabled,
				}),
			onViewMetrics: row => {
				setMetricsTableName(row.name)
				setMetricsModalOpen(true)
			},
			onConfigure: row => {
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

	const handleRetry = () => {
		if (isCatalogsError) {
			void refetchCatalogs()
		} else {
			void refetchTables()
		}
	}

	return (
		<>
			<div className="min-h-full bg-white px-6 pt-6">
				{showPageError ? (
					<PageErrorState
						title={
							isCatalogsError
								? "Failed to load catalogs"
								: "Failed to load tables"
						}
						description="Please check your connection and try again."
						onRetry={handleRetry}
					/>
				) : showCatalogEmptyState ? (
					<TableEmptyState />
				) : (
					<>
						<TablePageHeader
							catalogs={catalogs}
							isCatalogsPending={isCatalogsPending}
							databaseOptions={databaseOptions}
							selectedCatalog={selectedCatalog}
							selectedDatabase={selectedDatabase}
							loading={loading}
							isRefreshDisabled={showDatabaseEmptyState}
							onCatalogChange={handleCatalogChange}
							onDatabaseChange={handleDatabaseChange}
							onRefresh={refetchTables}
						/>
						<div className="mt-8 w-full">
							{showDatabaseEmptyState ? (
								<PageErrorState
									title="No Database Found"
									description="There are no databases in the selected catalog."
									onRetry={() => void refetchCatalogs()}
								/>
							) : (
								<div className="flex flex-col gap-6">
									<TableFilterBar
										searchTerm={searchTerm}
										onSearchChange={setSearchTerm}
										activeFilter={activeFilter}
										onFilterChange={setActiveFilter}
									/>
									<DataTable
										columns={columns}
										rows={paginatedRows}
										rowKey={row => row.id}
										loading={loading}
										emptyStateConfig={{
											title:
												tables.length === 0 ? "No Tables" : "No Tables Found.",
											subtitle:
												tables.length > 0
													? "Try a different search or filter."
													: "There are no tables in the selected catalog.",
											onRefetch: () => navigate("/maintenance/catalogs"),
											refetchLabel: "Add Catalog",
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
					</>
				)}
			</div>

			<ConfigureOptimizationModal
				open={configureModalOpen}
				onClose={() => setConfigureModalOpen(false)}
				catalog={selectedCatalog ?? ""}
				database={selectedDatabase ?? ""}
				tableName={configureTable?.name ?? ""}
				tableSize={configureTable?.totalSize ?? ""}
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
		</>
	)
}

export default Tables

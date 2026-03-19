import { Button } from "antd"
import { useMemo, useState } from "react"
import { useLocation, useNavigate } from "react-router-dom"

import { DataTable, PageErrorState } from "@/common/components"
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
import {
	useCancelTableRun,
	useCatalogDatabaseSelection,
	useFilteredTables,
	useTables,
	useToggleTableOptimizing,
} from "../hooks"
import type { Table } from "../types"
import type { TableActions } from "../utils"
import { getTableColumns } from "../utils"

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
		setSelectedDatabase,
		handleCatalogChange,
		catalogParam,
		databaseParam,
		catalogNotAvailableOpen,
		setCatalogNotAvailableOpen,
		databaseNotAvailableOpen,
		setDatabaseNotAvailableOpen,
	} = useCatalogDatabaseSelection(catalogs, isCatalogsPending)

	const {
		data: tables = [],
		isFetching: isTablesFetching,
		isError: isTablesError,
		refetch: refetchTables,
	} = useTables(selectedCatalog ?? "", selectedDatabase ?? "")

	const {
		searchTerm,
		setSearchTerm,
		activeFilter,
		setActiveFilter,
		currentPage,
		setCurrentPage,
		paginatedRows,
		totalPages,
	} = useFilteredTables(tables)

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

	const loading = isCatalogsPending || isTablesFetching
	const showPageError = isCatalogsError || isTablesError
	const showCatalogEmptyState =
		!isCatalogsPending && !isCatalogsError && catalogs.length === 0

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
			onCancelRun: row =>
				cancelTableRun({
					catalog: selectedCatalog ?? "",
					database: selectedDatabase ?? "",
					tableName: row.name,
				}),
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

	const emptySearchState = (
		<div className="flex h-56 items-center justify-center">
			<div className="text-center">
				<p className="text-xl font-medium leading-7 text-olake-heading-strong">
					No Tables Found.
				</p>
				<p className="mt-1 text-sm leading-[22px] text-olake-body">
					Try a different search or filter.
				</p>
				<Button
					type="primary"
					className="mt-4"
					onClick={() => navigate("/maintenance/catalogs")}
				>
					Add Catalog
				</Button>
			</div>
		</div>
	)

	const handleUnavailableSelectionClose = () => {
		setCatalogNotAvailableOpen(false)
		setDatabaseNotAvailableOpen(false)
		window.location.assign(location.pathname)
	}

	const handleRetry = () => {
		void refetchCatalogs()
		void refetchTables()
	}

	return (
		<>
			<div className="min-h-full bg-white px-6 pt-6">
				{showPageError ? (
					<PageErrorState
						title="Failed to load tables"
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
							selectedCatalog={selectedCatalog}
							selectedDatabase={selectedDatabase}
							loading={loading}
							onCatalogChange={handleCatalogChange}
							onDatabaseChange={setSelectedDatabase}
							onRefresh={refetchTables}
						/>
						<div className="mt-8 w-full">
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
									emptyState={emptySearchState}
									pagination={{
										currentPage,
										totalPages,
										onPageChange: setCurrentPage,
									}}
								/>
							</div>
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
				onClose={handleUnavailableSelectionClose}
				catalogName={catalogParam}
			/>
			<DatabaseNotAvailableModal
				open={databaseNotAvailableOpen}
				onClose={handleUnavailableSelectionClose}
				databaseName={databaseParam}
			/>
		</>
	)
}

export default Tables

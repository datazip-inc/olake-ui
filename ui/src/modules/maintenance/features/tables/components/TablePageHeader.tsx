import { ArrowsClockwiseIcon } from "@phosphor-icons/react"
import { useIsFetching } from "@tanstack/react-query"
import { Button, Select } from "antd"

import { catalogKeys } from "@/modules/maintenance/features/catalogs/constants"
import type { Catalog } from "@/modules/maintenance/features/catalogs/types"

type Props = {
	catalogs: Catalog[]
	isCatalogsPending: boolean
	databaseOptions: string[]
	selectedCatalog: string | undefined
	selectedDatabase: string | undefined
	loading: boolean
	onCatalogChange: (catalog: string) => void
	onDatabaseChange: (database: string) => void
	onRefresh: () => void
	isRefreshDisabled?: boolean
}

const TablePageHeader: React.FC<Props> = ({
	catalogs,
	isCatalogsPending,
	databaseOptions,
	selectedCatalog,
	selectedDatabase,
	loading,
	onCatalogChange,
	onDatabaseChange,
	onRefresh,
	isRefreshDisabled,
}) => {
	const catalogNameForDatabasesQuery =
		selectedCatalog ?? catalogs[0]?.name ?? ""

	const isDatabasesPending =
		useIsFetching({
			queryKey: catalogKeys.databases(catalogNameForDatabasesQuery),
		}) > 0

	const catalogOptions = catalogs.map(c => ({ label: c.name, value: c.name }))
	const databaseSelectOptions = databaseOptions.map(db => ({
		label: db,
		value: db,
	}))

	return (
		<div className="w-full">
			<h1 className="text-xl font-medium leading-7 text-olake-heading-strong">
				Tables
			</h1>
			<p className="mt-1 text-sm leading-[22px] text-olake-heading-strong">
				Select Catalog &amp; Database to view tables &amp; run optimization
			</p>
			<div className="mt-4 flex items-center gap-4">
				<Select
					value={selectedCatalog}
					onChange={onCatalogChange}
					className="w-72"
					options={catalogOptions}
					loading={isCatalogsPending}
					placeholder="Select Catalog"
				/>
				<Select
					placeholder="Select Database"
					value={selectedDatabase}
					onChange={onDatabaseChange}
					className="w-72"
					options={databaseSelectOptions}
					loading={isDatabasesPending}
					disabled={!selectedCatalog || databaseSelectOptions.length === 0}
				/>
				<Button
					type="primary"
					icon={<ArrowsClockwiseIcon size={16} />}
					loading={loading}
					onClick={onRefresh}
					disabled={isRefreshDisabled}
				>
					Refresh Tables
				</Button>
			</div>
		</div>
	)
}

export default TablePageHeader

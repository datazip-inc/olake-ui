import { ArrowsClockwiseIcon } from "@phosphor-icons/react"
import { Button, Select } from "antd"

import type { Catalog } from "@/modules/maintenance/features/catalogs/types"

type Props = {
	catalogs: Catalog[]
	isCatalogsPending: boolean
	selectedCatalog: string | undefined
	selectedDatabase: string | undefined
	loading: boolean
	onCatalogChange: (catalog: string) => void
	onDatabaseChange: (database: string) => void
	onRefresh: () => void
}

const TablePageHeader: React.FC<Props> = ({
	catalogs,
	isCatalogsPending,
	selectedCatalog,
	selectedDatabase,
	loading,
	onCatalogChange,
	onDatabaseChange,
	onRefresh,
}) => {
	const catalogOptions = catalogs.map(c => ({ label: c.name, value: c.name }))

	const selectedCatalogRow = catalogs.find(c => c.name === selectedCatalog)
	const databaseOptions = (selectedCatalogRow?.databases ?? []).map(db => ({
		label: db,
		value: db,
	}))

	return (
		<div className="w-full">
			<h1 className="text-xl font-medium leading-7 text-olake-heading-strong">
				Tables
			</h1>
			<p className="mt-1 text-sm leading-[22px] text-olake-heading-strong">
				Select Catalog &amp; Database to view tables &amp; run maintenance
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
					options={databaseOptions}
					disabled={!selectedCatalog || databaseOptions.length === 0}
				/>
				<Button
					type="primary"
					icon={<ArrowsClockwiseIcon size={16} />}
					loading={loading}
					onClick={onRefresh}
				>
					Refresh Tables
				</Button>
			</div>
		</div>
	)
}

export default TablePageHeader

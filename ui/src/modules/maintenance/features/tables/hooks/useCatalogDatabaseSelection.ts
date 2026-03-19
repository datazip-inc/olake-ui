import { useEffect, useState } from "react"
import { useSearchParams } from "react-router-dom"

import type { Catalog } from "@/modules/maintenance/features/catalogs/types"

type UseCatalogDatabaseSelectionReturn = {
	selectedCatalog: string | undefined
	selectedDatabase: string | undefined
	setSelectedDatabase: (db: string | undefined) => void
	handleCatalogChange: (catalogName: string) => void
	catalogParam: string | undefined
	databaseParam: string | undefined
	catalogNotAvailableOpen: boolean
	setCatalogNotAvailableOpen: (open: boolean) => void
	databaseNotAvailableOpen: boolean
	setDatabaseNotAvailableOpen: (open: boolean) => void
}

/** Syncs catalog/database selection with URL query params and validates them against loaded catalog data. */
export function useCatalogDatabaseSelection(
	catalogs: Catalog[],
	isCatalogsPending: boolean,
): UseCatalogDatabaseSelectionReturn {
	const [searchParams, setSearchParams] = useSearchParams()
	const catalogParam = searchParams.get("catalog") ?? undefined
	const databaseParam = searchParams.get("database") ?? undefined

	const [selectedCatalog, setSelectedCatalog] = useState<string | undefined>(
		catalogParam,
	)
	const [selectedDatabase, setSelectedDatabase] = useState<string | undefined>(
		databaseParam,
	)
	const [catalogNotAvailableOpen, setCatalogNotAvailableOpen] = useState(false)
	const [databaseNotAvailableOpen, setDatabaseNotAvailableOpen] =
		useState(false)

	// URL params → state: validate against loaded catalogs, auto-select when no params
	useEffect(() => {
		if (isCatalogsPending || catalogs.length === 0) return

		if (!catalogParam) {
			setSelectedCatalog(catalogs[0].name)
			setSelectedDatabase(catalogs[0].databases[0])
			return
		}

		const foundCatalog = catalogs.find(c => c.name === catalogParam)
		if (!foundCatalog) {
			setCatalogNotAvailableOpen(true)
			return
		}

		setSelectedCatalog(foundCatalog.name)

		if (!databaseParam) {
			setSelectedDatabase(foundCatalog.databases[0])
			return
		}

		const foundDb = foundCatalog.databases.find(db => db === databaseParam)
		if (!foundDb) {
			setDatabaseNotAvailableOpen(true)
			return
		}

		setSelectedDatabase(foundDb)
	}, [catalogs, isCatalogsPending, catalogParam, databaseParam])

	// state → URL params (skip until a catalog is selected)
	useEffect(() => {
		if (!selectedCatalog) return
		const params: Record<string, string> = { catalog: selectedCatalog }
		if (selectedDatabase) params.database = selectedDatabase
		setSearchParams(params, { replace: true })
	}, [selectedCatalog, selectedDatabase, setSearchParams])

	const handleCatalogChange = (catalogName: string) => {
		setSelectedCatalog(catalogName)
		const catalog = catalogs.find(c => c.name === catalogName)
		setSelectedDatabase(catalog?.databases[0])
	}

	return {
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
	}
}

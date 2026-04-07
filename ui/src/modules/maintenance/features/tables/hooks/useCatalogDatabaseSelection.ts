import { useIsFetching } from "@tanstack/react-query"
import { useEffect, useState } from "react"
import { useSearchParams } from "react-router-dom"

import { catalogKeys } from "@/modules/maintenance/features/catalogs/constants"
import { useCatalogDatabases } from "@/modules/maintenance/features/catalogs/hooks"
import type { Catalog } from "@/modules/maintenance/features/catalogs/types"

type UseCatalogDatabaseSelectionReturn = {
	selectedCatalog: string | undefined
	selectedDatabase: string | undefined
	databaseOptions: string[]
	setSelectedDatabase: (db: string | undefined) => void
	handleCatalogChange: (catalogName: string) => void
	handleDatabaseChange: (database: string) => void
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

	const isCatalogsPending = useIsFetching({ queryKey: catalogKeys.list() }) > 0

	const catalogName = selectedCatalog ?? catalogParam ?? catalogs[0]?.name ?? ""

	const { data: databaseOptions = [], isPending: isDatabasesPending } =
		useCatalogDatabases(catalogName)

	// URL params → state: validate against loaded catalogs, auto-select when no params.
	// When auto-selecting (no catalogParam), also writes catalog to URL
	useEffect(() => {
		if (isCatalogsPending || catalogs.length === 0) return

		if (!catalogParam) {
			setSelectedCatalog(catalogs[0].name)
			setSearchParams({ catalog: catalogs[0].name }, { replace: true })
			return
		}

		const foundCatalog = catalogs.find(c => c.name === catalogParam)
		if (!foundCatalog) {
			setSelectedCatalog(undefined)
			setCatalogNotAvailableOpen(true)
			return
		}

		setSelectedCatalog(foundCatalog.name)
	}, [catalogs, isCatalogsPending, catalogParam, setSearchParams])

	// Once databases for the selected catalog are loaded, validate / default the database.
	// When auto-selecting (no databaseParam), also writes database to URL
	useEffect(() => {
		if (!selectedCatalog || isDatabasesPending) return

		if (!databaseParam) {
			const firstDb = databaseOptions[0]
			setSelectedDatabase(firstDb)
			if (firstDb) {
				setSearchParams(
					{ catalog: selectedCatalog, database: firstDb },
					{ replace: true },
				)
			}
			return
		}

		const foundDb = databaseOptions.find(db => db === databaseParam)
		if (!foundDb) {
			setDatabaseNotAvailableOpen(true)
			return
		}

		setSelectedDatabase(foundDb)
	}, [
		selectedCatalog,
		isDatabasesPending,
		databaseParam,
		databaseOptions,
		setSearchParams,
	])

	const handleCatalogChange = (catalogName: string) => {
		setSelectedCatalog(catalogName)
		setSelectedDatabase(undefined)
		setSearchParams({ catalog: catalogName }, { replace: true })
	}

	const handleDatabaseChange = (database: string) => {
		setSelectedDatabase(database)
		setSearchParams(
			{ catalog: selectedCatalog ?? "", database },
			{ replace: true },
		)
	}

	return {
		selectedCatalog,
		selectedDatabase,
		databaseOptions,
		setSelectedDatabase,
		handleCatalogChange,
		handleDatabaseChange,
		catalogParam,
		databaseParam,
		catalogNotAvailableOpen,
		setCatalogNotAvailableOpen,
		databaseNotAvailableOpen,
		setDatabaseNotAvailableOpen,
	}
}

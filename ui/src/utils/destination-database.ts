import { StreamsDataStructure } from "../types"
import { FORMAT_OPTIONS, NAMESPACE_PLACEHOLDER } from "./constants"

type FormatType = (typeof FORMAT_OPTIONS)[keyof typeof FORMAT_OPTIONS]

/**
 * Formats a destination path by combining database and table names
 * If database contains ":", joins the parts with "_" instead
 */
export const formatDestinationPath = (
	destinationDatabase: string | null | undefined,
	destinationTable: string | null | undefined,
): string | null => {
	if (!destinationDatabase || !destinationTable) {
		return null
	}

	// Split destination_database with : and join with _
	// Example: "a:b" becomes "a_b"
	const formattedDatabase = destinationDatabase.includes(":")
		? destinationDatabase.split(":").join("_")
		: destinationDatabase

	return `${formattedDatabase}/${destinationTable}`
}

/**
 * Extracts the database prefix from a destination database string
 * For "database:suffix" returns "database", otherwise returns the full string
 */
export const extractDatabasePrefix = (
	destinationDatabase: string | null,
): string => {
	if (!destinationDatabase) return ""
	return destinationDatabase.includes(":")
		? destinationDatabase.split(":")[0]
		: destinationDatabase
}

/**
 * Generates database names by combining the base name with namespaces from streams
 */
export const generateDatabaseNames = (
	databaseName: string,
	allStreams: StreamsDataStructure | null,
	originalStreams: StreamsDataStructure | null,
): string[] => {
	if (!allStreams?.selected_streams || !databaseName) return []

	// Get unique namespaces from selected streams
	const selectedNamespaces = Object.keys(allStreams.selected_streams).filter(
		namespace => allStreams.selected_streams![namespace]?.length > 0,
	)

	if (selectedNamespaces.length === 0) return []

	// Map each selected namespace to its corresponding original namespace from destination_database
	return selectedNamespaces.map(namespace => {
		// Find a stream in this namespace from original streams
		const originalStream = originalStreams?.streams.find(
			s =>
				s.stream.namespace === namespace &&
				s.stream.destination_database?.includes(":"),
		)

		// Get namespace from original destination_database if exists, otherwise use current namespace
		const finalNamespace = extractNamespaceFromDestination(
			originalStream?.stream.destination_database,
			namespace,
		)

		return `${databaseName}_${finalNamespace}`
	})
}

/**
 * Extracts namespace from destination database string or returns fallback namespace
 * For "database:namespace" returns "namespace", otherwise returns fallback
 */
export const extractNamespaceFromDestination = (
	destinationDatabase: string | undefined | null,
	fallbackNamespace: string,
): string => {
	if (!destinationDatabase) return fallbackNamespace
	return destinationDatabase.includes(":")
		? destinationDatabase.split(":")[1]
		: fallbackNamespace
}

/**
 * Determines the default format based on the original database string
 * Returns DYNAMIC if the database contains namespace placeholder, otherwise CUSTOM
 */
export const determineDefaultFormat = (
	originalDatabase: string,
): FormatType => {
	return originalDatabase.includes(NAMESPACE_PLACEHOLDER)
		? FORMAT_OPTIONS.DYNAMIC
		: FORMAT_OPTIONS.CUSTOM
}

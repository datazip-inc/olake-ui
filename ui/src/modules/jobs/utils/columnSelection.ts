import type { SelectedStream } from "../../../types"

// Returns true if the selected stream supports explicit column selection via the `selected_columns` field.
export function isColumnSelectionSupported(
	selectedStream: SelectedStream,
): boolean {
	return selectedStream.selected_columns !== undefined
}

// Returns true if the specified column is enabled for the selected stream.
// For legacy drivers, all columns are considered enabled by default.
export function isColumnEnabled(
	columnName: string,
	selectedStream: SelectedStream,
): boolean {
	if (!isColumnSelectionSupported(selectedStream)) return true
	return selectedStream.selected_columns!.columns.includes(columnName)
}

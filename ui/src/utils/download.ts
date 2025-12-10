/**
 * Downloads a Blob as a file, extracting filename from Content-Disposition header
 * @param blob - The Blob object to download
 * @param contentDisposition - The Content-Disposition header value
 * @param fallbackFilename - Fallback filename if header extraction fails
 */
export const downloadBlob = (
	blob: Blob,
	contentDisposition: string | undefined,
	fallbackFilename: string,
): void => {
	// Extract filename from Content-Disposition header
	const filenameMatch = contentDisposition?.match(/filename="(.+)"/)
	const filename = filenameMatch?.[1] || fallbackFilename

	// Create and trigger download
	const url = window.URL.createObjectURL(blob)
	const link = document.createElement("a")
	link.href = url
	link.download = filename
	link.style.display = "none"
	document.body.appendChild(link)
	link.click()

	// Cleanup after download initiates
	document.body.removeChild(link)
	window.URL.revokeObjectURL(url)
}

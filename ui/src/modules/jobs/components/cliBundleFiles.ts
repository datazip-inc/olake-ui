export const REQUIRED_BUNDLE_FILES = [
	"source.json",
	"destination.json",
	"streams.json",
] as const

const OPTIONAL_BUNDLE_FILES = ["olake-ui.json", "state.json"] as const

const RECOGNIZED_BUNDLE_FILES = new Set<string>([
	...REQUIRED_BUNDLE_FILES,
	...OPTIONAL_BUNDLE_FILES,
])

const TAR_BLOCK_SIZE = 512

export type StagedCLIBundleKind = "archive" | "folder" | "json-files"

export interface StagedCLIBundle {
	id: string
	key: string
	name: string
	archiveRoot: string
	kind: StagedCLIBundleKind
	sourceLabel: string
	displayName: string
	archiveFile?: File
	entries: Array<{
		baseName: string
		file: File
	}>
	fileNames: string[]
	missingRequiredFiles: string[]
	duplicateFiles: string[]
	ignoredFileCount: number
}

export const isBundleArchive = (fileName: string) =>
	/(\.zip|\.tar\.gz|\.tgz)$/i.test(fileName)

const inferArchiveBundleName = (fileName: string) =>
	fileName.replace(/(\.tar\.gz|\.tgz|\.zip)$/i, "")

const sanitizeArchiveRoot = (name: string) => {
	const cleaned = name
		.trim()
		.toLowerCase()
		.replace(/[^a-z0-9._-]+/g, "-")
		.replace(/^-+|-+$/g, "")

	return cleaned || "cli-bundle"
}

const getRelativePath = (file: File) => {
	const relativePath = (
		file as File & {
			webkitRelativePath?: string
		}
	).webkitRelativePath

	return (relativePath || "").replace(/\\/g, "/")
}

const getGroupInfo = (file: File, fallbackName: string) => {
	const relativePath = getRelativePath(file)
	if (!relativePath) {
		return {
			groupKey: fallbackName,
			groupName: fallbackName,
		}
	}

	const parts = relativePath.split("/").filter(Boolean)
	if (parts.length <= 1) {
		return {
			groupKey: fallbackName,
			groupName: fallbackName,
		}
	}

	const groupPath = parts.slice(0, -1).join("/")
	const groupName = parts[parts.length - 2] || fallbackName

	return {
		groupKey: groupPath || fallbackName,
		groupName,
	}
}

const buildStructuredBundle = (
	groupKey: string,
	groupName: string,
	files: File[],
	kind: StagedCLIBundleKind,
): StagedCLIBundle => {
	const sourceLabel = kind === "folder" ? "Folder" : "JSON files"
	const recognizedEntries: Array<{ baseName: string; file: File }> = []
	const fileNames = new Set<string>()
	const duplicateFiles = new Set<string>()
	let ignoredFileCount = 0

	for (const file of files) {
		const baseName = file.name
		if (!RECOGNIZED_BUNDLE_FILES.has(baseName)) {
			ignoredFileCount += 1
			continue
		}

		if (fileNames.has(baseName)) {
			duplicateFiles.add(baseName)
			continue
		}

		fileNames.add(baseName)
		recognizedEntries.push({ baseName, file })
	}

	const missingRequiredFiles = REQUIRED_BUNDLE_FILES.filter(
		fileName => !fileNames.has(fileName),
	)

	return {
		id: crypto.randomUUID(),
		key: `${kind}:${groupKey}`,
		name: groupName,
		archiveRoot: sanitizeArchiveRoot(groupName),
		kind,
		sourceLabel,
		displayName: groupName,
		entries: recognizedEntries,
		fileNames: Array.from(fileNames).sort(),
		missingRequiredFiles,
		duplicateFiles: Array.from(duplicateFiles).sort(),
		ignoredFileCount,
	}
}

export const stageArchiveFiles = (files: File[]) =>
	files
		.filter(file => isBundleArchive(file.name))
		.map(file => {
			const bundleName = inferArchiveBundleName(file.name)
			return {
				id: crypto.randomUUID(),
				key: `archive:${file.name}`,
				name: bundleName,
				archiveRoot: sanitizeArchiveRoot(bundleName),
				kind: "archive" as const,
				sourceLabel: "Bundle archive",
				displayName: file.name,
				archiveFile: file,
				entries: [],
				fileNames: [file.name],
				missingRequiredFiles: [],
				duplicateFiles: [],
				ignoredFileCount: 0,
			}
		})

export const stageStructuredFiles = (
	files: File[],
	kind: Extract<StagedCLIBundleKind, "folder" | "json-files">,
) => {
	if (files.length === 0) {
		return []
	}

	const fallbackName = kind === "folder" ? "selected-folder" : "selected-files"
	const groupedFiles = new Map<string, { name: string; files: File[] }>()

	for (const file of files) {
		const { groupKey, groupName } = getGroupInfo(file, fallbackName)
		const currentGroup = groupedFiles.get(groupKey)
		if (currentGroup) {
			currentGroup.files.push(file)
			continue
		}

		groupedFiles.set(groupKey, {
			name: groupName,
			files: [file],
		})
	}

	return Array.from(groupedFiles.entries())
		.map(([groupKey, group]) =>
			buildStructuredBundle(groupKey, group.name, group.files, kind),
		)
		.sort((left, right) => left.name.localeCompare(right.name))
}

export const mergeStagedBundles = (
	currentBundles: StagedCLIBundle[],
	incomingBundles: StagedCLIBundle[],
) => {
	const mergedBundles = new Map(
		currentBundles.map(bundle => [bundle.key, bundle] as const),
	)

	for (const bundle of incomingBundles) {
		mergedBundles.set(bundle.key, bundle)
	}

	return Array.from(mergedBundles.values()).sort((left, right) =>
		left.name.localeCompare(right.name),
	)
}

const encodeASCII = (value: string) => new TextEncoder().encode(value)

const writeString = (
	buffer: Uint8Array,
	offset: number,
	length: number,
	value: string,
) => {
	const encoded = encodeASCII(value)
	buffer.set(encoded.slice(0, length), offset)
}

const writeOctal = (
	buffer: Uint8Array,
	offset: number,
	length: number,
	value: number,
	withTrailingSpace = false,
) => {
	const octal = Math.max(0, value).toString(8)
	const trimmedLength = withTrailingSpace ? length - 2 : length - 1
	const body = octal.padStart(trimmedLength, "0")
	const encoded = encodeASCII(body)
	buffer.set(encoded.slice(0, trimmedLength), offset)
	buffer[offset + trimmedLength] = 0
	if (withTrailingSpace) {
		buffer[offset + length - 1] = 32
	}
}

const createTarHeader = (
	fileName: string,
	size: number,
	modifiedAtMs: number,
) => {
	const header = new Uint8Array(TAR_BLOCK_SIZE)
	const normalizedName = fileName.replace(/^\/+/, "").slice(0, 99)

	writeString(header, 0, 100, normalizedName)
	writeOctal(header, 100, 8, 0o644)
	writeOctal(header, 108, 8, 0)
	writeOctal(header, 116, 8, 0)
	writeOctal(header, 124, 12, size)
	writeOctal(header, 136, 12, Math.floor(modifiedAtMs / 1000))
	for (let index = 148; index < 156; index += 1) {
		header[index] = 32
	}
	header[156] = "0".charCodeAt(0)
	writeString(header, 257, 6, "ustar")
	writeString(header, 263, 2, "00")

	let checksum = 0
	for (const byte of header) {
		checksum += byte
	}
	writeOctal(header, 148, 8, checksum, true)

	return header
}

const createTarArchive = async (bundle: StagedCLIBundle) => {
	const chunks: Uint8Array[] = []

	for (const entry of bundle.entries) {
		const fileBytes = new Uint8Array(await entry.file.arrayBuffer())
		const archivePath = `${bundle.archiveRoot}/${entry.baseName}`
		const header = createTarHeader(
			archivePath,
			fileBytes.byteLength,
			entry.file.lastModified || Date.now(),
		)
		chunks.push(header)
		chunks.push(fileBytes)

		const padding = fileBytes.byteLength % TAR_BLOCK_SIZE
		if (padding > 0) {
			chunks.push(new Uint8Array(TAR_BLOCK_SIZE - padding))
		}
	}

	chunks.push(new Uint8Array(TAR_BLOCK_SIZE * 2))

	const totalLength = chunks.reduce((sum, chunk) => sum + chunk.byteLength, 0)
	const archive = new Uint8Array(totalLength)
	let offset = 0
	for (const chunk of chunks) {
		archive.set(chunk, offset)
		offset += chunk.byteLength
	}

	return archive
}

const gzipBytes = async (archiveBytes: Uint8Array) => {
	if (typeof CompressionStream === "undefined") {
		throw new Error(
			"Folder and JSON imports require a browser with gzip compression support",
		)
	}

	const compressedStream = new Blob([archiveBytes])
		.stream()
		.pipeThrough(new CompressionStream("gzip"))
	const compressedBuffer = await new Response(compressedStream).arrayBuffer()
	return new Uint8Array(compressedBuffer)
}

export const materializeBundleFile = async (bundle: StagedCLIBundle) => {
	if (bundle.archiveFile) {
		return bundle.archiveFile
	}

	const tarArchive = await createTarArchive(bundle)
	const compressedArchive = await gzipBytes(tarArchive)

	return new File([compressedArchive], `${bundle.archiveRoot}.tar.gz`, {
		type: "application/gzip",
		lastModified: Date.now(),
	})
}

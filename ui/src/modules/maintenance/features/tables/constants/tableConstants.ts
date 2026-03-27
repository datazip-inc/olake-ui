export const PAGE_SIZE = 6

export const compactionSlots: Array<{
	key: "minor" | "major" | "full"
	tag: "L" | "M" | "F"
	name: string
}> = [
	{ key: "minor", tag: "L", name: "Lite" },
	{ key: "major", tag: "M", name: "Medium" },
	{ key: "full", tag: "F", name: "Full" },
]

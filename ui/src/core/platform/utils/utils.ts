import { formatDate } from "@/common/utils"
import { ReleasesResponse, ReleaseType, ReleaseTypeData } from "../types"

/* Processes release data for UI consumption
 * - Converts ISO dates to readable format: "2026-01-17T10:00:00Z" -> "Released on Jan 17, 2026"
 * - Converts kebab-case tags to Title Case: "new-release" -> "New Release"
 *
 * Before: {
 *   olake_ui_worker: { releases: [{ date: "2026-01-17T10:00:00Z", tags: ["new-release"] }] },
 *   ...
 * }
 *
 * After: {
 *   olake_ui_worker: { releases: [{ date: "Released on Jan 17, 2026", tags: ["New Release"] }] },
 *   ...
 * }
 */
export const processReleasesData = (
	releases: ReleasesResponse | null,
): ReleasesResponse | null => {
	if (!releases) {
		return null
	}

	const formatReleaseData = (releaseTypeData?: ReleaseTypeData) => {
		if (!releaseTypeData) {
			return undefined
		}
		return {
			...releaseTypeData,
			releases: releaseTypeData.releases.map(release => ({
				...release,
				date: `Released on ${formatDate(release.date)}`,
				tags: release.tags.map(tag =>
					tag
						.replace(/-/g, " ")
						.split(" ")
						.map(word => word.charAt(0).toUpperCase() + word.slice(1))
						.join(" "),
				),
			})),
		}
	}
	return {
		[ReleaseType.OLAKE_UI_WORKER]: formatReleaseData(
			releases[ReleaseType.OLAKE_UI_WORKER],
		),
		[ReleaseType.OLAKE_HELM]: formatReleaseData(
			releases[ReleaseType.OLAKE_HELM],
		),
		[ReleaseType.OLAKE]: formatReleaseData(releases[ReleaseType.OLAKE]),
		[ReleaseType.FEATURES]: formatReleaseData(releases[ReleaseType.FEATURES]),
	}
}

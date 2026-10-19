import { useMemo, useRef } from "react"

import { SOURCE_SECTION_DOC_ANCHORS } from "@/modules/ingestion/common/constants"
import { DocumentationPanelHandle } from "@/modules/ingestion/common/types"

// Scrolls the docs panel to the heading of the source form section the user clicks or tabs
// into. Pass docsPanelRef to DocumentationPanel and formContext to the source Form.
export const useSourceDocSections = () => {
	const docsPanelRef = useRef<DocumentationPanelHandle>(null)
	const formContext = useMemo(
		() => ({
			onSectionFocus: (sectionId: string) => {
				const anchor = SOURCE_SECTION_DOC_ANCHORS[sectionId]
				if (anchor) docsPanelRef.current?.showAnchor(anchor)
			},
		}),
		[],
	)
	return { docsPanelRef, formContext }
}

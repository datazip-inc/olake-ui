import type { MenuProps } from "antd"

export type ActionMenuItemConfig = NonNullable<MenuProps["items"]>[number] & {
	feature?: string
}

export const resolveActionMenuItem = (
	item: ActionMenuItemConfig,
): ActionMenuItemConfig | null => item

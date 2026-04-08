import clsx from "clsx"
import type { HTMLAttributes } from "react"

export type TagColor =
	| "primary"
	| "success"
	| "warning"
	| "error"
	| "default"
	| "muted"

const tagBaseClassName =
	"inline-flex h-5 min-h-0 max-w-full shrink-0 items-center justify-center whitespace-nowrap rounded-[20px] border-0 px-2 py-0 text-[10px] font-medium leading-[20px]"

const tagColorClassName: Record<TagColor, string> = {
	primary: "bg-olake-primary-bg text-olake-primary",
	success: "bg-olake-success-bg text-olake-success",
	warning: "bg-olake-warning-bg text-olake-warning",
	error: "bg-olake-error-bg text-olake-error",
	default: "bg-gray-100 text-gray-700",
	muted: "bg-olake-surface-muted text-olake-text-secondary",
}

export type TagProps = Omit<HTMLAttributes<HTMLSpanElement>, "color"> & {
	children: React.ReactNode
	color?: TagColor
}

export function Tag({
	className,
	children,
	color = "primary",
	...rest
}: TagProps) {
	return (
		<span
			className={clsx(tagBaseClassName, tagColorClassName[color], className)}
			{...rest}
		>
			{children}
		</span>
	)
}

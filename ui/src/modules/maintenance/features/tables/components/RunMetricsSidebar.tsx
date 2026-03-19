import { XIcon } from "@phosphor-icons/react"
import { Button, Drawer, Spin } from "antd"
import { useMemo } from "react"

type RunMetricsSidebarProps = {
	open: boolean
	onClose: () => void
	payload: unknown
	loading?: boolean
	runId?: string
}

const isRecord = (value: unknown): value is Record<string, unknown> =>
	typeof value === "object" && value !== null && !Array.isArray(value)

const normalizeKey = (key: string): string =>
	key
		.split(/[_-]+/)
		.filter(Boolean)
		.map(part => part.charAt(0).toUpperCase() + part.slice(1))
		.join(" ")

const toDisplayValue = (value: unknown): string => {
	if (value === null || value === undefined || value === "") return "--"
	if (typeof value === "boolean") return value ? "True" : "False"
	if (typeof value === "object") return JSON.stringify(value)
	return String(value)
}

const extractMetricsObject = (payload: unknown): Record<string, unknown> => {
	if (!isRecord(payload)) return {}

	const data = payload.data
	if (isRecord(data)) {
		const nestedMetrics = data.metrics
		if (isRecord(nestedMetrics)) return nestedMetrics
	}

	const rootMetrics = payload.metrics
	if (isRecord(rootMetrics)) return rootMetrics

	return payload
}

const RunMetricsSidebar: React.FC<RunMetricsSidebarProps> = ({
	open,
	onClose,
	payload,
	loading = false,
	runId,
}) => {
	const rows = useMemo(() => {
		const metricsObj = extractMetricsObject(payload)
		return Object.entries(metricsObj).map(([key, value]) => ({
			label: normalizeKey(key),
			value: toDisplayValue(value),
		}))
	}, [payload])

	return (
		<Drawer
			placement="right"
			open={open}
			onClose={onClose}
			destroyOnHidden
			width={707}
			closeIcon={null}
			styles={{
				body: {
					padding: 0,
				},
			}}
		>
			<div className="min-h-full bg-olake-surface">
				<div className="relative px-7 pb-5 pt-10">
					<Button
						type="text"
						icon={<XIcon size={24} />}
						onClick={onClose}
						className="absolute right-4 top-4 text-olake-text-tertiary"
					/>
					<h2 className="font-sans text-xl font-medium leading-7 text-olake-text">
						Run Metrics for{" "}
						<span className="text-olake-primary">Run ID {runId ?? "--"}</span>
					</h2>
					<p className="mt-1 font-sans text-base font-normal leading-normal text-olake-text">
						View run metrics for the last run 2 days ago
					</p>
				</div>

				{loading ? (
					<div className="flex h-40 items-center justify-center">
						<Spin />
					</div>
				) : rows.length === 0 ? (
					<div className="px-7 py-8 font-sans text-sm font-normal leading-5 text-olake-text-tertiary">
						No metrics found.
					</div>
				) : (
					<div>
						{rows.map(item => (
							<div
								key={item.label}
								className="flex h-16 items-center justify-between border-b border-olake-border px-6"
							>
								<p className="font-sans text-base font-normal leading-6 text-olake-text">
									{item.label}
								</p>
								<p className="font-sans text-xl font-medium leading-7 text-olake-text">
									{item.value}
								</p>
							</div>
						))}
					</div>
				)}
			</div>
		</Drawer>
	)
}

export default RunMetricsSidebar

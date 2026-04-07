import { Button } from "antd"

type PageErrorStateProps = {
	title: string
	description: string
	onRetry: () => void
}

const PageErrorState: React.FC<PageErrorStateProps> = ({
	title,
	description,
	onRetry,
}) => {
	return (
		<div className="flex h-56 items-center justify-center rounded-lg border border-olake-border px-6">
			<div className="text-center">
				<p className="text-xl font-medium leading-7 text-olake-heading-strong">
					{title}
				</p>
				<p className="mt-1 text-sm leading-[22px] text-olake-body">
					{description}
				</p>
				<Button
					type="primary"
					className="mt-4"
					onClick={onRetry}
				>
					Retry
				</Button>
			</div>
		</div>
	)
}

export default PageErrorState

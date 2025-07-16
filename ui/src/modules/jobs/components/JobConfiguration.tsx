import { Input, Select, Radio } from "antd"
import StepTitle from "../../common/components/StepTitle"
import { generateCronExpression } from "../../../utils/utils"
import { JobConfigurationProps } from "../../../types"
import { useEffect, useState } from "react"
import { DAYS, FREQUENCY_OPTIONS } from "../../../utils/constants"
import parser from "cron-parser"

const JobConfiguration: React.FC<JobConfigurationProps> = ({
	jobName,
	setJobName,
	cronExpression = "* * * * *",
	setCronExpression,
	stepNumber = 4,
	stepTitle = "Job Configuration",
}) => {
	const [selectedTime, setSelectedTime] = useState("1")
	const [selectedAmPm, setSelectedAmPm] = useState<"AM" | "PM">("AM")
	const [selectedDay, setSelectedDay] = useState("Sunday")
	const [frequency, setFrequency] = useState("minutes")
	const [customCronExpression, setCustomCronExpression] = useState("")
	const [cronValue, setCronValue] = useState(cronExpression || "* * * * *")
	const [nextRuns, setNextRuns] = useState<string[]>([])

	// Configuration object for all select options
	const selectConfig = {
		frequency: FREQUENCY_OPTIONS,
		time: Array.from({ length: 12 }, (_, i) => ({
			value: (i + 1).toString(),
			label: (i + 1).toString(),
		})),
		days: DAYS.map(day => ({ value: day, label: day })),
	}

	const getParsedDate = (value: Date) => value.toUTCString()

	const updateNextRuns = (cronValue: string) => {
		try {
			const interval = parser.parse(cronValue, {
				currentDate: new Date(),
				tz: "UTC",
			})
			const data = []
			for (let i = 0; i < 3; i++) {
				data.push(getParsedDate(interval.next().toDate()))
			}
			setNextRuns(data)
		} catch (error) {
			// Clear next runs if cron expression is invalid
			console.error(
				"Invalid cron expression:",
				error instanceof Error ? error.message : String(error),
			)
			setNextRuns([])
		}
	}

	// Parse initial cron expression and set states
	useEffect(() => {
		if (!cronExpression) return

		try {
			const parts = cronExpression.split(" ")
			if (parts.length !== 5) return

			const [minute, hour, dayOfMonth, month, dayOfWeek] = parts

			// Determine frequency and set states based on cron pattern
			if (minute === "*" && hour === "*") {
				setFrequency("minutes")
			} else if (minute === "0" && hour === "*") {
				setFrequency("hours")
			} else if (
				minute === "0" &&
				dayOfMonth === "*" &&
				month === "*" &&
				dayOfWeek === "*"
			) {
				setFrequency("days")
				const hourNum = parseInt(hour)
				setSelectedTime(
					hourNum > 12
						? (hourNum - 12).toString()
						: hourNum === 0
							? "12"
							: hourNum.toString(),
				)
				setSelectedAmPm(hourNum >= 12 ? "PM" : "AM")
			} else if (
				minute === "0" &&
				dayOfMonth === "*" &&
				month === "*" &&
				/^[0-6]$/.test(dayOfWeek)
			) {
				setFrequency("weeks")
				const hourNum = parseInt(hour)
				setSelectedTime(
					hourNum > 12
						? (hourNum - 12).toString()
						: hourNum === 0
							? "12"
							: hourNum.toString(),
				)
				setSelectedAmPm(hourNum >= 12 ? "PM" : "AM")
				setSelectedDay(DAYS[parseInt(dayOfWeek)])
			} else {
				setFrequency("custom")
				setCustomCronExpression(cronExpression)
			}

			setCronValue(cronExpression)
		} catch (error) {
			console.error("Error parsing cron expression:", error)
			setFrequency("minutes")
		}
	}, [cronExpression])

	useEffect(() => {
		if (cronValue) {
			updateNextRuns(cronValue)
		}
	}, [cronValue])

	// Unified handler for all cron expression updates
	const updateCronExpression = (
		freq?: string,
		time?: string,
		amPm?: "AM" | "PM",
		day?: string,
	) => {
		const f = freq || frequency
		const t = time || selectedTime
		const ap = amPm || selectedAmPm
		const d = day || selectedDay

		if (f === "custom") {
			setCronExpression(customCronExpression)
			setCronValue(customCronExpression)
		} else {
			const newCronExpression = generateCronExpression(f, t, ap, d)
			setCronExpression(newCronExpression)
			setCronValue(newCronExpression)
		}
	}

	const handleFrequencyChange = (selectedUnit: string) => {
		setFrequency(selectedUnit)
		updateCronExpression(selectedUnit)
	}

	const handleTimeChange = (value: string) => {
		setSelectedTime(value)
		if (frequency !== "custom") {
			updateCronExpression(undefined, value)
		}
	}

	const handleAmPmChange = (value: "AM" | "PM") => {
		setSelectedAmPm(value)
		if (frequency !== "custom") {
			updateCronExpression(undefined, undefined, value)
		}
	}

	const handleDayChange = (value: string) => {
		setSelectedDay(value)
		if (frequency !== "custom") {
			updateCronExpression(undefined, undefined, undefined, value)
		}
	}

	const handleCustomCronChange = (value: string) => {
		setCustomCronExpression(value)
		setCronExpression(value)
		setCronValue(value)
	}

	// Helper to determine if time selection should be shown
	const isTimeSelectionFrequency = (freq: string): boolean => {
		return freq === "days" || freq === "weeks"
	}

	const shouldShowTimeSelection =
		isTimeSelectionFrequency(frequency) && frequency !== "custom"

	return (
		<div className="w-full p-6">
			{stepNumber && stepTitle && (
				<StepTitle
					stepNumber={stepNumber}
					stepTitle={stepTitle}
				/>
			)}
			<div className="flex flex-col gap-4 rounded-xl border border-[#D9D9D9] p-4">
				<div className="flex gap-4">
					<div className="mb-4 w-2/5">
						<label className="mb-2 block text-sm font-medium">
							Job name:<span className="text-red-500">*</span>
						</label>
						<Input
							placeholder="Enter your job name"
							value={jobName}
							onChange={e => setJobName(e.target.value)}
						/>
					</div>

					<div className="mb-4 w-3/5">
						<div className="flex gap-4">
							{/* Frequency Select */}
							<div>
								<label className="mb-2 block text-sm">Frequency</label>
								<Select
									className="w-40"
									value={frequency}
									onChange={handleFrequencyChange}
									options={selectConfig.frequency}
								/>
							</div>

							{/* Custom Cron Input */}
							{frequency === "custom" && (
								<div>
									<label className="mb-2 block text-sm">Cron Expression</label>
									<Input
										className="w-64"
										placeholder="Enter cron expression (e.g., */5 * * * *)"
										value={customCronExpression}
										onChange={e => handleCustomCronChange(e.target.value)}
									/>
								</div>
							)}

							{/* Day Select - only for weekly frequency */}
							{frequency === "weeks" && (
								<div>
									<label className="mb-2 block text-sm">Select Day</label>
									<Select
										className="w-36"
										value={selectedDay}
										onChange={handleDayChange}
										options={selectConfig.days}
										placeholder="Select Day"
									/>
								</div>
							)}

							{/* Time Selection - for daily and weekly frequencies */}
							{shouldShowTimeSelection && (
								<div className={frequency === "weeks" ? "" : "ml-4"}>
									<label className="mb-2 block text-sm">
										Job Start Time{" "}
										<span className="text-[#A7A7A7]">(12H Format UTC)</span>
									</label>
									<div className="flex items-center gap-1">
										<Select
											className="w-24"
											value={selectedTime}
											onChange={handleTimeChange}
											options={selectConfig.time}
										/>
										<Radio.Group
											value={selectedAmPm}
											onChange={e => handleAmPmChange(e.target.value)}
										>
											<Radio.Button value="AM">AM</Radio.Button>
											<Radio.Button value="PM">PM</Radio.Button>
										</Radio.Group>
									</div>
								</div>
							)}
						</div>
					</div>
				</div>
				{nextRuns.length > 0 && (
					<div className="flex gap-2">
						<span className="text-sm font-medium">Next 3 Runs (UTC):</span>
						<div className="flex gap-4">
							{nextRuns.map((run, index) => (
								<span
									key={index}
									className="text-sm text-gray-600"
								>
									{run}
								</span>
							))}
						</div>
					</div>
				)}
			</div>
		</div>
	)
}

export default JobConfiguration

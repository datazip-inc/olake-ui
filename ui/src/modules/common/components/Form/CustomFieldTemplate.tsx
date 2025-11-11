import { FieldTemplateProps } from "@rjsf/utils"
import { InfoIcon, PlusIcon, TrashIcon } from "@phosphor-icons/react"
import { Tooltip, Button } from "antd"
import { useState, useEffect } from "react"

// --- KeyValueRow: Renders a single key-value pair with edit/delete ---
function KeyValueRow({
	keyName,
	value,
	onKeyChange,
	onValueChange,
	onDelete,
}: {
	keyName: string
	value: string
	onKeyChange: (oldKey: string, newKey: string, value: string) => void
	onValueChange: (key: string, value: string) => void
	onDelete: (key: string) => void
}) {
	const [editedKey, setEditedKey] = useState(keyName)

	useEffect(() => {
		setEditedKey(keyName)
	}, [keyName])
	return (
		<div className="flex items-center gap-2">
			<input
				type="text"
				value={editedKey}
				onChange={e => setEditedKey(e.target.value)}
				onBlur={() => onKeyChange(keyName, editedKey, value)}
				className="h-8 w-1/3 rounded-[6px] border border-gray-300 px-3 py-2 text-sm shadow-sm focus:border-blue-500 focus:outline-none focus:ring-1 focus:ring-blue-500"
				placeholder="Key"
			/>
			<input
				type="text"
				value={value}
				onChange={e => onValueChange(keyName, e.target.value)}
				className="h-8 w-1/3 rounded-[6px] border border-gray-300 px-3 py-2 text-sm shadow-sm focus:border-blue-500 focus:outline-none focus:ring-1 focus:ring-blue-500"
				placeholder="Value"
			/>
			<Button
				type="text"
				onClick={() => onDelete(keyName)}
				icon={<TrashIcon className="text-red-500" />}
			/>
		</div>
	)
}

// --- NewKeyValueRow: Renders inputs for adding a new key-value pair for type = object ---
function NewKeyValueRow({
	newKey,
	newValue,
	setNewKey,
	setNewValue,
	onAdd,
}: {
	newKey: string
	newValue: string
	setNewKey: (k: string) => void
	setNewValue: (v: string) => void
	onAdd: () => void
}) {
	return (
		<div className="flex items-center gap-2">
			<input
				type="text"
				value={newKey}
				onChange={e => setNewKey(e.target.value)}
				className="h-8 w-1/3 rounded-[6px] border border-gray-300 px-3 py-2 text-sm shadow-sm focus:border-blue-500 focus:outline-none focus:ring-1 focus:ring-blue-500"
				placeholder="New Key"
			/>
			<input
				type="text"
				value={newValue}
				onChange={e => setNewValue(e.target.value)}
				className="h-8 w-1/3 rounded-[6px] border border-gray-300 px-3 py-2 text-sm shadow-sm focus:border-blue-500 focus:outline-none focus:ring-1 focus:ring-blue-500"
				placeholder="New Value"
			/>
			<Button
				type="text"
				onClick={onAdd}
				icon={<PlusIcon className="text-blue-500" />}
			/>
		</div>
	)
}

/**
 * CustomFieldTemplate is a component that renders a custom field template this overrides the default field template of rjsf Field specific layout is defined here
 */

export default function CustomFieldTemplate(props: FieldTemplateProps) {
	// --- Props Destructure ---
	const {
		id,
		label,
		displayLabel,
		description,
		rawDescription,
		children,
		required,
		schema,
		formData,
		onChange,
		rawErrors,
	} = props

	// --- Derived Variables ---
	const hasDescription = !!rawDescription
	const isDynamicObject = schema.additionalProperties === true
	const objectTitle =
		typeof schema.title === "string" && schema.title.trim().length > 0
			? (schema.title as string)
			: undefined
	const isArrayField = schema.type === "array"
	const shouldShowLabel =
		(isDynamicObject && !!objectTitle) ||
		(isArrayField && !!objectTitle) ||
		(displayLabel && !!label)
	const labelText = objectTitle ?? label
	const [newKey, setNewKey] = useState("")
	const [newValue, setNewValue] = useState("")
	const [currentTempKey, setCurrentTempKey] = useState<string | null>(null)
	const inputErrorWrapperClass =
		rawErrors && rawErrors.length > 0 ? "rjsf-error" : ""

	const handleAddKeyValue = () => {
		if (!newKey.trim()) return
		const updatedFormData = {
			...(formData || {}),
			[newKey]: newValue,
		}
		onChange(updatedFormData)
		setNewKey("")
		setNewValue("")
		setCurrentTempKey(null)
	}

	const handleDeleteKeyValue = (key: string) => {
		const updatedFormData = { ...formData }
		delete updatedFormData[key]

		// If we're deleting the current temp key, clear the reference
		if (currentTempKey === key) {
			setCurrentTempKey(null)
		}

		onChange(updatedFormData)
	}

	const handleKeyChange = (oldKey: string, newKey: string, value: string) => {
		if (!newKey.trim()) return
		const updatedFormData = { ...formData }
		delete updatedFormData[oldKey]
		updatedFormData[newKey] = value
		onChange(updatedFormData)
	}

	const handleValueChange = (key: string, value: string) => {
		const updatedFormData = {
			...formData,
			[key]: value,
		}
		onChange(updatedFormData)
	}

	const handleNewKeyInputChange = (nextKey: string) => {
		const trimmedNextKey = nextKey.trim()
		setNewKey(nextKey)

		const updatedFormData = { ...(formData || {}) }

		// Remove the old temp key if it exists and is different
		if (currentTempKey && currentTempKey !== trimmedNextKey) {
			delete updatedFormData[currentTempKey]
		}

		// Create/update the new temp key if both key and value exist
		if (trimmedNextKey && newValue.trim()) {
			updatedFormData[trimmedNextKey] = newValue
			setCurrentTempKey(trimmedNextKey)
		} else {
			setCurrentTempKey(null)
		}

		onChange(updatedFormData)
	}

	const handleNewValueInputChange = (nextValue: string) => {
		const trimmedNextValue = nextValue.trim()
		const trimmedKey = newKey.trim()
		setNewValue(nextValue)

		const updatedFormData = { ...(formData || {}) }

		// If we have a key, create or update the temp pair
		if (trimmedKey) {
			if (trimmedNextValue) {
				updatedFormData[trimmedKey] = trimmedNextValue
				setCurrentTempKey(trimmedKey)
			} else if (currentTempKey === trimmedKey) {
				// Value was cleared, remove the temp key
				delete updatedFormData[trimmedKey]
				setCurrentTempKey(null)
			}
		}

		onChange(updatedFormData)
	}

	// --- Render ---
	return (
		<div className="mb-2">
			{shouldShowLabel && (
				<label
					htmlFor={id}
					className="mb-2 mt-1 flex items-center gap-1 text-sm font-medium text-gray-700"
				>
					{labelText}
					{required && <span className="text-red-500">*</span>}
					{hasDescription && (
						<Tooltip
							title={description || rawDescription}
							placement="right"
						>
							<InfoIcon className="ml-1 text-gray-500 hover:text-gray-600" />
						</Tooltip>
					)}
				</label>
			)}

			{isDynamicObject ? (
				<div className="space-y-3">
					{/* Existing key-value pairs */}
					{formData &&
						Object.entries(formData)
							.filter(([key]) => key !== currentTempKey)
							.map(([key, value]) => (
								<KeyValueRow
									key={key}
									keyName={key}
									value={value as string}
									onKeyChange={handleKeyChange}
									onValueChange={handleValueChange}
									onDelete={handleDeleteKeyValue}
								/>
							))}
					{/* Add new key-value pair */}
					<NewKeyValueRow
						newKey={newKey}
						newValue={newValue}
						setNewKey={handleNewKeyInputChange}
						setNewValue={handleNewValueInputChange}
						onAdd={handleAddKeyValue}
					/>
				</div>
			) : (
				<div className={inputErrorWrapperClass}>{children}</div>
			)}
		</div>
	)
}

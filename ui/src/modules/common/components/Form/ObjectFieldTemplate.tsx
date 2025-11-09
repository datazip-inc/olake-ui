import clsx from "clsx"
import {
	FormContextType,
	GenericObjectType,
	ObjectFieldTemplateProps,
	ObjectFieldTemplatePropertyType,
	RJSFSchema,
	StrictRJSFSchema,
	UiSchema,
	canExpand,
	descriptionId,
	getTemplate,
	getUiOptions,
	titleId,
} from "@rjsf/utils"
import Col from "antd/lib/col"
import Row from "antd/lib/row"
import {
	ConfigConsumer,
	ConfigConsumerProps,
} from "antd/lib/config-provider/context"

const DESCRIPTION_COL_STYLE = {
	paddingBottom: "8px",
}

/** The `ObjectFieldTemplate` is the template to use to render all the inner properties of an object along with the
 * title and description if available. If the object is expandable, then an `AddButton` is also rendered after all
 * the properties. This component defines the layout for the object field template.
 *
 * @param props - The `ObjectFieldTemplateProps` for this component
 */
export default function ObjectFieldTemplate<
	T = any,
	S extends StrictRJSFSchema = RJSFSchema,
	F extends FormContextType = any,
>(props: ObjectFieldTemplateProps<T, S, F>) {
	const {
		description,
		disabled,
		formContext,
		formData,
		idSchema,
		onAddClick,
		properties,
		readonly,
		required,
		registry,
		schema,
		title,
		uiSchema,
	} = props
	const uiOptions = getUiOptions<T, S, F>(uiSchema)
	const TitleFieldTemplate = getTemplate<"TitleFieldTemplate", T, S, F>(
		"TitleFieldTemplate",
		registry,
		uiOptions,
	)
	const DescriptionFieldTemplate = getTemplate<
		"DescriptionFieldTemplate",
		T,
		S,
		F
	>("DescriptionFieldTemplate", registry, uiOptions)
	// Button templates are not overridden in the uiSchema
	const {
		ButtonTemplates: { AddButton },
	} = registry.templates
	const {
		colSpan = 24,
		labelAlign = "left",
		rowGutter = 24,
	} = formContext as GenericObjectType

	const findSchema = (element: ObjectFieldTemplatePropertyType): S =>
		element.content.props.schema

	const findSchemaType = (element: ObjectFieldTemplatePropertyType) =>
		findSchema(element).type

	const findUiSchema = (
		element: ObjectFieldTemplatePropertyType,
	): UiSchema<T, S, F> | undefined => element.content.props.uiSchema

	const findUiSchemaField = (element: ObjectFieldTemplatePropertyType) =>
		getUiOptions(findUiSchema(element)).field

	const findUiSchemaWidget = (element: ObjectFieldTemplatePropertyType) =>
		getUiOptions(findUiSchema(element)).widget

	const calculateColSpan = (element: ObjectFieldTemplatePropertyType) => {
		const type = findSchemaType(element)
		const field = findUiSchemaField(element)
		const widget = findUiSchemaWidget(element)

		const defaultColSpan =
			properties.length < 2 || // Single or no field in object.
			type === "object" ||
			type === "array" ||
			widget === "textarea"
				? 24
				: 12

		if (typeof colSpan === "object" && colSpan !== null) {
			const colSpanObj: GenericObjectType = colSpan
			// Only use string keys that exist in colSpanObj
			if (typeof widget === "string" && widget in colSpanObj) {
				return colSpanObj[widget]
			}
			if (typeof field === "string" && field in colSpanObj) {
				return colSpanObj[field]
			}
			if (typeof type === "string" && type in colSpanObj) {
				return colSpanObj[type]
			}
		}
		if (typeof colSpan === "number") {
			return colSpan
		}
		return defaultColSpan
	}

	return (
		<ConfigConsumer>
			{(configProps: ConfigConsumerProps) => {
				const { getPrefixCls } = configProps
				const prefixCls = getPrefixCls("form")
				const labelClsBasic = `${prefixCls}-item-label`
				const labelColClassName = clsx(
					labelClsBasic,
					labelAlign === "left" && `${labelClsBasic}-left`,
					// labelCol.className,
				)

				return (
					<fieldset id={idSchema.$id}>
						<Row gutter={rowGutter}>
							{title && (
								<Col
									className={labelColClassName}
									span={24}
								>
									<TitleFieldTemplate
										id={titleId<T>(idSchema)}
										title={title}
										required={required}
										schema={schema}
										uiSchema={uiSchema}
										registry={registry}
									/>
								</Col>
							)}
							{description && (
								<Col
									span={24}
									style={DESCRIPTION_COL_STYLE}
								>
									<DescriptionFieldTemplate
										id={descriptionId<T>(idSchema)}
										description={description}
										schema={schema}
										uiSchema={uiSchema}
										registry={registry}
									/>
								</Col>
							)}
							{uiSchema?.["ui:grid"] && Array.isArray(uiSchema["ui:grid"])
								? uiSchema["ui:grid"].map((ui_row, rowIdx) => {
										const rowElements = Object.keys(ui_row).map(row_item => {
											const element = properties.find(p => p.name == row_item)
											if (element) {
												return (
													<Col
														key={element.name}
														span={ui_row[row_item]}
													>
														{element.content}
													</Col>
												)
											}
											return null
										})
										// Calculate total span for the row
										const totalSpan = Object.values(ui_row).reduce(
											(sum: number, val) =>
												sum + (typeof val === "number" ? val : 0),
											0,
										)
										// If totalSpan < 24, add an empty Col to fill the row
										if (totalSpan < 24) {
											rowElements.push(
												<Col
													key={`empty-col-${rowIdx}`}
													span={24 - totalSpan}
												></Col>,
											)
										}
										return rowElements
									})
								: properties
										.filter(e => !e.hidden)
										.map((element: ObjectFieldTemplatePropertyType) => (
											<Col
												key={element.name}
												span={calculateColSpan(element)}
											>
												{element.content}
											</Col>
										))}
						</Row>

						{canExpand(schema, uiSchema, formData) && (
							<Col span={24}>
								<Row
									gutter={rowGutter}
									justify="end"
								>
									<Col flex="192px">
										<AddButton
											className="object-property-expand"
											disabled={disabled || readonly}
											onClick={onAddClick(schema)}
											uiSchema={uiSchema}
											registry={registry}
										/>
									</Col>
								</Row>
							</Col>
						)}
					</fieldset>
				)
			}}
		</ConfigConsumer>
	)
}

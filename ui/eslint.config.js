import globals from "globals"
import reactHooks from "eslint-plugin-react-hooks"
import reactRefresh from "eslint-plugin-react-refresh"
import tseslint from "@typescript-eslint/eslint-plugin"
import tsParser from "@typescript-eslint/parser"
import react from "eslint-plugin-react"
import eslintPluginPrettier from "eslint-plugin-prettier"

export default [
	{
		files: ["**/*.{ts,tsx,js,jsx}"],
		languageOptions: {
			ecmaVersion: 2020,
			parser: tsParser,
			globals: globals.browser,
		},
		plugins: {
			"react-hooks": reactHooks,
			"react-refresh": reactRefresh,
			"@typescript-eslint": tseslint,
			react,
			prettier: eslintPluginPrettier,
		},
		rules: {
			...(reactRefresh.configs ? reactRefresh.configs.recommended.rules : {}),
			...tseslint.configs.recommended.rules,
			...react.configs.recommended.rules,
			"react/react-in-jsx-scope": "off", // Disable the need for React to be in scope with JSX
			"react/prop-types": "off", // Disable prop-types rule for TypeScript
			"react-refresh/only-export-components": "off", // Disable Fast Refresh rule
			"@typescript-eslint/ban-ts-comment": "warn",
			"@typescript-eslint/no-explicit-any": "off",
			"prettier/prettier": "warn",
			"react-refresh/only-export-components": [
				"warn",
				{ allowConstantExport: true },
			],
		},
		settings: {
			react: {
				version: "detect",
			},
		},
	},
	{
		ignores: ["node_modules/", "dist/", "build/", "src/components"],
	},
]

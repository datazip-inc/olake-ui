/** @type {import('tailwindcss').Config} */
export default {
	darkMode: ["class"],
	content: ["./index.html", "./src/**/*.{js,ts,jsx,tsx}"],
	theme: {
		extend: {
			fontFamily: {
				sans: ['"Helvetica Neue"'],
			},
			colors: {
				primary: {
					DEFAULT: "#203FDD",
					50: "#F5F7FF",
					100: "#E9EBFC",
					200: "#E6F4FF",
					500: "#203FDD",
					600: "#132685",
					700: "#0958D9",
				},
				brand: {
					DEFAULT: "#203FDD",
					light: "#E9EBFC",
					lighter: "#F5F7FF",
					dark: "#132685",
					hover: "#132685",
				},
				neutral: {
					light: "#F0F0F0",
					border: "#E3E3E3",
					text: "#575757",
					disabled: "#D9D9D9",
				},
			},
		},
	},
}

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
					blue: "#193AE6",
				},
				neutral: {
					light: "#F0F0F0",
					border: "#E3E3E3",
					text: "#575757",
					disabled: "#D9D9D9",
				},
				danger: {
					DEFAULT: "#F5222D",
					light: "#FFF1F0",
					dark: "#b81922",
				},
				success: {
					DEFAULT: "#389E0D",
				},
				warning: {
					DEFAULT: "#DAAC06",
					light: "#FFF7E6",
					dark: "#6E5807",
				},
				gray: {
					50: "#f8f8f8",
					100: "#f6f6f6",
					200: "#E5E7EB",
					300: "#D1D5DB",
					400: "#D9D9D9",
					500: "#A7A7A7",
					600: "#4B5563",
					700: "#374151",
					800: "#575757",
					900: "#383838",
					950: "#2B2B2B",
				},
				text: {
					primary: "#0A0A0A",
					secondary: "#575757",
					tertiary: "#8A8A8A",
					muted: "#A3A3A3",
					disabled: "#c1c1c1",
					placeholder: "#9F9F9F",
					link: "#6E6E6E",
				},
			},
		},
	},
}

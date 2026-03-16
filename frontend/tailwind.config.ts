import type { Config } from "tailwindcss";

const config: Config = {
  content: ["./index.html", "./src/**/*.{ts,tsx}"],
  theme: {
    extend: {
      colors: {
        brand: {
          50: "#eff9ff",
          100: "#daf1ff",
          200: "#bde6ff",
          300: "#8ed8ff",
          400: "#58c2ff",
          500: "#289fff",
          600: "#157ff5",
          700: "#1364e1",
          800: "#1652b6",
          900: "#1a488f",
        },
      },
    },
  },
  plugins: [],
};

export default config;

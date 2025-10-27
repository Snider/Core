/** @type {import('tailwindcss').Config} */
module.exports = {
  content: [
    "./html/**/*.html",
    "./src/**/*.{js,ts,jsx,tsx}",
  ],
  theme: {
    extend: {
      colors: {
        'core-gray': {
          'light': '#333333',
          'DEFAULT': '#1a1a1a',
          'dark': '#0d0d0d',
        },
        'core-blue': '#00aaff',
      }
    },
  },
  plugins: [],
}

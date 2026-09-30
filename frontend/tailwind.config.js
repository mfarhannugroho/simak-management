/** @type {import('tailwindcss').Config} */
export default {
  content: [
    "./index.html",
    "./src/**/*.{vue,js,ts,jsx,tsx}",
  ],
  theme: {
    extend: {
      colors: {
        canvas: '#F9FAFB',
        card: '#FFFFFF',
        border: '#E5E7EB',
        inka: {
          50: '#FFEBEE',
          100: '#FFCDD2',
          200: '#EF9A9A',
          500: '#F44336',
          600: '#E53935',
          700: '#D32F2F', // Official Primary INKA Red
          800: '#C62828',
          900: '#B71C1C', // Hover Red
        },
        slate: {
          850: '#151F32',
          950: '#0B0F19',
        }
      },
      fontFamily: {
        sans: ['"Plus Jakarta Sans"', 'Inter', 'system-ui', '-apple-system', 'sans-serif'],
      },
      boxShadow: {
        subtle: '0 1px 2px 0 rgba(0, 0, 0, 0.05)',
        card: '0 1px 3px 0 rgba(0, 0, 0, 0.07), 0 1px 2px -1px rgba(0, 0, 0, 0.07)',
      },
      borderRadius: {
        'xl': '12px',
        '2xl': '16px',
      }
    },
  },
  plugins: [],
}

import type { Config } from 'tailwindcss'

export default {
  content: [
    './components/**/*.{vue,ts}',
    './layouts/**/*.vue',
    './pages/**/*.vue',
    './composables/**/*.ts',
    './app.vue',
  ],
  theme: {
    extend: {
      fontFamily: {
        display: ['Syne', 'sans-serif'],
        body: ['DM Sans', 'sans-serif'],
      },
      colors: {
        surface: {
          deep: '#06080D',
          base: '#0C0F17',
          DEFAULT: '#111621',
          elevated: '#171C28',
          hover: '#1D2335',
          border: '#242B3D',
        },
        accent: {
          DEFAULT: '#E8A838',
          hover: '#F0BE5A',
          muted: 'rgba(232,168,56,0.15)',
          subtle: 'rgba(232,168,56,0.08)',
        },
        muted: '#4A5268',
        secondary: '#8B95A8',
      },
      animation: {
        'fade-in-up': 'fade-in-up 0.55s ease-out both',
        'fade-in': 'fade-in 0.45s ease-out both',
        'scale-in': 'scale-in 0.45s ease-out both',
      },
      keyframes: {
        'fade-in-up': {
          from: { opacity: '0', transform: 'translateY(24px)' },
          to: { opacity: '1', transform: 'translateY(0)' },
        },
        'fade-in': {
          from: { opacity: '0' },
          to: { opacity: '1' },
        },
        'scale-in': {
          from: { opacity: '0', transform: 'scale(0.95)' },
          to: { opacity: '1', transform: 'scale(1)' },
        },
      },
    },
  },
  plugins: [],
} satisfies Config

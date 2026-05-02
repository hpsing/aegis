/** @type {import('tailwindcss').Config} */
export default {
  content: ['./index.html', './src/**/*.{vue,ts,js}'],
  darkMode: 'class',
  theme: {
    extend: {
      colors: {
        bg: {
          base: '#0b0d12',
          panel: '#11141b',
          subtle: '#181c25',
          border: '#222732',
        },
        ink: {
          0: '#f3f5f8',
          1: '#c2c8d3',
          2: '#7a8294',
          3: '#4a5160',
        },
        accent: {
          pass: '#3ddc97',
          fail: '#ff6b6b',
          chain: '#7aa6ff',
          axl: '#c084fc',
          kh: '#fbbf24',
          og: '#34d399',
        },
      },
      fontFamily: {
        mono: ['"JetBrains Mono"', 'ui-monospace', 'SFMono-Regular', 'Menlo', 'monospace'],
      },
    },
  },
  plugins: [],
}

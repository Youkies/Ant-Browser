import { createContext, useContext, useEffect, useMemo, useState, ReactNode } from 'react'
import { WindowSetBackgroundColour, WindowSetDarkTheme, WindowSetLightTheme, WindowSetSystemDefaultTheme } from '../../wailsjs/runtime/runtime'
import { ThemeType, DEFAULT_THEME } from './types'

interface ThemeContextValue {
  theme: ThemeType
  resolvedTheme: Exclude<ThemeType, 'system'>
  setTheme: (theme: ThemeType) => void
}

const ThemeContext = createContext<ThemeContextValue | undefined>(undefined)

const THEME_STORAGE_KEY = 'app-theme'
const APP_SETTINGS_STORAGE_KEY = 'app_settings'
const VALID_THEMES: ThemeType[] = ['system', 'light', 'dark']
const LEGACY_LIGHT_THEMES = new Set(['cream', 'mint', 'ocean'])

interface ThemeProviderProps {
  children: ReactNode
  defaultTheme?: ThemeType
}

function isThemeType(value: string | null): value is ThemeType {
  return !!value && VALID_THEMES.includes(value as ThemeType)
}

function normalizeThemeValue(value: string | null): ThemeType | null {
  if (isThemeType(value)) {
    return value
  }
  if (value && LEGACY_LIGHT_THEMES.has(value)) {
    return 'light'
  }
  return null
}

function resolveStoredTheme(defaultTheme: ThemeType): ThemeType {
  const settingsRaw = localStorage.getItem(APP_SETTINGS_STORAGE_KEY)
  if (settingsRaw) {
    try {
      const parsed = JSON.parse(settingsRaw) as { theme?: string }
      const normalizedTheme = normalizeThemeValue(parsed?.theme ?? null)
      if (normalizedTheme) {
        return normalizedTheme
      }
    } catch (error) {
      console.warn('Failed to parse app settings while restoring theme:', error)
    }
  }

  const savedTheme = localStorage.getItem(THEME_STORAGE_KEY)
  const normalizedSavedTheme = normalizeThemeValue(savedTheme)
  if (normalizedSavedTheme) {
    return normalizedSavedTheme
  }

  return defaultTheme
}

function getSystemTheme(): Exclude<ThemeType, 'system'> {
  return window.matchMedia('(prefers-color-scheme: dark)').matches ? 'dark' : 'light'
}

function syncThemeToSettings(theme: ThemeType) {
  try {
    const settingsRaw = localStorage.getItem(APP_SETTINGS_STORAGE_KEY)
    const settings = settingsRaw ? JSON.parse(settingsRaw) : {}
    localStorage.setItem(APP_SETTINGS_STORAGE_KEY, JSON.stringify({ ...settings, theme }))
  } catch (error) {
    console.warn('Failed to sync theme to app settings:', error)
  }
}

function applyWindowTheme(theme: ThemeType, resolvedTheme: Exclude<ThemeType, 'system'>) {
  try {
    if (theme === 'system') {
      WindowSetSystemDefaultTheme()
    } else if (resolvedTheme === 'dark') {
      WindowSetDarkTheme()
    } else {
      WindowSetLightTheme()
    }

    if (resolvedTheme === 'dark') {
      WindowSetBackgroundColour(12, 12, 14, 255)
    } else {
      WindowSetBackgroundColour(248, 250, 252, 255)
    }
  } catch (error) {
    console.warn('Failed to apply native window theme:', error)
  }
}

export function ThemeProvider({ children, defaultTheme = DEFAULT_THEME }: ThemeProviderProps) {
  const [theme, setThemeState] = useState<ThemeType>(() => resolveStoredTheme(defaultTheme))
  const [systemTheme, setSystemTheme] = useState<Exclude<ThemeType, 'system'>>(() => getSystemTheme())

  const resolvedTheme = useMemo<Exclude<ThemeType, 'system'>>(
    () => (theme === 'system' ? systemTheme : theme),
    [theme, systemTheme]
  )

  useEffect(() => {
    const mediaQuery = window.matchMedia('(prefers-color-scheme: dark)')
    const handleChange = (event: MediaQueryListEvent) => {
      setSystemTheme(event.matches ? 'dark' : 'light')
    }

    setSystemTheme(mediaQuery.matches ? 'dark' : 'light')
    mediaQuery.addEventListener('change', handleChange)
    return () => mediaQuery.removeEventListener('change', handleChange)
  }, [])

  const setTheme = (newTheme: ThemeType) => {
    setThemeState(newTheme)
    localStorage.setItem(THEME_STORAGE_KEY, newTheme)
    syncThemeToSettings(newTheme)
  }

  useEffect(() => {
    document.documentElement.setAttribute('data-theme', resolvedTheme)
    document.documentElement.setAttribute('data-theme-mode', theme)
    document.documentElement.style.colorScheme = resolvedTheme
    applyWindowTheme(theme, resolvedTheme)
  }, [theme, resolvedTheme])

  return (
    <ThemeContext.Provider value={{ theme, resolvedTheme, setTheme }}>
      {children}
    </ThemeContext.Provider>
  )
}

export function useTheme() {
  const context = useContext(ThemeContext)
  if (!context) {
    throw new Error('useTheme must be used within a ThemeProvider')
  }
  return context
}

import { Check, Laptop, MoonStar, SunMedium } from 'lucide-react'
import clsx from 'clsx'
import { useTheme, themeConfigs, ThemeType } from '../theme'

interface ThemeSwitcherProps {
  className?: string
  value?: ThemeType
  onChange?: (theme: ThemeType) => void
}

const themeIcons: Record<ThemeType, typeof Laptop> = {
  system: Laptop,
  light: SunMedium,
  dark: MoonStar,
}

const themePreview: Record<ThemeType, { shell: string; surface: string; accent: string; text: string }> = {
  system: { shell: 'linear-gradient(135deg, #f8fafc 0%, #e2e8f0 50%, #111827 50%, #0f172a 100%)', surface: '#ffffff', accent: '#334155', text: '#0f172a' },
  light: { shell: '#f8fafc', surface: '#ffffff', accent: '#1e293b', text: '#1e293b' },
  dark: { shell: '#09090b', surface: '#18181b', accent: '#f8fafc', text: '#f8fafc' },
}

export function ThemeSwitcher({ className, value, onChange }: ThemeSwitcherProps) {
  const themeContext = useTheme()
  const selectedTheme = value ?? themeContext.theme

  const handleSelect = (nextTheme: ThemeType) => {
    onChange?.(nextTheme)
    if (value === undefined) {
      themeContext.setTheme(nextTheme)
    }
  }

  return (
    <div className={clsx('space-y-4', className)}>
      <div className="grid grid-cols-1 gap-3 md:grid-cols-3">
        {themeConfigs.map((config) => {
          const isActive = selectedTheme === config.id
          const Icon = themeIcons[config.id]
          const preview = themePreview[config.id]

          return (
            <button
              key={config.id}
              type="button"
              onClick={() => handleSelect(config.id)}
              className={clsx(
                'group relative overflow-hidden rounded-2xl border p-4 text-left transition-all duration-200',
                isActive
                  ? 'border-[var(--color-accent)] bg-[var(--color-accent-muted)] shadow-[var(--shadow-md)]'
                  : 'border-[var(--color-border-default)] bg-[var(--color-bg-surface)] hover:-translate-y-0.5 hover:border-[var(--color-border-strong)] hover:shadow-[var(--shadow-sm)]'
              )}
              title={config.description}
            >
              <div className="flex items-start justify-between gap-3">
                <div className="space-y-1.5">
                  <div className="inline-flex h-10 w-10 items-center justify-center rounded-xl bg-[var(--color-bg-muted)] text-[var(--color-text-secondary)]">
                    <Icon className="h-4.5 w-4.5" />
                  </div>
                  <div>
                    <p className="text-sm font-semibold text-[var(--color-text-primary)]">{config.name}</p>
                    <p className="mt-1 text-xs leading-5 text-[var(--color-text-muted)]">{config.description}</p>
                  </div>
                </div>
                {isActive && (
                  <div className="inline-flex h-6 w-6 items-center justify-center rounded-full bg-[var(--color-accent)] text-[var(--color-text-inverse)] shadow-sm">
                    <Check className="h-3.5 w-3.5" />
                  </div>
                )}
              </div>

              <div
                className="mt-4 overflow-hidden rounded-xl border border-black/5 p-3"
                style={{ background: preview.shell }}
              >
                <div
                  className="overflow-hidden rounded-lg border border-black/5"
                  style={{ backgroundColor: preview.surface }}
                >
                  <div className="flex items-center gap-1.5 border-b border-black/5 px-3 py-2">
                    <span className="h-2 w-2 rounded-full bg-black/10" />
                    <span className="h-2 w-2 rounded-full bg-black/10" />
                    <span className="h-2 w-2 rounded-full bg-black/10" />
                    <div className="ml-2 h-2 w-24 rounded-full bg-black/10" />
                  </div>
                  <div className="grid grid-cols-[60px_1fr] gap-2 p-3">
                    <div className="space-y-2">
                      <div className="h-12 rounded-lg bg-black/5" />
                      <div className="h-2 rounded-full bg-black/8" />
                      <div className="h-2 rounded-full bg-black/8" />
                    </div>
                    <div className="space-y-2">
                      <div className="h-8 rounded-lg" style={{ backgroundColor: preview.accent, opacity: 0.12 }} />
                      <div className="grid grid-cols-2 gap-2">
                        <div className="h-12 rounded-lg bg-black/5" />
                        <div className="h-12 rounded-lg bg-black/5" />
                      </div>
                      <div className="h-2 w-20 rounded-full" style={{ backgroundColor: preview.text, opacity: 0.16 }} />
                    </div>
                  </div>
                </div>
              </div>
            </button>
          )
        })}
      </div>
    </div>
  )
}

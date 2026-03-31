export type ThemeType = 'light' | 'dark' | 'system'

export interface ThemeConfig {
  id: ThemeType
  name: string
  description: string
}

export const themeConfigs: ThemeConfig[] = [
  { id: 'system', name: '跟随设备', description: '自动跟随系统的浅色或深色外观' },
  { id: 'light', name: '浅色模式', description: '适合白天和明亮环境的清爽界面' },
  { id: 'dark', name: '深色模式', description: '适合夜间和低亮度环境的沉浸界面' },
]

export const DEFAULT_THEME: ThemeType = 'system'

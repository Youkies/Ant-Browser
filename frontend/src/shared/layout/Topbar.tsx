import { useState, useRef, useEffect } from 'react'
import type { CSSProperties } from 'react'
import { Bell, Search, User, Settings, Check, Trash2, Info, AlertCircle, CheckCircle, Minus, Square, Copy, X } from 'lucide-react'
import { Link, useLocation } from 'react-router-dom'
import clsx from 'clsx'
import { useNotificationStore, type Notification } from '../../store/notificationStore'
import { EventsEmit, WindowIsMaximised, WindowMinimise, WindowToggleMaximise } from '../../wailsjs/runtime/runtime'
import { navigationConfig } from '../../config'

const dragRegionStyle = {
  ['--wails-draggable' as string]: 'drag',
} as CSSProperties

const noDragRegionStyle = {
  ['--wails-draggable' as string]: 'no-drag',
} as CSSProperties

const dynamicPageTitles: Array<{ match: (pathname: string) => boolean; title: string }> = [
  { match: (pathname) => pathname.startsWith('/browser/detail/'), title: '实例详情' },
  { match: (pathname) => pathname.startsWith('/browser/edit/'), title: '编辑实例' },
  { match: (pathname) => pathname.startsWith('/browser/copy/'), title: '复制实例' },
]

const pageTitleMap = new Map(
  navigationConfig.flatMap((section) => section.items.map((item) => [item.path, item.name] as const))
)

function resolvePageTitle(pathname: string): string {
  if (pageTitleMap.has(pathname)) {
    return pageTitleMap.get(pathname) || 'Youkies Browser'
  }

  const matchedDynamicTitle = dynamicPageTitles.find((item) => item.match(pathname))
  if (matchedDynamicTitle) {
    return matchedDynamicTitle.title
  }

  return 'Youkies Browser'
}

function NotificationDropdown({
  notifications,
  onMarkAsRead,
  onMarkAllAsRead,
  onClear
}: {
  notifications: Notification[]
  onMarkAsRead: (id: string) => void
  onMarkAllAsRead: () => void
  onClear: () => void
}) {
  const unreadCount = notifications.filter(n => !n.read).length

  const getIcon = (type: Notification['type']) => {
    switch (type) {
      case 'success': return <CheckCircle className="w-4 h-4 text-[var(--color-success)]" />
      case 'warning': return <AlertCircle className="w-4 h-4 text-[var(--color-warning)]" />
      case 'error': return <AlertCircle className="w-4 h-4 text-[var(--color-error)]" />
      default: return <Info className="w-4 h-4 text-[var(--color-accent)]" />
    }
  }

  return (
    <div className="absolute right-0 top-full mt-2 w-80 bg-[var(--color-bg-surface)] border border-[var(--color-border-default)] rounded-xl shadow-xl overflow-hidden z-50 animate-fade-in">
      <div className="px-4 py-3 border-b border-[var(--color-border-muted)] flex items-center justify-between">
        <div className="flex items-center gap-2">
          <span className="text-sm font-semibold text-[var(--color-text-primary)]">通知</span>
          {unreadCount > 0 && (
            <span className="px-1.5 py-0.5 text-xs font-medium bg-[var(--color-accent)] text-white rounded-full">
              {unreadCount}
            </span>
          )}
        </div>
        <div className="flex items-center gap-1">
          {unreadCount > 0 && (
            <button
              onClick={onMarkAllAsRead}
              className="p-1.5 text-xs text-[var(--color-text-muted)] hover:text-[var(--color-accent)] hover:bg-[var(--color-bg-muted)] rounded transition-colors"
              title="全部标为已读"
            >
              <Check className="w-3.5 h-3.5" />
            </button>
          )}
          <button
            onClick={onClear}
            className="p-1.5 text-xs text-[var(--color-text-muted)] hover:text-[var(--color-error)] hover:bg-[var(--color-bg-muted)] rounded transition-colors"
            title="清空通知"
          >
            <Trash2 className="w-3.5 h-3.5" />
          </button>
        </div>
      </div>

      <div className="max-h-80 overflow-y-auto">
        {notifications.length === 0 ? (
          <div className="py-8 text-center text-[var(--color-text-muted)]">
            <Bell className="w-8 h-8 mx-auto mb-2 opacity-50" />
            <p className="text-sm">暂无通知</p>
          </div>
        ) : (
          notifications.map((notification) => (
            <div
              key={notification.id}
              onClick={() => onMarkAsRead(notification.id)}
              className={clsx(
                'px-4 py-3 border-b border-[var(--color-border-muted)] last:border-0 cursor-pointer transition-colors hover:bg-[var(--color-bg-muted)]',
                !notification.read && 'bg-[var(--color-accent)]/5'
              )}
            >
              <div className="flex gap-3">
                <div className="shrink-0 mt-0.5">
                  {getIcon(notification.type)}
                </div>
                <div className="flex-1 min-w-0">
                  <div className="flex items-start justify-between gap-2">
                    <p className={clsx(
                      'text-sm truncate',
                      notification.read ? 'text-[var(--color-text-secondary)]' : 'text-[var(--color-text-primary)] font-medium'
                    )}>
                      {notification.title}
                    </p>
                    {!notification.read && (
                      <span className="w-2 h-2 rounded-full bg-[var(--color-accent)] shrink-0 mt-1.5" />
                    )}
                  </div>
                  <p className="text-xs text-[var(--color-text-muted)] mt-0.5 line-clamp-2">
                    {notification.message}
                  </p>
                  <p className="text-[10px] text-[var(--color-text-muted)] mt-1">
                    {notification.time}
                  </p>
                </div>
              </div>
            </div>
          ))
        )}
      </div>

      {notifications.length > 0 && (
        <div className="px-4 py-2 border-t border-[var(--color-border-muted)] bg-[var(--color-bg-muted)]/50">
          <button className="w-full text-xs text-center text-[var(--color-accent)] hover:underline">
            查看全部通知
          </button>
        </div>
      )}
    </div>
  )
}

export function Topbar() {
  const [showNotifications, setShowNotifications] = useState(false)
  const [isMaximised, setIsMaximised] = useState(false)
  const { notifications, markAsRead, markAllAsRead, clearNotifications } = useNotificationStore()
  const dropdownRef = useRef<HTMLDivElement>(null)
  const location = useLocation()

  const unreadCount = notifications.filter(n => !n.read).length
  const pageTitle = resolvePageTitle(location.pathname)

  const syncWindowState = async () => {
    try {
      setIsMaximised(await WindowIsMaximised())
    } catch {
      setIsMaximised(false)
    }
  }

  useEffect(() => {
    function handleClickOutside(event: MouseEvent) {
      if (dropdownRef.current && !dropdownRef.current.contains(event.target as Node)) {
        setShowNotifications(false)
      }
    }
    document.addEventListener('mousedown', handleClickOutside)
    return () => document.removeEventListener('mousedown', handleClickOutside)
  }, [])

  useEffect(() => {
    void syncWindowState()

    const handleResize = () => {
      void syncWindowState()
    }

    window.addEventListener('resize', handleResize)
    window.addEventListener('focus', handleResize)
    return () => {
      window.removeEventListener('resize', handleResize)
      window.removeEventListener('focus', handleResize)
    }
  }, [])

  const handleToggleMaximise = () => {
    WindowToggleMaximise()
    window.setTimeout(() => {
      void syncWindowState()
    }, 120)
  }

  const handleRequestClose = () => {
    EventsEmit('app:request-close')
  }

  const handleOpenQuickLaunch = () => {
    window.dispatchEvent(new CustomEvent('ant:quick-launch:open'))
  }

  return (
    <header
      className="h-14 bg-[var(--color-bg-surface)] border-b border-[var(--color-border-default)] pl-4 pr-2 flex items-center justify-between gap-3 select-none"
      style={dragRegionStyle}
      onDoubleClick={handleToggleMaximise}
    >
      <div className="flex items-center gap-4 min-w-0 flex-1">
        <div
          className="hidden md:flex items-center min-w-[96px] h-full"
        >
          <span className="text-sm font-semibold text-[var(--color-text-secondary)]">
            {pageTitle}
          </span>
        </div>

        <div className="w-56 lg:w-72" style={noDragRegionStyle}>
          <button
            type="button"
            onClick={handleOpenQuickLaunch}
            className="relative w-full h-9 pl-9 pr-14 overflow-hidden bg-[var(--color-bg-muted)] border border-transparent rounded-xl text-sm text-left text-[var(--color-text-primary)] hover:bg-[var(--color-bg-subtle)] hover:border-[var(--color-border-default)] focus:outline-none focus:bg-[var(--color-bg-surface)] focus:border-[var(--color-border-strong)] transition-all duration-150"
            title="打开全局快速启动"
          >
            <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-[var(--color-text-muted)]" />
            <span className="block truncate whitespace-nowrap text-[13px] text-[var(--color-text-muted)]">快速启动 / 搜索实例、Code、标签</span>
            <span className="absolute right-3 top-1/2 -translate-y-1/2 rounded-md border border-[var(--color-border-default)] bg-[var(--color-bg-surface)] px-1.5 py-0.5 text-[10px] font-medium text-[var(--color-text-muted)]">
              Ctrl K
            </span>
          </button>
        </div>

        <div
          className="flex-1 min-w-[48px] h-full"
        />
      </div>

      <div className="flex items-center gap-1" style={noDragRegionStyle}>
        <div className="relative" ref={dropdownRef}>
          <button
            onClick={() => setShowNotifications(!showNotifications)}
            className={clsx(
              'relative w-8 h-8 flex items-center justify-center rounded-md transition-colors duration-150',
              showNotifications
                ? 'text-[var(--color-accent)] bg-[var(--color-accent-muted)]'
                : 'text-[var(--color-text-muted)] hover:text-[var(--color-text-secondary)] hover:bg-[var(--color-accent-muted)]'
            )}
            title="通知"
          >
            <Bell className="w-4 h-4" />
            {unreadCount > 0 && (
              <span className="absolute -top-0.5 -right-0.5 w-4 h-4 text-[10px] font-medium bg-[var(--color-error)] text-white rounded-full flex items-center justify-center">
                {unreadCount > 9 ? '9+' : unreadCount}
              </span>
            )}
          </button>

          {showNotifications && (
            <NotificationDropdown
              notifications={notifications}
              onMarkAsRead={markAsRead}
              onMarkAllAsRead={markAllAsRead}
              onClear={() => {
                clearNotifications()
                setShowNotifications(false)
              }}
            />
          )}
        </div>

        <Link
          to="/settings"
          className="w-8 h-8 flex items-center justify-center text-[var(--color-text-muted)] hover:text-[var(--color-text-secondary)] hover:bg-[var(--color-accent-muted)] rounded-md transition-colors duration-150"
          title="设置"
        >
          <Settings className="w-4 h-4" />
        </Link>

        <div className="w-px h-5 bg-[var(--color-border-default)] mx-1.5" />

        <Link
          to="/profile"
          className="flex items-center gap-2 pl-1 pr-2.5 py-1 rounded-md hover:bg-[var(--color-accent-muted)] transition-colors duration-150"
        >
          <div className="w-7 h-7 bg-[var(--color-accent)] rounded-md flex items-center justify-center">
            <User className="w-3.5 h-3.5 text-[var(--color-text-inverse)]" />
          </div>
          <span className="text-sm font-medium text-[var(--color-text-secondary)]">Admin</span>
        </Link>

        <div className="w-px h-5 bg-[var(--color-border-default)] mx-1.5" />

        <div className="flex items-center rounded-xl border border-[var(--color-border-default)] bg-[var(--color-bg-base)]/55 overflow-hidden">
          <button
            type="button"
            onClick={() => WindowMinimise()}
            className="w-10 h-9 flex items-center justify-center text-[var(--color-text-muted)] hover:text-[var(--color-text-primary)] hover:bg-[var(--color-accent-muted)] transition-colors"
            title="最小化"
          >
            <Minus className="w-4 h-4" />
          </button>
          <button
            type="button"
            onClick={handleToggleMaximise}
            className="w-10 h-9 flex items-center justify-center text-[var(--color-text-muted)] hover:text-[var(--color-text-primary)] hover:bg-[var(--color-accent-muted)] transition-colors border-l border-[var(--color-border-default)]"
            title={isMaximised ? '还原' : '最大化'}
          >
            {isMaximised ? <Copy className="w-3.5 h-3.5" /> : <Square className="w-3.5 h-3.5" />}
          </button>
          <button
            type="button"
            onClick={handleRequestClose}
            className="w-10 h-9 flex items-center justify-center text-[var(--color-text-muted)] hover:text-white hover:bg-[#ef4444] transition-colors border-l border-[var(--color-border-default)]"
            title="关闭"
          >
            <X className="w-4 h-4" />
          </button>
        </div>
      </div>
    </header>
  )
}

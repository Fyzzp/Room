import { useState } from 'react'
import { Link, Outlet, useLocation, useNavigate } from 'react-router-dom'
import { cn } from '@/lib/utils'
import {
  LayoutDashboard, Server, Activity, Link2, LogOut,
  ChevronLeft, ChevronRight, Menu, Globe, Shield,
} from 'lucide-react'

const navItems = [
  { to: '/', icon: LayoutDashboard, label: '仪表盘', desc: '系统概览' },
  { to: '/servers', icon: Server, label: '服务器', desc: '节点管理' },
  { to: '/traffic', icon: Activity, label: '流量统计', desc: '数据监控' },
  { to: '/subscription', icon: Link2, label: '订阅管理', desc: '用户订阅' },
]

export function Layout() {
  const location = useLocation()
  const navigate = useNavigate()
  const [collapsed, setCollapsed] = useState(false)
  const user = JSON.parse(localStorage.getItem('user') || '{}')

  const handleLogout = () => {
    localStorage.removeItem('token')
    localStorage.removeItem('user')
    navigate('/login')
  }

  return (
    <div className="flex h-screen bg-gray-50">
      {/* 侧边栏 */}
      <aside className={cn(
        'flex flex-col bg-gradient-to-b from-gray-900 via-gray-900 to-gray-800 border-r border-gray-800 transition-all duration-300',
        collapsed ? 'w-[72px]' : 'w-64'
      )}>
        {/* Logo */}
        <div className="h-16 flex items-center px-5 border-b border-gray-800">
          <div className="flex items-center gap-3">
            <div className="w-9 h-9 bg-gradient-to-br from-blue-500 to-blue-700 rounded-xl flex items-center justify-center flex-shrink-0">
              <Globe className="w-5 h-5 text-white" />
            </div>
            {!collapsed && <span className="text-lg font-bold text-white tracking-tight">Room</span>}
          </div>
        </div>

        {/* 导航 */}
        <nav className="flex-1 py-4 px-3 space-y-1 overflow-y-auto">
          {navItems.map(({ to, icon: Icon, label, desc }) => {
            const active = location.pathname === to || (to !== '/' && location.pathname.startsWith(to))
            return (
              <Link
                key={to}
                to={to}
                className={cn(
                  'flex items-center gap-3 px-3 py-2.5 rounded-xl text-sm transition-all group relative',
                  active
                    ? 'bg-blue-600/20 text-blue-400 shadow-sm'
                    : 'text-gray-400 hover:text-gray-200 hover:bg-gray-800/50'
                )}
              >
                <div className={cn(
                  'w-8 h-8 rounded-lg flex items-center justify-center flex-shrink-0 transition-colors',
                  active ? 'bg-blue-600/30' : 'bg-gray-800 group-hover:bg-gray-700/50'
                )}>
                  <Icon className={cn('w-4.5 h-4.5', active ? 'text-blue-400' : 'text-gray-500 group-hover:text-gray-300')} />
                </div>
                {!collapsed && (
                  <div className="flex-1 min-w-0">
                    <div className="text-sm font-medium">{label}</div>
                    <div className="text-xs text-gray-600">{desc}</div>
                  </div>
                )}
              </Link>
            )
          })}
        </nav>

        {/* 底部 */}
        <div className="border-t border-gray-800 p-3 space-y-1">
          {!collapsed && (
            <div className="px-3 py-2 mb-2">
              <div className="text-xs text-gray-600 mb-1">登录身份</div>
              <div className="flex items-center gap-2">
                <div className="w-7 h-7 bg-gradient-to-br from-emerald-400 to-emerald-600 rounded-lg flex items-center justify-center text-white text-xs font-bold">
                  {user.email?.[0]?.toUpperCase() || 'A'}
                </div>
                <div className="text-sm text-gray-300 truncate">{user.email || 'Admin'}</div>
              </div>
            </div>
          )}
          <button
            onClick={handleLogout}
            className="flex items-center gap-3 px-3 py-2.5 w-full rounded-xl text-sm text-gray-500 hover:text-red-400 hover:bg-red-500/10 transition-colors"
          >
            <LogOut className="w-5 h-5" />
            {!collapsed && '退出登录'}
          </button>
          <button
            onClick={() => setCollapsed(!collapsed)}
            className="flex items-center gap-3 px-3 py-2 w-full rounded-xl text-sm text-gray-600 hover:text-gray-300 transition-colors"
          >
            {collapsed ? <ChevronRight className="w-4 h-4 mx-auto" /> : <ChevronLeft className="w-4 h-4" />}
            {!collapsed && '收起菜单'}
          </button>
        </div>
      </aside>

      {/* 主内容 */}
      <main className="flex-1 overflow-auto">
        {/* 顶部栏 */}
        <header className="h-16 bg-white border-b border-gray-200 flex items-center justify-between px-6 sticky top-0 z-10">
          <div className="flex items-center gap-3">
            <Menu className="w-5 h-5 text-gray-400 lg:hidden" />
            <div>
              <h2 className="text-sm font-medium text-gray-700">
                {navItems.find(i => location.pathname === i.to || (i.to !== '/' && location.pathname.startsWith(i.to)))?.label || '仪表盘'}
              </h2>
            </div>
          </div>
          <div className="flex items-center gap-4">
            <div className="flex items-center gap-2 px-3 py-1.5 bg-green-50 border border-green-200 rounded-lg">
              <div className="w-2 h-2 bg-green-500 rounded-full animate-pulse" />
              <span className="text-xs text-green-700 font-medium">系统运行中</span>
            </div>
            <div className="w-8 h-8 bg-gray-100 rounded-lg flex items-center justify-center">
              <Shield className="w-4 h-4 text-gray-500" />
            </div>
          </div>
        </header>

        <div className="p-6">
          <Outlet />
        </div>
      </main>
    </div>
  )
}

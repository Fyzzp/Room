import { Server, Users, Activity, TrendingUp, ArrowUpRight, Cpu } from 'lucide-react'
import { Link } from 'react-router-dom'
import { useState } from 'react'

interface Stats {
  servers: number
  users: number
  traffic: string
  bandwidth: string
  cpu: string
  disk: string
}

export function DashboardPage() {
  const [stats] = useState<Stats>({
    servers: 0, users: 0, traffic: '0 B', bandwidth: '0 B/s', cpu: '0%', disk: '0%'
  })

  return (
    <div className="space-y-6">
      {/* 欢迎横幅 */}
      <div className="bg-gradient-to-r from-blue-600 via-blue-700 to-indigo-700 rounded-2xl p-8 text-white shadow-lg shadow-blue-500/20">
        <div className="flex items-center justify-between">
          <div>
            <p className="text-blue-200 text-sm mb-1">欢迎回来</p>
            <h2 className="text-2xl font-bold">系统仪表盘</h2>
            <p className="text-blue-300/80 text-sm mt-2">实时监控您的 Xray 服务器集群</p>
          </div>
          <div className="hidden md:flex gap-3">
            <Link to="/servers" className="px-4 py-2 bg-white/20 hover:bg-white/30 rounded-xl text-sm font-medium backdrop-blur transition-colors">
              管理服务器
            </Link>
            <Link to="/traffic" className="px-4 py-2 bg-white/10 hover:bg-white/20 rounded-xl text-sm font-medium backdrop-blur transition-colors">
              查看流量
            </Link>
          </div>
        </div>
      </div>

      {/* 统计卡片 */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
        {[
          { label: '在线服务器', value: stats.servers, icon: Server, color: 'from-emerald-500 to-emerald-600', bg: 'bg-emerald-50', text: 'text-emerald-700' },
          { label: '活跃用户', value: stats.users, icon: Users, color: 'from-blue-500 to-blue-600', bg: 'bg-blue-50', text: 'text-blue-700' },
          { label: '今日流量', value: stats.traffic, icon: Activity, color: 'from-amber-500 to-amber-600', bg: 'bg-amber-50', text: 'text-amber-700' },
          { label: '当前带宽', value: stats.bandwidth, icon: TrendingUp, color: 'from-purple-500 to-purple-600', bg: 'bg-purple-50', text: 'text-purple-700' },
        ].map(({ label, value, icon: Icon, color }) => (
          <div key={label} className="bg-white rounded-2xl p-5 border border-gray-100 shadow-sm hover:shadow-md transition-shadow">
            <div className="flex items-center justify-between mb-4">
              <span className="text-sm text-gray-500">{label}</span>
              <div className={`w-10 h-10 rounded-xl bg-gradient-to-br ${color} flex items-center justify-center`}>
                <Icon className="w-5 h-5 text-white" />
              </div>
            </div>
            <div className="text-2xl font-bold text-gray-800">{value}</div>
            <div className="flex items-center gap-1 mt-2 text-xs text-green-600">
              <ArrowUpRight className="w-3 h-3" /> 正常
            </div>
          </div>
        ))}
      </div>

      {/* 系统和流量图表区域 */}
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-4">
        {/* 系统资源 */}
        <div className="bg-white rounded-2xl border border-gray-100 shadow-sm p-6">
          <h3 className="text-base font-semibold text-gray-800 mb-4 flex items-center gap-2">
            <Cpu className="w-5 h-5 text-blue-500" /> 系统资源
          </h3>
          <div className="space-y-4">
            {[
              { label: 'CPU', value: stats.cpu, percent: 15, color: 'bg-blue-500' },
              { label: '内存', value: '1.2GB / 4GB', percent: 30, color: 'bg-purple-500' },
              { label: '磁盘', value: stats.disk, percent: 45, color: 'bg-amber-500' },
            ].map(({ label, value, percent, color }) => (
              <div key={label}>
                <div className="flex justify-between text-sm mb-1.5">
                  <span className="text-gray-600">{label}</span>
                  <span className="text-gray-800 font-medium">{value}</span>
                </div>
                <div className="h-2 bg-gray-100 rounded-full overflow-hidden">
                  <div className={`h-full ${color} rounded-full transition-all`} style={{ width: `${percent}%` }} />
                </div>
              </div>
            ))}
          </div>
        </div>

        {/* 流量趋势 */}
        <div className="bg-white rounded-2xl border border-gray-100 shadow-sm p-6">
          <h3 className="text-base font-semibold text-gray-800 mb-4 flex items-center gap-2">
            <Activity className="w-5 h-5 text-green-500" /> 流量趋势 (24h)
          </h3>
          <div className="flex items-end justify-between h-40 gap-1 px-2">
            {[15, 22, 18, 35, 28, 45, 38, 55, 42, 60, 48, 52, 40, 58, 50, 62, 45, 55, 38, 48, 52, 45, 55, 42].map((h, i) => (
              <div key={i} className="flex-1 bg-gradient-to-t from-blue-500/30 to-blue-500/80 rounded-t" style={{ height: `${h}%` }} />
            ))}
          </div>
          <div className="flex justify-between mt-3 text-xs text-gray-400">
            <span>00:00</span><span>06:00</span><span>12:00</span><span>18:00</span><span>23:59</span>
          </div>
        </div>
      </div>

      {/* 快速操作 */}
      <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
        {[
          { title: '添加服务器', desc: '部署新的 Xray 节点', to: '/servers', color: 'bg-blue-50 border-blue-200 hover:bg-blue-100' },
          { title: '流量统计', desc: '查看详细流量报告', to: '/traffic', color: 'bg-green-50 border-green-200 hover:bg-green-100' },
          { title: '订阅管理', desc: '管理用户订阅链接', to: '/subscription', color: 'bg-purple-50 border-purple-200 hover:bg-purple-100' },
        ].map(({ title, desc, to, color }) => (
          <Link key={title} to={to} className={`rounded-2xl border p-5 transition-colors ${color}`}>
            <h4 className="font-semibold text-gray-800 mb-1">{title}</h4>
            <p className="text-sm text-gray-500">{desc}</p>
          </Link>
        ))}
      </div>
    </div>
  )
}

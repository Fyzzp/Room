import { Server, Users, Activity, TrendingUp } from 'lucide-react'
import { Link } from 'react-router-dom'

export function DashboardPage() {
  const stats = [
    { label: '在线服务器', value: '0', icon: Server, color: 'text-green-500' },
    { label: '活跃用户', value: '0', icon: Users, color: 'text-blue-500' },
    { label: '今日流量', value: '0 B', icon: Activity, color: 'text-yellow-500' },
    { label: '总带宽', value: '0 B/s', icon: TrendingUp, color: 'text-purple-500' },
  ]

  return (
    <div>
      <h2 className="text-2xl font-bold mb-6">仪表盘</h2>

      {/* 统计卡片 */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4 mb-8">
        {stats.map(({ label, value, icon: Icon, color }) => (
          <div key={label} className="bg-white dark:bg-gray-900 rounded-xl p-5 border border-gray-200 dark:border-gray-800">
            <div className="flex items-center justify-between mb-3">
              <span className="text-sm text-gray-500 dark:text-gray-400">{label}</span>
              <Icon className={`w-5 h-5 ${color}`} />
            </div>
            <div className="text-2xl font-semibold">{value}</div>
          </div>
        ))}
      </div>

      {/* 快速操作 */}
      <div className="bg-white dark:bg-gray-900 rounded-xl p-6 border border-gray-200 dark:border-gray-800">
        <h3 className="text-lg font-semibold mb-4">快速操作</h3>
        <div className="flex gap-3">
          <Link
            to="/servers"
            className="px-4 py-2 bg-blue-600 text-white rounded-lg text-sm hover:bg-blue-700 transition-colors"
          >
            添加服务器
          </Link>
          <Link
            to="/subscription"
            className="px-4 py-2 border border-gray-300 dark:border-gray-700 rounded-lg text-sm hover:bg-gray-50 dark:hover:bg-gray-800 transition-colors"
          >
            管理订阅
          </Link>
        </div>
      </div>
    </div>
  )
}

import { Link2, QrCode, Users } from 'lucide-react'

export function SubscriptionPage() {
  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h2 className="text-2xl font-bold text-gray-800">订阅管理</h2>
          <p className="text-gray-500 mt-1">管理用户订阅链接和节点分配</p>
        </div>
      </div>

      <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
        {[
          { label: '活跃订阅', value: '0', icon: Link2, color: 'from-blue-500 to-blue-600' },
          { label: '总节点', value: '0', icon: QrCode, color: 'from-emerald-500 to-emerald-600' },
          { label: '订阅用户', value: '0', icon: Users, color: 'from-purple-500 to-purple-600' },
        ].map(({ label, value, icon: Icon, color }) => (
          <div key={label} className="bg-white rounded-2xl border border-gray-100 p-6 shadow-sm">
            <div className="flex items-center justify-between mb-4">
              <span className="text-sm text-gray-500">{label}</span>
              <div className={`w-10 h-10 rounded-xl bg-gradient-to-br ${color} flex items-center justify-center`}>
                <Icon className="w-5 h-5 text-white" />
              </div>
            </div>
            <div className="text-2xl font-bold text-gray-800">{value}</div>
          </div>
        ))}
      </div>

      <div className="bg-white rounded-2xl border border-gray-200 border-dashed p-12 text-center">
        <Link2 className="w-12 h-12 text-gray-300 mx-auto mb-4" />
        <p className="text-gray-500">暂无订阅，添加服务器后可创建订阅链接</p>
      </div>
    </div>
  )
}

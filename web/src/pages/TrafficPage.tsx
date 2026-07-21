import { Activity, Download, Upload, BarChart3 } from 'lucide-react'

export function TrafficPage() {
  return (
    <div className="space-y-6">
      <div>
        <h2 className="text-2xl font-bold text-gray-800">流量统计</h2>
        <p className="text-gray-500 mt-1">实时监控各服务器和用户的流量使用情况</p>
      </div>

      {/* 总览卡片 */}
      <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
        {[
          { label: '今日上行', value: '0 B', icon: Upload, color: 'from-blue-500 to-blue-600' },
          { label: '今日下行', value: '0 B', icon: Download, color: 'from-emerald-500 to-emerald-600' },
          { label: '总流量', value: '0 B', icon: Activity, color: 'from-purple-500 to-purple-600' },
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

      {/* 图表占位 */}
      <div className="bg-white rounded-2xl border border-gray-100 shadow-sm p-6">
        <div className="flex items-center gap-2 mb-4">
          <BarChart3 className="w-5 h-5 text-blue-500" />
          <h3 className="font-semibold text-gray-800">30天流量趋势</h3>
        </div>
        <div className="flex items-end justify-between h-48 gap-1 px-2">
          {[25, 30, 22, 45, 38, 55, 42, 60, 35, 48, 52, 40, 58, 50, 62, 45, 55, 38, 48, 52, 45, 55, 42, 35, 28, 40, 50, 45, 55, 42].map((h, i) => (
            <div key={i} className="flex-1 bg-gradient-to-t from-blue-100 to-blue-400 rounded-t" style={{ height: `${h}%` }} />
          ))}
        </div>
        <div className="flex justify-between mt-3 text-xs text-gray-400">
          <span>Day 1</span><span>Day 10</span><span>Day 20</span><span>Day 30</span>
        </div>
      </div>

      {/* 空状态 */}
      <div className="bg-white rounded-2xl border border-gray-200 border-dashed p-12 text-center">
        <Activity className="w-12 h-12 text-gray-300 mx-auto mb-4" />
        <p className="text-gray-500">连接服务器后流量数据将在此展示</p>
      </div>
    </div>
  )
}

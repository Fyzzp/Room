import { Link } from 'react-router-dom'
import { Server, Plus, Wifi, WifiOff, Settings, Trash2 } from 'lucide-react'
import { useState } from 'react'

const mockServers = [
  { id: 1, name: '东京-01', ip: '160.25.135.232', port: 23889, status: 'online', xray: 'external', traffic: '12.5 GB', uptime: '3天 12时' },
]

export function ServersPage() {
  const [servers] = useState(mockServers)

  return (
    <div className="space-y-6">
      {/* 头部 */}
      <div className="flex items-center justify-between">
        <div>
          <h2 className="text-2xl font-bold text-gray-800">服务器管理</h2>
          <p className="text-gray-500 mt-1">管理您的所有 Xray 远程节点</p>
        </div>
        <button className="flex items-center gap-2 px-4 py-2.5 bg-gradient-to-r from-blue-600 to-blue-700 text-white rounded-xl font-medium hover:from-blue-700 hover:to-blue-800 transition-all shadow-lg shadow-blue-500/25">
          <Plus className="w-4 h-4" /> 添加服务器
        </button>
      </div>

      {/* 统计摘要 */}
      <div className="grid grid-cols-3 gap-4">
        {[
          { label: '总数', value: servers.length, color: 'text-blue-600' },
          { label: '在线', value: servers.filter(s => s.status === 'online').length, color: 'text-green-600' },
          { label: '离线', value: servers.filter(s => s.status !== 'online').length, color: 'text-red-600' },
        ].map(({ label, value, color }) => (
          <div key={label} className="bg-white rounded-2xl border border-gray-100 p-4 text-center shadow-sm">
            <div className={`text-2xl font-bold ${color}`}>{value}</div>
            <div className="text-sm text-gray-500 mt-1">{label}</div>
          </div>
        ))}
      </div>

      {/* 服务器卡片列表 */}
      {servers.length === 0 ? (
        <div className="bg-white rounded-2xl border border-gray-200 border-dashed p-16 text-center">
          <Server className="w-12 h-12 text-gray-300 mx-auto mb-4" />
          <p className="text-gray-500 mb-4">还没有服务器，添加第一台开始吧</p>
          <button className="inline-flex items-center gap-2 px-4 py-2 bg-blue-600 text-white rounded-xl text-sm hover:bg-blue-700 transition-colors">
            <Plus className="w-4 h-4" /> 添加服务器
          </button>
        </div>
      ) : (
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
          {servers.map((server) => (
            <div key={server.id} className="bg-white rounded-2xl border border-gray-100 shadow-sm hover:shadow-md transition-all group">
              {/* 卡片头 */}
              <div className="p-5 pb-4">
                <div className="flex items-start justify-between mb-3">
                  <div className="flex items-center gap-3">
                    <div className={`w-12 h-12 rounded-2xl flex items-center justify-center ${
                      server.status === 'online' ? 'bg-gradient-to-br from-emerald-400 to-emerald-600' : 'bg-gradient-to-br from-gray-400 to-gray-500'
                    }`}>
                      <Server className="w-6 h-6 text-white" />
                    </div>
                    <div>
                      <h3 className="font-semibold text-gray-800">{server.name}</h3>
                      <p className="text-xs text-gray-400 font-mono">{server.ip}</p>
                    </div>
                  </div>
                  <div className="flex items-center gap-1.5">
                    {server.status === 'online' ? (
                      <span className="flex items-center gap-1 px-2 py-0.5 bg-green-50 text-green-700 rounded-lg text-xs font-medium">
                        <Wifi className="w-3 h-3" /> 在线
                      </span>
                    ) : (
                      <span className="flex items-center gap-1 px-2 py-0.5 bg-red-50 text-red-700 rounded-lg text-xs font-medium">
                        <WifiOff className="w-3 h-3" /> 离线
                      </span>
                    )}
                  </div>
                </div>

                {/* 信息行 */}
                <div className="grid grid-cols-2 gap-3 text-sm">
                  <div className="bg-gray-50 rounded-xl p-2.5">
                    <div className="text-gray-400 text-xs mb-0.5">Xray 模式</div>
                    <div className="text-gray-700 font-medium">{server.xray}</div>
                  </div>
                  <div className="bg-gray-50 rounded-xl p-2.5">
                    <div className="text-gray-400 text-xs mb-0.5">流量</div>
                    <div className="text-gray-700 font-medium">{server.traffic}</div>
                  </div>
                  <div className="bg-gray-50 rounded-xl p-2.5">
                    <div className="text-gray-400 text-xs mb-0.5">运行时间</div>
                    <div className="text-gray-700 font-medium">{server.uptime}</div>
                  </div>
                  <div className="bg-gray-50 rounded-xl p-2.5">
                    <div className="text-gray-400 text-xs mb-0.5">端口</div>
                    <div className="text-gray-700 font-mono font-medium text-xs">{server.port}</div>
                  </div>
                </div>
              </div>

              {/* 操作栏 */}
              <div className="px-5 py-3 border-t border-gray-100 flex items-center gap-2">
                <Link
                  to={`/servers/${server.id}/xray`}
                  className="flex-1 flex items-center justify-center gap-1.5 py-2 bg-blue-50 text-blue-700 rounded-xl text-sm font-medium hover:bg-blue-100 transition-colors"
                >
                  <Settings className="w-4 h-4" /> Xray 配置
                </Link>
                <button className="p-2 text-gray-400 hover:text-red-500 hover:bg-red-50 rounded-xl transition-colors">
                  <Trash2 className="w-4 h-4" />
                </button>
              </div>
            </div>
          ))}
        </div>
      )}
    </div>
  )
}

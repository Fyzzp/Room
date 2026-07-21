import { Server, Plus, Wifi, WifiOff, X, Trash2, Clipboard, Check, Terminal } from 'lucide-react'
import { useState, useEffect, useCallback } from 'react'

interface ServerItem {
  id: number; name: string; token: string; ip_address: string | null; listen_port: number
  connection_mode: string; xray_mode: string; status: string
  upload_speed: number; download_speed: number
}

export function ServersPage() {
  const [servers, setServers] = useState<ServerItem[]>([])
  const [showForm, setShowForm] = useState(false)
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState('')
  const [form, setForm] = useState({ name: '', connection_mode: 'websocket', xray_mode: 'external' })
  const [selectedServer, setSelectedServer] = useState<ServerItem | null>(null)
  const [copied, setCopied] = useState(false)

  const token = localStorage.getItem('token') || ''

  const fetchServers = useCallback(async () => {
    try {
      const res = await fetch('/api/servers', { headers: { Authorization: `Bearer ${token}` } })
      const data = await res.json()
      setServers(data.servers || [])
    } catch { /* ignore */ }
  }, [token])

  useEffect(() => { fetchServers() }, [fetchServers])

  const handleCreate = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!form.name) { setError('请输入服务器名称'); return }
    setLoading(true); setError('')
    try {
      const res = await fetch('/api/servers', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json', Authorization: `Bearer ${token}` },
        body: JSON.stringify(form),
      })
      if (!res.ok) { const d = await res.json(); throw new Error(d.error || '创建失败') }
      setShowForm(false); setForm({ name: '', connection_mode: 'websocket', xray_mode: 'external' })
      fetchServers()
    } catch (err: unknown) { setError(err instanceof Error ? err.message : '创建失败') }
    finally { setLoading(false) }
  }

  const handleDelete = async (id: number) => {
    if (!confirm('确定删除此服务器？相关配置将被清除。')) return
    try {
      await fetch(`/api/servers?id=${id}`, { method: 'DELETE', headers: { Authorization: `Bearer ${token}` } })
      fetchServers()
    } catch { /* ignore */ }
  }

  const getInstallCmd = (s: ServerItem) => {
    const host = window.location.host
    return `curl -fsSL https://raw.githubusercontent.com/Fyzzp/Room-Agent/main/scripts/install.sh | bash -s -- --master http://${host} --token ${s.token} --mode ${s.xray_mode}`
  }

  const copyCmd = (cmd: string) => { navigator.clipboard.writeText(cmd); setCopied(true); setTimeout(() => setCopied(false), 2000) }

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h2 className="text-2xl font-bold text-gray-800">服务器管理</h2>
          <p className="text-gray-500 mt-1">管理您的所有 Xray 远程节点</p>
        </div>
        <button onClick={() => setShowForm(true)} className="flex items-center gap-2 px-4 py-2.5 bg-gradient-to-r from-blue-600 to-blue-700 text-white rounded-xl font-medium hover:from-blue-700 hover:to-blue-800 transition-all shadow-lg shadow-blue-500/25">
          <Plus className="w-4 h-4" /> 添加服务器
        </button>
      </div>

      {/* 添加表单弹窗 */}
      {showForm && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/50 backdrop-blur-sm p-4" onClick={() => setShowForm(false)}>
          <div className="bg-white rounded-2xl shadow-2xl max-w-md w-full p-6" onClick={e => e.stopPropagation()}>
            <div className="flex items-center justify-between mb-4">
              <h3 className="text-lg font-bold text-gray-800">添加服务器</h3>
              <button onClick={() => setShowForm(false)} className="text-gray-400 hover:text-gray-600"><X className="w-5 h-5" /></button>
            </div>
            {error && <div className="mb-3 p-2 bg-red-50 text-red-600 rounded-lg text-sm">{error}</div>}
            <form onSubmit={handleCreate} className="space-y-4">
              <div>
                <label className="block text-sm font-medium text-gray-600 mb-1">服务器名称</label>
                <input value={form.name} onChange={e => setForm({...form, name: e.target.value})}
                  className="w-full px-3 py-2 border border-gray-200 rounded-xl focus:ring-2 focus:ring-blue-500 outline-none"
                  placeholder="例: 东京-01" required />
              </div>
              <div>
                <label className="block text-sm font-medium text-gray-600 mb-1">连接模式</label>
                <select value={form.connection_mode} onChange={e => setForm({...form, connection_mode: e.target.value})}
                  className="w-full px-3 py-2 border border-gray-200 rounded-xl focus:ring-2 focus:ring-blue-500 outline-none">
                  <option value="websocket">WebSocket (推荐)</option>
                  <option value="http">HTTP</option>
                  <option value="pull">Pull</option>
                </select>
              </div>
              <div>
                <label className="block text-sm font-medium text-gray-600 mb-1">Xray 模式</label>
                <select value={form.xray_mode} onChange={e => setForm({...form, xray_mode: e.target.value})}
                  className="w-full px-3 py-2 border border-gray-200 rounded-xl focus:ring-2 focus:ring-blue-500 outline-none">
                  <option value="external">External (独立安装)</option>
                  <option value="embedded">Embedded (内嵌)</option>
                </select>
              </div>
              <button type="submit" disabled={loading}
                className="w-full py-2.5 bg-blue-600 text-white rounded-xl font-medium hover:bg-blue-700 disabled:opacity-50">
                {loading ? '创建中...' : '创建服务器'}
              </button>
            </form>
          </div>
        </div>
      )}

      {/* 连接详情弹窗 */}
      {selectedServer && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/50 backdrop-blur-sm p-4" onClick={() => setSelectedServer(null)}>
          <div className="bg-white rounded-2xl shadow-2xl max-w-lg w-full p-6" onClick={e => e.stopPropagation()}>
            <div className="flex items-center justify-between mb-4">
              <h3 className="text-lg font-bold text-gray-800">{selectedServer.name} — 连接配置</h3>
              <button onClick={() => setSelectedServer(null)} className="text-gray-400 hover:text-gray-600"><X className="w-5 h-5" /></button>
            </div>
            <div className="space-y-3">
              <div className="bg-gray-50 rounded-xl p-3">
                <div className="text-xs text-gray-400 mb-1">服务器 Token</div>
                <div className="text-sm font-mono break-all text-gray-700">{selectedServer.token}</div>
              </div>
              <div className="bg-gray-50 rounded-xl p-3">
                <div className="text-xs text-gray-400 mb-1">连接模式</div>
                <div className="text-sm text-gray-700">{selectedServer.connection_mode} · {selectedServer.xray_mode}</div>
              </div>
              <div className="bg-gray-900 rounded-xl p-4">
                <div className="flex items-center justify-between mb-2">
                  <span className="text-xs text-gray-400 flex items-center gap-1"><Terminal className="w-3 h-3" /> Agent 一键安装</span>
                  <button onClick={() => copyCmd(getInstallCmd(selectedServer))} className="text-xs text-blue-400 hover:text-blue-300 flex items-center gap-1">
                    {copied ? <><Check className="w-3 h-3" />已复制</> : <><Clipboard className="w-3 h-3" />复制</>}
                  </button>
                </div>
                <pre className="text-xs text-green-400 overflow-x-auto whitespace-pre-wrap">{getInstallCmd(selectedServer)}</pre>
              </div>
            </div>
          </div>
        </div>
      )}

      {/* 统计 */}
      <div className="grid grid-cols-3 gap-4">
        {[
          { label: '总数', value: servers.length, color: 'text-blue-600' },
          { label: '在线', value: servers.filter(s => s.status === 'connected').length, color: 'text-green-600' },
          { label: '离线', value: servers.filter(s => s.status !== 'connected').length, color: 'text-red-600' },
        ].map(({ label, value, color }) => (
          <div key={label} className="bg-white rounded-2xl border border-gray-100 p-4 text-center shadow-sm">
            <div className={`text-2xl font-bold ${color}`}>{value}</div>
            <div className="text-sm text-gray-500 mt-1">{label}</div>
          </div>
        ))}
      </div>

      {/* 服务器列表 */}
      {servers.length === 0 ? (
        <div className="bg-white rounded-2xl border border-gray-200 border-dashed p-16 text-center">
          <Server className="w-12 h-12 text-gray-300 mx-auto mb-4" />
          <p className="text-gray-500 mb-4">还没有服务器，添加第一台开始吧</p>
          <button onClick={() => setShowForm(true)} className="inline-flex items-center gap-2 px-4 py-2 bg-blue-600 text-white rounded-xl text-sm hover:bg-blue-700 transition-colors">
            <Plus className="w-4 h-4" /> 添加服务器
          </button>
        </div>
      ) : (
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
          {servers.map((s) => (
            <div key={s.id} className="bg-white rounded-2xl border border-gray-100 shadow-sm hover:shadow-md transition-all group">
              <div className="p-5 pb-4">
                <div className="flex items-start justify-between mb-3">
                  <div className="flex items-center gap-3">
                    <div className={`w-12 h-12 rounded-2xl flex items-center justify-center ${s.status === 'connected' ? 'bg-gradient-to-br from-emerald-400 to-emerald-600' : 'bg-gradient-to-br from-gray-400 to-gray-500'}`}>
                      <Server className="w-6 h-6 text-white" />
                    </div>
                    <div>
                      <h3 className="font-semibold text-gray-800">{s.name}</h3>
                      <p className="text-xs text-gray-400 font-mono">{s.ip_address || '等待连接'}</p>
                    </div>
                  </div>
                  <span className={`flex items-center gap-1 px-2 py-0.5 rounded-lg text-xs font-medium ${s.status === 'connected' ? 'bg-green-50 text-green-700' : 'bg-gray-100 text-gray-500'}`}>
                    {s.status === 'connected' ? <><Wifi className="w-3 h-3" />在线</> : <><WifiOff className="w-3 h-3" />{s.status === 'pending' ? '待连接' : '离线'}</>}
                  </span>
                </div>
                <div className="grid grid-cols-2 gap-3 text-sm">
                  <div className="bg-gray-50 rounded-xl p-2.5"><div className="text-gray-400 text-xs mb-0.5">Xray</div><div className="text-gray-700 font-medium">{s.xray_mode}</div></div>
                  <div className="bg-gray-50 rounded-xl p-2.5"><div className="text-gray-400 text-xs mb-0.5">端口</div><div className="text-gray-700 font-mono font-medium text-xs">{s.listen_port}</div></div>
                </div>
              </div>
              <div className="px-5 py-3 border-t border-gray-100 flex items-center gap-2">
                <button onClick={() => setSelectedServer(s)} className="flex-1 flex items-center justify-center gap-1.5 py-2 bg-blue-50 text-blue-700 rounded-xl text-sm font-medium hover:bg-blue-100 transition-colors">
                  <Terminal className="w-4 h-4" /> 连接配置
                </button>
                <button onClick={() => handleDelete(s.id)} className="p-2 text-gray-400 hover:text-red-500 hover:bg-red-50 rounded-xl transition-colors" title="删除服务器">
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

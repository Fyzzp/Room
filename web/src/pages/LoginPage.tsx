import { useState } from 'react'
import { useNavigate, Link } from 'react-router-dom'
import { Eye, EyeOff, Server, Shield, Zap, Globe } from 'lucide-react'

function PrivacyModal({ show, onClose }: { show: boolean; onClose: () => void }) {
  if (!show) return null
  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/50 backdrop-blur-sm p-4" onClick={onClose}>
      <div className="bg-white rounded-2xl shadow-2xl max-w-lg w-full max-h-[80vh] overflow-y-auto p-6" onClick={e => e.stopPropagation()}>
        <div className="flex items-center justify-between mb-4">
          <h3 className="text-xl font-bold text-gray-800">隐私政策</h3>
          <button onClick={onClose} className="text-gray-400 hover:text-gray-600 text-2xl">&times;</button>
        </div>
        <div className="text-gray-600 text-sm leading-relaxed space-y-3">
          <p><strong>最后更新：2026年7月21日</strong></p>
          <h4 className="text-gray-800 font-semibold">1. 信息收集</h4>
          <p>Room 仅收集运行必需的信息：您的邮箱地址用于账户识别，服务器 IP 和端口信息用于节点管理。我们不会收集您的浏览历史、代理流量内容等个人隐私数据。</p>
          <h4 className="text-gray-800 font-semibold">2. 信息使用</h4>
          <p>收集的信息仅用于：提供面板管理服务、流量统计、服务器状态监控。未经您明确同意，我们不会将数据分享给第三方。</p>
          <h4 className="text-gray-800 font-semibold">3. 数据存储</h4>
          <p>所有数据存储在您自己的服务器上（PostgreSQL 数据库）。Room 不会将数据上传至任何云端服务。</p>
          <h4 className="text-gray-800 font-semibold">4. 数据安全</h4>
          <p>密码使用 bcrypt 加密存储，通信通过 HTTPS/TLS 保护。建议您启用防火墙并定期更新系统。</p>
          <h4 className="text-gray-800 font-semibold">5. 联系我们</h4>
          <p>如有隐私相关问题，请通过 GitHub Issues 联系我们。</p>
        </div>
      </div>
    </div>
  )
}

function TermsModal({ show, onClose }: { show: boolean; onClose: () => void }) {
  if (!show) return null
  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/50 backdrop-blur-sm p-4" onClick={onClose}>
      <div className="bg-white rounded-2xl shadow-2xl max-w-lg w-full max-h-[80vh] overflow-y-auto p-6" onClick={e => e.stopPropagation()}>
        <div className="flex items-center justify-between mb-4">
          <h3 className="text-xl font-bold text-gray-800">用户协议</h3>
          <button onClick={onClose} className="text-gray-400 hover:text-gray-600 text-2xl">&times;</button>
        </div>
        <div className="text-gray-600 text-sm leading-relaxed space-y-3">
          <p><strong>最后更新：2026年7月21日</strong></p>
          <h4 className="text-gray-800 font-semibold">1. 服务说明</h4>
          <p>Room 是一个开源的多机 Xray 管理面板。本软件按"现状"提供，不提供任何明示或暗示的保证。</p>
          <h4 className="text-gray-800 font-semibold">2. 用户责任</h4>
          <p>您同意：不使用本软件进行任何违法活动；遵守您所在地区的法律法规；自行负责服务器和数据的安全维护。</p>
          <h4 className="text-gray-800 font-semibold">3. 免责声明</h4>
          <p>Room 开发者不对因使用本软件而产生的任何直接或间接损失承担责任，包括但不限于数据丢失、服务中断、安全漏洞等。</p>
          <h4 className="text-gray-800 font-semibold">4. 开源许可</h4>
          <p>本软件基于 MIT 许可证开源。您可以自由使用、修改和分发，但需保留原始版权声明。</p>
        </div>
      </div>
    </div>
  )
}

export function LoginPage() {
  const [email, setEmail] = useState('')
  const [password, setPassword] = useState('')
  const [showPwd, setShowPwd] = useState(false)
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState('')
  const [showPrivacy, setShowPrivacy] = useState(false)
  const [showTerms, setShowTerms] = useState(false)
  const navigate = useNavigate()

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    setError(''); setLoading(true)
    try {
      const res = await fetch('/api/auth/login', {
        method: 'POST', headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ email, password }),
      })
      const data = await res.json()
      if (!res.ok) throw new Error(data.error || '登录失败')
      localStorage.setItem('token', data.token)
      localStorage.setItem('user', JSON.stringify(data.user))
      navigate('/')
    } catch (err: unknown) {
      setError(err instanceof Error ? err.message : '登录失败')
    } finally { setLoading(false) }
  }

  return (
    <div className="min-h-screen flex bg-gray-50">
      {/* 左侧品牌区 */}
      <div className="hidden lg:flex lg:w-1/2 bg-gradient-to-br from-gray-900 via-gray-800 to-gray-900 text-white flex-col justify-center p-16 relative overflow-hidden">
        <div className="absolute top-20 -right-20 w-80 h-80 bg-blue-500/10 rounded-full blur-3xl" />
        <div className="absolute bottom-20 -left-20 w-96 h-96 bg-indigo-500/10 rounded-full blur-3xl" />
        
        <div className="relative z-10 max-w-md">
          <div className="flex items-center gap-3 mb-8">
            <div className="w-12 h-12 bg-gradient-to-br from-blue-500 to-blue-600 rounded-2xl flex items-center justify-center">
              <Globe className="w-7 h-7 text-white" />
            </div>
            <h1 className="text-3xl font-bold tracking-tight">Room</h1>
          </div>
          
          <h2 className="text-2xl font-semibold mb-4">Xray 多机管理面板</h2>
          <p className="text-gray-400 text-lg mb-10 leading-relaxed">
            统一管理分布在全球的 Xray 服务器节点，实时流量监控，一键部署。
          </p>

          <div className="space-y-6">
            {[
              { icon: Server, title: '多服务器管理', desc: '一个面板掌控所有 Xray 节点' },
              { icon: Zap, title: '实时流量监控', desc: '精确的流量统计和带宽图表' },
              { icon: Shield, title: '安全可靠', desc: 'bcrypt 加密 + JWT 认证保护' },
            ].map(({ icon: Icon, title, desc }) => (
              <div key={title} className="flex items-start gap-4">
                <div className="w-10 h-10 bg-white/10 rounded-xl flex items-center justify-center flex-shrink-0">
                  <Icon className="w-5 h-5 text-blue-400" />
                </div>
                <div>
                  <div className="font-medium">{title}</div>
                  <div className="text-sm text-gray-500">{desc}</div>
                </div>
              </div>
            ))}
          </div>
        </div>

        <div className="absolute bottom-8 left-16 text-sm text-gray-600">
          © 2026 Room Panel · MIT License
        </div>
      </div>

      {/* 右侧登录框 */}
      <div className="flex-1 flex items-center justify-center p-8">
        <div className="w-full max-w-sm">
          {/* 移动端 Logo */}
          <div className="lg:hidden text-center mb-8">
            <div className="inline-flex items-center justify-center w-14 h-14 bg-gradient-to-br from-gray-800 to-gray-900 rounded-2xl mb-3">
              <Globe className="w-7 h-7 text-white" />
            </div>
            <h1 className="text-2xl font-bold text-gray-800">Room</h1>
          </div>

          <h2 className="text-xl font-semibold text-gray-800 mb-1">欢迎回来</h2>
          <p className="text-gray-500 text-sm mb-8">登录您的 Room 账户</p>

          {error && (
            <div className="mb-4 p-3 bg-red-50 border border-red-100 text-red-600 rounded-xl text-sm">{error}</div>
          )}

          <form onSubmit={handleSubmit} className="space-y-4">
            <div>
              <label className="block text-sm font-medium text-gray-600 mb-1.5">邮箱</label>
              <input type="email" value={email} onChange={(e) => setEmail(e.target.value)}
                className="w-full px-4 py-3 border border-gray-200 rounded-xl bg-white focus:ring-2 focus:ring-gray-400 focus:border-transparent outline-none transition-all"
                placeholder="admin@example.com" required />
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-600 mb-1.5">密码</label>
              <div className="relative">
                <input type={showPwd ? 'text' : 'password'} value={password} onChange={(e) => setPassword(e.target.value)}
                  className="w-full px-4 py-3 pr-12 border border-gray-200 rounded-xl bg-white focus:ring-2 focus:ring-gray-400 focus:border-transparent outline-none transition-all"
                  placeholder="••••••••" required />
                <button type="button" onClick={() => setShowPwd(!showPwd)} className="absolute right-3 top-1/2 -translate-y-1/2 text-gray-400 hover:text-gray-600">
                  {showPwd ? <EyeOff className="w-5 h-5" /> : <Eye className="w-5 h-5" />}
                </button>
              </div>
            </div>
            <button type="submit" disabled={loading}
              className="w-full py-3 bg-gray-900 text-white rounded-xl font-medium hover:bg-gray-800 transition-all disabled:opacity-50">
              {loading ? '登录中...' : '登录'}
            </button>
          </form>

          <div className="mt-6 text-center text-sm text-gray-500">
            还没有账户？{' '}
            <Link to="/register" className="text-gray-900 hover:text-gray-700 font-medium">立即注册</Link>
          </div>

          {/* 协议链接 */}
          <div className="mt-8 text-center text-xs text-gray-400 space-x-3">
            <button onClick={() => setShowPrivacy(true)} className="hover:text-gray-600 transition-colors">隐私政策</button>
            <span>·</span>
            <button onClick={() => setShowTerms(true)} className="hover:text-gray-600 transition-colors">用户协议</button>
          </div>
        </div>
      </div>

      <PrivacyModal show={showPrivacy} onClose={() => setShowPrivacy(false)} />
      <TermsModal show={showTerms} onClose={() => setShowTerms(false)} />
    </div>
  )
}

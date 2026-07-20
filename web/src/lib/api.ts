const API_BASE = '/api'

interface RequestOptions extends RequestInit {
  token?: string
}

export async function apiFetch<T = unknown>(
  endpoint: string,
  options: RequestOptions = {}
): Promise<T> {
  const { token, ...init } = options

  const headers: Record<string, string> = {
    'Content-Type': 'application/json',
    ...(init.headers as Record<string, string>),
  }

  if (token) {
    headers['Authorization'] = `Bearer ${token}`
  }

  const res = await fetch(`${API_BASE}${endpoint}`, {
    ...init,
    headers,
  })

  if (!res.ok) {
    const error = await res.json().catch(() => ({ message: res.statusText }))
    throw new Error(error.message || `HTTP ${res.status}`)
  }

  return res.json()
}

// API 方法
export const api = {
  // 认证
  login: (email: string, password: string) =>
    apiFetch<{ token: string; user: Record<string, unknown> }>('/auth/login', {
      method: 'POST',
      body: JSON.stringify({ email, password }),
    }),

  // 服务器
  getServers: () => apiFetch<{ servers: unknown[] }>('/servers'),

  // 入站
  getInbounds: (serverId: number) =>
    apiFetch<{ inbounds: unknown[] }>(`/inbounds?server_id=${serverId}`),

  // 流量统计
  getTraffic: (params: Record<string, string>) => {
    const qs = new URLSearchParams(params).toString()
    return apiFetch<{ records: unknown[] }>(`/traffic?${qs}`)
  },

  // 健康检查
  health: () => apiFetch<{ status: string; version: string }>('/health'),
}

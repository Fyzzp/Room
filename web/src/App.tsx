import { Routes, Route, Navigate } from 'react-router-dom'
import { Layout } from '@/components/layout/Layout'
import { DashboardPage } from '@/pages/DashboardPage'
import { ServersPage } from '@/pages/ServersPage'
import { XrayConfigPage } from '@/pages/XrayConfigPage'
import { TrafficPage } from '@/pages/TrafficPage'
import { SubscriptionPage } from '@/pages/SubscriptionPage'
import { LoginPage } from '@/pages/LoginPage'
import { RegisterPage } from '@/pages/RegisterPage'

function PrivateRoute({ children }: { children: React.ReactNode }) {
  const token = localStorage.getItem('token')
  if (!token) return <Navigate to="/login" replace />
  return <>{children}</>
}

export default function App() {
  return (
    <Routes>
      <Route path="/login" element={<LoginPage />} />
      <Route path="/register" element={<RegisterPage />} />
      <Route path="/" element={<PrivateRoute><Layout /></PrivateRoute>}>
        <Route index element={<DashboardPage />} />
        <Route path="servers" element={<ServersPage />} />
        <Route path="servers/:id/xray" element={<XrayConfigPage />} />
        <Route path="traffic" element={<TrafficPage />} />
        <Route path="subscription" element={<SubscriptionPage />} />
      </Route>
    </Routes>
  )
}

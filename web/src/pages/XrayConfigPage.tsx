export function XrayConfigPage() {
  return (
    <div>
      <h2 className="text-2xl font-bold mb-6">Xray 配置管理</h2>
      <div className="grid grid-cols-1 lg:grid-cols-3 gap-4 mb-6">
        <div className="bg-white dark:bg-gray-900 rounded-xl p-5 border border-gray-200 dark:border-gray-800">
          <h3 className="font-semibold mb-2">入站代理</h3>
          <p className="text-2xl font-bold text-blue-600">0</p>
          <p className="text-sm text-gray-500 mt-1">活跃入站</p>
        </div>
        <div className="bg-white dark:bg-gray-900 rounded-xl p-5 border border-gray-200 dark:border-gray-800">
          <h3 className="font-semibold mb-2">出站代理</h3>
          <p className="text-2xl font-bold text-green-600">0</p>
          <p className="text-sm text-gray-500 mt-1">活跃出站</p>
        </div>
        <div className="bg-white dark:bg-gray-900 rounded-xl p-5 border border-gray-200 dark:border-gray-800">
          <h3 className="font-semibold mb-2">路由规则</h3>
          <p className="text-2xl font-bold text-purple-600">0</p>
          <p className="text-sm text-gray-500 mt-1">活跃规则</p>
        </div>
      </div>

      {/* 入站列表 */}
      <div className="bg-white dark:bg-gray-900 rounded-xl border border-gray-200 dark:border-gray-800 p-6">
        <div className="flex items-center justify-between mb-4">
          <h3 className="font-semibold">入站列表</h3>
          <button className="text-sm text-blue-600 hover:text-blue-700">+ 添加入站</button>
        </div>
        <p className="text-gray-500 text-center py-8">暂无入站配置</p>
      </div>
    </div>
  )
}

export function ServersPage() {
  return (
    <div>
      <div className="flex items-center justify-between mb-6">
        <h2 className="text-2xl font-bold">服务器管理</h2>
        <button className="px-4 py-2 bg-blue-600 text-white rounded-lg text-sm hover:bg-blue-700">
          + 添加服务器
        </button>
      </div>
      <div className="bg-white dark:bg-gray-900 rounded-xl border border-gray-200 dark:border-gray-800 p-8 text-center">
        <p className="text-gray-500">暂无服务器，点击上方按钮添加第一台服务器</p>
      </div>
    </div>
  )
}

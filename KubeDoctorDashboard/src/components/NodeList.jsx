import StatusBadge from './StatusBadge'

function NodeList({ nodes, loading }) {
  return (
    <div className="bg-white shadow rounded-lg overflow-hidden">
      <table className="min-w-full divide-y divide-gray-200">
        <thead className="bg-gray-50">
          <tr>
            <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Name</th>
            <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Status</th>
            <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">CPU Capacity (m)</th>
            <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Memory Capacity (Mi)</th>
          </tr>
        </thead>
        <tbody className="bg-white divide-y divide-gray-200">
          {nodes.map((node) => (
            <tr key={node.name} className={loading ? 'opacity-50' : ''}>
              <td className="px-6 py-4 whitespace-nowrap text-sm font-medium text-gray-900">{node.name}</td>
              <td className="px-6 py-4 whitespace-nowrap text-sm">
                <StatusBadge status={node.status} />
              </td>
              <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-600">{node.cpuCapacity}</td>
              <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-600">{node.memCapacity}</td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  )
}

export default NodeList

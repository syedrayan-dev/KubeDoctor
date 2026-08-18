import StatusBadge from './StatusBadge'

function PodList({ pods, loading }) {
  return (
    <div className="bg-white shadow rounded-lg overflow-hidden">
      <table className="min-w-full divide-y divide-gray-200">
        <thead className="bg-gray-50">
          <tr>
            <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Name</th>
            <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Namespace</th>
            <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Status</th>
            <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Restarts</th>
            <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">CPU (m)</th>
            <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Memory (Mi)</th>
            <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Node</th>
            <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Age</th>
          </tr>
        </thead>
        <tbody className="bg-white divide-y divide-gray-200">
          {pods.map((pod) => (
            <tr key={pod.id} className={loading ? 'opacity-50' : ''}>
              <td className="px-6 py-4 whitespace-nowrap text-sm font-medium text-gray-900">{pod.name}</td>
              <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-600">{pod.namespace}</td>
              <td className="px-6 py-4 whitespace-nowrap text-sm">
                <StatusBadge status={pod.status} />
              </td>
              <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-600">{pod.restarts}</td>
              <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-600">{pod.cpu}</td>
              <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-600">{pod.mem}</td>
              <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-600">{pod.node || '-'}</td>
              <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-600">{pod.age}</td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  )
}

export default PodList

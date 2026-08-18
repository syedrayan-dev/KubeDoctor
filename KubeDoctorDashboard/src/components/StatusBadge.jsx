function StatusBadge({ status }) {
  const getStatusColor = (status) => {
    switch (status) {
      case 'Running':
        return 'bg-green-100 text-green-800'
      case 'Pending':
        return 'bg-yellow-100 text-yellow-800'
      case 'Failed':
        return 'bg-red-100 text-red-800'
      case 'CrashLoopBackOff':
        return 'bg-red-100 text-red-800'
      case 'ImagePullBackOff':
        return 'bg-orange-100 text-orange-800'
      case 'ErrImagePull':
        return 'bg-orange-100 text-orange-800'
      case 'OOMKilled':
        return 'bg-red-100 text-red-800'
      case 'Ready':
        return 'bg-green-100 text-green-800'
      case 'NotReady':
        return 'bg-red-100 text-red-800'
      default:
        return 'bg-gray-100 text-gray-800'
    }
  }

  return (
    <span className={`px-3 py-1 rounded-full text-sm font-medium ${
      getStatusColor(status)
    }`}>
      {status}
    </span>
  )
}

export default StatusBadge

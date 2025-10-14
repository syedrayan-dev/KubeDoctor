import express from 'express';
import cors from 'cors';

const app = express();
const PORT = 8080;

// Middleware
app.use(cors());
app.use(express.json());

// Mock data
const mockNamespaces = ['default', 'kube-system', 'apollo-system', 'monitoring'];

const mockPods = [
  {
    metadata: {
      name: 'nginx-deployment-abc123',
      namespace: 'default',
      uid: 'pod-1',
      creationTimestamp: '2024-01-15T10:30:00Z'
    },
    spec: {
      nodeName: 'worker-node-1',
      containers: [
        { name: 'nginx', image: 'nginx:1.21' }
      ]
    },
    status: {
      phase: 'Running',
      conditions: [
        { type: 'Ready', status: 'True' },
        { type: 'PodScheduled', status: 'True' }
      ]
    }
  },
  {
    metadata: {
      name: 'crashloop-pod-def456',
      namespace: 'default',
      uid: 'pod-2',
      creationTimestamp: '2024-01-15T11:00:00Z'
    },
    spec: {
      nodeName: 'worker-node-2',
      containers: [
        { name: 'app', image: 'myapp:latest' }
      ]
    },
    status: {
      phase: 'Failed',
      conditions: [
        { type: 'Ready', status: 'False', reason: 'ContainerFailed' },
        { type: 'PodScheduled', status: 'True' }
      ]
    }
  },
  {
    metadata: {
      name: 'image-pull-pod-ghi789',
      namespace: 'apollo-system',
      uid: 'pod-3',
      creationTimestamp: '2024-01-15T11:15:00Z'
    },
    spec: {
      nodeName: 'worker-node-1',
      containers: [
        { name: 'worker', image: 'nonexistent:latest' }
      ]
    },
    status: {
      phase: 'Pending',
      conditions: [
        { type: 'PodScheduled', status: 'False', reason: 'SchedulingGated' }
      ]
    }
  }
];

const mockPolicies = [
  {
    metadata: {
      name: 'general-diagnosis-policy',
      namespace: 'apollo-system',
      uid: 'policy-1',
      creationTimestamp: '2024-01-10T09:00:00Z'
    },
    spec: {
      targetNamespaces: ['default', 'apollo-system'],
      llmConfig: {
        provider: 'openai',
        model: 'gpt-4',
        secretRef: {
          name: 'openai-secret',
          key: 'api-key'
        }
      }
    }
  },
  {
    metadata: {
      name: 'production-diagnosis-policy',
      namespace: 'apollo-system',
      uid: 'policy-2',
      creationTimestamp: '2024-01-10T09:30:00Z'
    },
    spec: {
      targetNamespaces: ['production'],
      llmConfig: {
        provider: 'openai',
        model: 'gpt-4-turbo',
        secretRef: {
          name: 'openai-secret',
          key: 'api-key'
        }
      }
    }
  }
];

const mockRequests = [
  {
    metadata: {
      name: 'nginx-deployment-abc123-manual-1705567800',
      namespace: 'default',
      uid: 'request-1',
      creationTimestamp: '2024-01-15T12:00:00Z'
    },
    spec: {
      targetPod: {
        name: 'nginx-deployment-abc123',
        namespace: 'default'
      },
      policyRef: {
        name: 'general-diagnosis-policy',
        namespace: 'apollo-system'
      },
      type: 'manual',
      manual: true
    },
    status: {
      phase: 'Completed',
      message: 'Diagnosis completed successfully',
      lastUpdateTime: '2024-01-15T12:05:00Z',
      completionTime: '2024-01-15T12:05:00Z'
    }
  },
  {
    metadata: {
      name: 'crashloop-pod-def456-manual-1705568400',
      namespace: 'default',
      uid: 'request-2',
      creationTimestamp: '2024-01-15T12:10:00Z'
    },
    spec: {
      targetPod: {
        name: 'crashloop-pod-def456',
        namespace: 'default'
      },
      policyRef: {
        name: 'general-diagnosis-policy',
        namespace: 'apollo-system'
      },
      type: 'manual',
      manual: true
    },
    status: {
      phase: 'InProgress',
      message: 'Collecting pod data and performing analysis',
      lastUpdateTime: '2024-01-15T12:12:00Z'
    }
  }
];

const mockReports = [
  {
    metadata: {
      name: 'nginx-deployment-abc123-report-1705567800',
      namespace: 'default',
      uid: 'report-1',
      creationTimestamp: '2024-01-15T12:05:00Z'
    },
    spec: {
      targetPod: {
        name: 'nginx-deployment-abc123',
        namespace: 'default'
      },
      triggerCondition: {
        type: 'Manual',
        timestamp: '2024-01-15T12:00:00Z'
      },
      analysis: {
        summary: 'Pod is running normally with no issues detected.',
        rootCause: 'No issues found. This appears to be a routine check.',
        recommendations: [
          'Continue monitoring pod performance',
          'Consider implementing health checks if not already present',
          'Review resource usage patterns'
        ],
        confidence: 0.95,
        generatedAt: '2024-01-15T12:05:00Z'
      }
    }
  }
];

// API Routes

// Namespaces
app.get('/api/v1/namespaces', (req, res) => {
  res.json({
    items: mockNamespaces.map(name => ({
      metadata: { name }
    }))
  });
});

// Pods
app.get('/api/v1/namespaces/:namespace/pods', (req, res) => {
  const { namespace } = req.params;
  const { search } = req.query;
  
  let filteredPods = mockPods.filter(pod => pod.metadata.namespace === namespace);
  
  if (search) {
    filteredPods = filteredPods.filter(pod =>
      pod.metadata.name.toLowerCase().includes(search.toLowerCase())
    );
  }
  
  res.json({ items: filteredPods });
});

// Policies
app.get('/api/diagnosis/v1alpha1/policies', (req, res) => {
  res.json({ items: mockPolicies });
});

// Diagnosis Requests
app.get('/api/diagnosis/v1alpha1/requests', (req, res) => {
  const { type } = req.query;
  
  let filteredRequests = mockRequests;
  if (type) {
    filteredRequests = mockRequests.filter(req => req.spec.type === type);
  }
  
  res.json({ items: filteredRequests });
});

app.get('/api/diagnosis/v1alpha1/requests/:id', (req, res) => {
  const { id } = req.params;
  const request = mockRequests.find(req => req.metadata.name === id);
  
  if (!request) {
    return res.status(404).json({ error: 'Request not found' });
  }
  
  res.json(request);
});

app.post('/api/diagnosis/v1alpha1/namespaces/:namespace/requests', (req, res) => {
  const { namespace } = req.params;
  const requestBody = req.body;
  
  // Generate a new request
  const timestamp = Math.floor(Date.now() / 1000);
  const newRequest = {
    metadata: {
      name: `${requestBody.spec.targetPod.name}-${requestBody.spec.type}-${timestamp}`,
      namespace: namespace,
      uid: `request-${Date.now()}`,
      creationTimestamp: new Date().toISOString()
    },
    spec: requestBody.spec,
    status: {
      phase: 'Pending',
      message: 'Request queued for processing',
      lastUpdateTime: new Date().toISOString()
    }
  };
  
  mockRequests.push(newRequest);
  
  // Simulate request processing
  setTimeout(() => {
    newRequest.status.phase = 'InProgress';
    newRequest.status.message = 'Collecting pod data and performing analysis';
    newRequest.status.lastUpdateTime = new Date().toISOString();
  }, 2000);
  
  setTimeout(() => {
    newRequest.status.phase = 'Completed';
    newRequest.status.message = 'Diagnosis completed successfully';
    newRequest.status.lastUpdateTime = new Date().toISOString();
    newRequest.status.completionTime = new Date().toISOString();
  }, 8000);
  
  res.status(201).json(newRequest);
});

app.delete('/api/diagnosis/v1alpha1/requests/:id', (req, res) => {
  const { id } = req.params;
  const index = mockRequests.findIndex(req => req.metadata.name === id);
  
  if (index === -1) {
    return res.status(404).json({ error: 'Request not found' });
  }
  
  mockRequests.splice(index, 1);
  res.status(204).send();
});

// Reports
app.get('/api/diagnosis/v1alpha1/reports', (req, res) => {
  res.json({ 
    items: mockReports,
    totalCount: mockReports.length
  });
});

app.get('/api/diagnosis/v1alpha1/reports/:id', (req, res) => {
  const { id } = req.params;
  const report = mockReports.find(rep => rep.metadata.name === id);
  
  if (!report) {
    return res.status(404).json({ error: 'Report not found' });
  }
  
  res.json(report);
});

// Dashboard metrics
app.get('/api/diagnosis/v1alpha1/metrics', (req, res) => {
  res.json({
    totalReports: mockReports.length,
    totalRequests: mockRequests.length,
    activeRequests: mockRequests.filter(req => req.status.phase === 'InProgress' || req.status.phase === 'Pending').length,
    recentReports: mockReports.slice(-5),
    problemDistribution: [
      { name: 'Failed', value: 2 },
      { name: 'Pending', value: 1 },
      { name: 'Running', value: 1 },
      { name: 'Unknown', value: 0 }
    ]
  });
});

// Health check
app.get('/health', (req, res) => {
  res.json({ status: 'healthy', timestamp: new Date().toISOString() });
});

// Start server
app.listen(PORT, () => {
  console.log(`Mock API server running on http://localhost:${PORT}`);
  console.log('Available endpoints:');
  console.log('  GET  /api/v1/namespaces');
  console.log('  GET  /api/v1/namespaces/:namespace/pods');
  console.log('  GET  /api/diagnosis/v1alpha1/policies');
  console.log('  GET  /api/diagnosis/v1alpha1/requests');
  console.log('  POST /api/diagnosis/v1alpha1/namespaces/:namespace/requests');
  console.log('  GET  /api/diagnosis/v1alpha1/reports');
  console.log('  GET  /api/diagnosis/v1alpha1/metrics');
  console.log('  GET  /health');
});
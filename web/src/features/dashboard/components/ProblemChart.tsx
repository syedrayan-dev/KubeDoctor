import React from 'react';
import Card, { CardHeader, CardTitle, CardBody } from '../../../shared/components/ui/Card';
import PieChart from '../../../shared/components/charts/PieChart';
import { ChartPieIcon } from '@heroicons/react/24/outline';
import { CHART_COLORS, POD_PHASE_COLORS } from '../../../shared/constants/colors';
import { PROBLEM_TYPE_LABELS } from '../../../shared/constants/status';
import type { PieChartData } from '../../../shared/types/common';

interface ProblemDistribution {
  type: string;
  count: number;
  percentage: number;
}

interface ProblemChartProps {
  data: ProblemDistribution[];
  loading?: boolean;
}

const ProblemChart: React.FC<ProblemChartProps> = ({ data, loading = false }) => {
  // Define order for consistent display
  const typeOrder = ['Failed', 'Pending', 'Running', 'Unknown'];
  
  // Sort data by predefined order
  const sortedData = [...data].sort((a, b) => {
    const aIndex = typeOrder.indexOf(a.type);
    const bIndex = typeOrder.indexOf(b.type);
    return (aIndex === -1 ? 999 : aIndex) - (bIndex === -1 ? 999 : bIndex);
  });
  
  // Convert data to chart format
  const chartData: PieChartData[] = sortedData.map((item, index) => ({
    name: PROBLEM_TYPE_LABELS[item.type as keyof typeof PROBLEM_TYPE_LABELS] || item.type,
    value: item.count,
    color: POD_PHASE_COLORS[item.type as keyof typeof POD_PHASE_COLORS] || CHART_COLORS[index % CHART_COLORS.length],
  }));

  if (loading) {
    return (
      <Card className="h-[480px] flex flex-col">
        <CardHeader>
          <CardTitle>Problem Distribution</CardTitle>
        </CardHeader>
        <CardBody className="flex-1 overflow-hidden">
          <div className="animate-pulse h-full">
            <div className="h-full bg-gray-200 rounded"></div>
          </div>
        </CardBody>
      </Card>
    );
  }

  // Check if data is empty or all counts are zero
  const totalCount = data.reduce((sum, item) => sum + item.count, 0);
  
  if (data.length === 0 || totalCount === 0) {
    return (
      <Card className="h-[480px] flex flex-col">
        <CardHeader>
          <CardTitle>Problem Distribution</CardTitle>
        </CardHeader>
        <CardBody className="flex-1 overflow-hidden">
          <div className="text-center h-full flex flex-col justify-center">
            <ChartPieIcon className="h-12 w-12 text-gray-400 mx-auto mb-4" />
            <h3 className="text-base font-medium text-gray-900 mb-2">No diagnosis reports yet</h3>
            <p className="text-sm text-gray-600">
              Problem distribution will appear here once Apollo starts analyzing pods.
            </p>
          </div>
        </CardBody>
      </Card>
    );
  }

  const total = sortedData.reduce((sum, item) => sum + item.count, 0);

  return (
    <Card className="h-[480px] flex flex-col">
      <CardHeader>
        <CardTitle>Problem Distribution</CardTitle>
      </CardHeader>
      <CardBody className="flex-1 overflow-hidden">
        <div className="h-full flex flex-col">
          {/* Chart */}
          <div className="flex justify-center flex-shrink-0 mb-4">
            <PieChart 
              data={chartData} 
              height={240}
              showLegend={false}
            />
          </div>
          
          {/* Statistics */}
          <div className="space-y-4 flex-1 flex flex-col justify-start">
            <div>
              <div className="grid grid-cols-2 gap-3">
                {sortedData.map((item, index) => (
                  <div key={item.type} className="flex items-center justify-between">
                    <div className="flex items-center">
                      <div 
                        className="w-3 h-3 rounded-full mr-2"
                        style={{ backgroundColor: POD_PHASE_COLORS[item.type as keyof typeof POD_PHASE_COLORS] || CHART_COLORS[index % CHART_COLORS.length] }}
                      ></div>
                      <span className="text-sm text-gray-700">
                        {PROBLEM_TYPE_LABELS[item.type as keyof typeof PROBLEM_TYPE_LABELS] || item.type}
                      </span>
                    </div>
                    <div className="text-right ml-2">
                      <div className="text-sm font-medium text-gray-900">
                        {item.count} ({item.percentage.toFixed(1)}%)
                      </div>
                    </div>
                  </div>
                ))}
              </div>
            </div>
            
            {/* Total */}
            <div className="pt-3 border-t border-gray-200">
              <div className="flex items-center justify-between">
                <span className="text-sm font-medium text-gray-700">Total Issues</span>
                <span className="text-lg font-bold text-gray-900">{total}</span>
              </div>
            </div>
          </div>
        </div>
      </CardBody>
    </Card>
  );
};

export default ProblemChart;
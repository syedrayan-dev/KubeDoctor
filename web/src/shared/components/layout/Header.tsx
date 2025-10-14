import React from 'react';
import { 
  Bars3Icon, 
  ArrowPathIcon 
} from '@heroicons/react/24/outline';
import Button from '../ui/Button';

interface HeaderProps {
  onMenuClick: () => void;
  title?: string;
  subtitle?: string;
  onRefresh?: () => void;
}

const Header: React.FC<HeaderProps> = ({ 
  onMenuClick, 
  title = 'Dashboard', 
  subtitle,
  onRefresh
}) => {
  const [lastUpdated, setLastUpdated] = React.useState<Date>(new Date());

  const handleRefresh = () => {
    setLastUpdated(new Date());
    if (onRefresh) {
      onRefresh();
    }
  };

  const formatLastUpdated = (date: Date) => {
    return date.toLocaleTimeString('en-US', {
      hour12: false,
      hour: '2-digit',
      minute: '2-digit',
      second: '2-digit',
    });
  };

  return (
    <header className="bg-white border-b border-gray-200 px-4 py-4">
      <div className="flex items-center justify-between">
        {/* Left side */}
        <div className="flex items-center">
          {/* Mobile menu button */}
          <button
            type="button"
            className="lg:hidden p-2 rounded-md text-gray-400 hover:text-gray-500 hover:bg-gray-100 mr-4"
            onClick={onMenuClick}
          >
            <Bars3Icon className="h-6 w-6" />
          </button>

          {/* Page title */}
          <div>
            <h1 className="text-xl font-semibold text-gray-900">{title}</h1>
            {subtitle && (
              <p className="text-sm text-gray-500 mt-1">{subtitle}</p>
            )}
          </div>
        </div>

        {/* Right side */}
        <div className="flex items-center space-x-4">

          {/* Last updated info */}
          <div className="hidden md:flex items-center text-sm text-gray-500">
            <span>Last updated: {formatLastUpdated(lastUpdated)}</span>
          </div>

          {/* Refresh button */}
          <Button
            variant="ghost"
            size="sm"
            onClick={handleRefresh}
            className="p-2"
          >
            <ArrowPathIcon className="h-5 w-5" />
          </Button>

        </div>
      </div>
    </header>
  );
};

export default Header;
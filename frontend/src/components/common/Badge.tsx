import React from 'react';
import { LeadStatus, CallOutcome, AIIntent } from '../../types';

interface BadgeProps {
  status: string;
  size?: 'sm' | 'md' | 'lg';
  className?: string;
}

export const Badge: React.FC<BadgeProps> = ({ status, size = 'md', className = '' }) => {
  const getBadgeStyle = (val: string) => {
    switch (val) {
      // Urgent alerts / follow-up required
      case 'FOLLOW_UP_REQUIRED':
        return 'bg-red-50 text-red-700 border-red-200 animate-alert-pulse font-semibold';
      case 'OVERDUE':
        return 'bg-rose-50 text-rose-700 border-rose-200 font-semibold';
      case 'CALLED_ONCE':
        return 'bg-amber-50 text-amber-700 border-amber-200 font-semibold';

      // Do Not Call / Not Interested
      case 'DO_NOT_CALL':
        return 'bg-purple-50 text-purple-700 border-purple-200 font-bold';
      case 'NOT_INTERESTED':
        return 'bg-slate-100 text-slate-600 border-slate-200 font-medium';

      // Positive & active
      case 'INTERESTED':
        return 'bg-emerald-50 text-emerald-700 border-emerald-200 font-bold';
      case 'CONVERTED':
        return 'bg-green-50 text-green-700 border-green-200 font-bold';
      case 'FOLLOW_UP_SCHEDULED':
        return 'bg-blue-50 text-blue-700 border-blue-200 font-semibold';
      case 'CALL_BACK_LATER':
        return 'bg-cyan-50 text-cyan-700 border-cyan-200 font-semibold';
      case 'INFO_REQUESTED':
        return 'bg-indigo-50 text-indigo-700 border-indigo-200 font-semibold';
      case 'UNDECIDED':
        return 'bg-yellow-50 text-yellow-700 border-yellow-200 font-semibold';

      // Negative closures
      case 'LOST':
        return 'bg-zinc-100 text-zinc-600 border-zinc-200';
      case 'WRONG_NUMBER':
        return 'bg-orange-50 text-orange-700 border-orange-200';

      // New & Assigned
      case 'NEW':
        return 'bg-cyan-50 text-cyan-700 border-cyan-200';
      case 'ASSIGNED':
        return 'bg-teal-50 text-teal-700 border-teal-200';
      case 'CONTACTED':
        return 'bg-blue-50 text-blue-700 border-blue-200';

      default:
        return 'bg-slate-100 text-slate-700 border-slate-200';
    }
  };

  const sizeClasses = {
    sm: 'text-[10px] px-2 py-0.5',
    md: 'text-xs px-2.5 py-1',
    lg: 'text-sm px-3 py-1.5',
  }[size];

  const formatText = (text: string) => {
    return text.replace(/_/g, ' ');
  };

  return (
    <span
      className={`inline-flex items-center gap-1 rounded-full border tracking-wide uppercase font-medium ${sizeClasses} ${getBadgeStyle(
        status
      )} ${className}`}
    >
      <span className="w-1.5 h-1.5 rounded-full bg-current opacity-80" />
      {formatText(status)}
    </span>
  );
};

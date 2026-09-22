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
        return 'bg-red-500/15 text-red-400 border-red-500/30 animate-alert-pulse font-semibold';
      case 'OVERDUE':
        return 'bg-rose-500/15 text-rose-400 border-rose-500/30';
      case 'CALLED_ONCE':
        return 'bg-amber-500/15 text-amber-300 border-amber-500/30';

      // Do Not Call / Not Interested
      case 'DO_NOT_CALL':
        return 'bg-purple-500/20 text-purple-300 border-purple-500/40 font-bold';
      case 'NOT_INTERESTED':
        return 'bg-slate-700/50 text-slate-300 border-slate-600/50';

      // Positive & active
      case 'INTERESTED':
        return 'bg-emerald-500/15 text-emerald-400 border-emerald-500/30';
      case 'CONVERTED':
        return 'bg-green-500/20 text-green-300 border-green-500/40 font-bold';
      case 'FOLLOW_UP_SCHEDULED':
        return 'bg-blue-500/15 text-blue-400 border-blue-500/30';
      case 'CALL_BACK_LATER':
        return 'bg-sky-500/15 text-sky-300 border-sky-500/30';
      case 'INFO_REQUESTED':
        return 'bg-indigo-500/15 text-indigo-300 border-indigo-500/30';
      case 'UNDECIDED':
        return 'bg-yellow-500/15 text-yellow-400 border-yellow-500/30';

      // Negative closures
      case 'LOST':
        return 'bg-zinc-800 text-zinc-400 border-zinc-700';
      case 'WRONG_NUMBER':
        return 'bg-orange-500/15 text-orange-400 border-orange-500/30';

      // New & Assigned
      case 'NEW':
        return 'bg-cyan-500/15 text-cyan-400 border-cyan-500/30';
      case 'ASSIGNED':
        return 'bg-teal-500/15 text-teal-400 border-teal-500/30';
      case 'CONTACTED':
        return 'bg-blue-600/15 text-blue-300 border-blue-600/30';

      default:
        return 'bg-gray-800 text-gray-300 border-gray-700';
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

import React from 'react';
import { NavLink } from 'react-router-dom';
import {
  LayoutDashboard,
  Users,
  PhoneCall,
  PhoneForwarded,
  Sparkles,
} from 'lucide-react';
import { useQuery } from '@tanstack/react-query';
import { leadsApi } from '../../services/api';

export const Sidebar: React.FC = () => {
  const { data: queueData } = useQuery({
    queryKey: ['bdaQueueCount'],
    queryFn: () => leadsApi.getBdaQueue({ category: 'all', page: 1, pageSize: 1 }),
    refetchInterval: 30000,
  });

  const pendingCount = queueData?.categoryCounts?.all ?? 0;

  const navItems = [
    {
      to: '/dashboard',
      label: 'Dashboard',
      icon: <LayoutDashboard className="w-4 h-4" />,
    },
    {
      to: '/call-queue',
      label: 'BDA Call Station',
      icon: <PhoneCall className="w-4 h-4 text-emerald-400" />,
      badge: pendingCount > 0 ? `${pendingCount} Due` : undefined,
      badgeColor: 'bg-emerald-500/20 text-emerald-300 border border-emerald-500/30',
    },
    {
      to: '/leads',
      label: 'My Students',
      icon: <Users className="w-4 h-4 text-indigo-400" />,
    },
  ];

  return (
    <aside className="w-64 border-r border-slate-800 bg-slate-900/50 backdrop-blur-md flex flex-col h-[calc(100vh-4rem)] sticky top-16">
      <div className="p-4 flex-1 space-y-1.5">
        <div className="px-3 py-2 text-[10px] uppercase font-bold text-slate-500 tracking-wider">
          Sales BDA Workstation
        </div>
        {navItems.map((item) => (
          <NavLink
            key={item.to}
            to={item.to}
            className={({ isActive }) =>
              `flex items-center justify-between px-3.5 py-3 rounded-xl text-xs font-semibold transition-all ${isActive
                ? 'bg-indigo-600 text-white border border-indigo-500/50 shadow-md shadow-indigo-600/20'
                : 'text-slate-400 hover:text-slate-200 hover:bg-slate-800/50'
              }`
            }
          >
            <div className="flex items-center gap-3">
              {item.icon}
              <span>{item.label}</span>
            </div>
            {item.badge && (
              <span className={`text-[10px] font-bold px-2 py-0.5 rounded-full ${item.badgeColor}`}>
                {item.badge}
              </span>
            )}
          </NavLink>
        ))}

        {/* Quick BDA Focus Helper */}
        <div className="mt-8 p-3.5 rounded-xl bg-slate-800/40 border border-slate-800 text-slate-400 text-xs space-y-2">
          <div className="flex items-center gap-1.5 text-indigo-300 font-bold text-[11px]">
            <Sparkles className="w-3.5 h-3.5" />
            <span>Calling Rule</span>
          </div>
          <p className="text-[11px] leading-relaxed text-slate-400">
            Never give up after a single missed attempt. Re-try 1-call & 2-call unattended students, and fulfill callbacks on time!
          </p>
        </div>
      </div>

      {/* Footer status */}
      <div className="p-4 border-t border-slate-800/60 bg-slate-950/40">
        <div className="flex items-center justify-between text-[11px] text-slate-400">
          <span className="flex items-center gap-1.5">
            <span className="w-2 h-2 rounded-full bg-emerald-400 animate-pulse" />
            Calling Queue Active
          </span>
          <span className="text-[10px] text-slate-500 font-mono">Auto-Sync</span>
        </div>
      </div>
    </aside>
  );
};

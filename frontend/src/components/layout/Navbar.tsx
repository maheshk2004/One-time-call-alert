import React, { useState } from 'react';
import { useAuth } from '../../context/AuthContext';
import { NotificationCenter } from './NotificationCenter';
import { LogOut, Play, RefreshCw, UserCheck, PhoneCall } from 'lucide-react';
import { followUpsApi } from '../../services/api';

export const Navbar: React.FC = () => {
  const { user, login, logout } = useAuth();
  const [isDetecting, setIsDetecting] = useState(false);
  const [detectMsg, setDetectMsg] = useState<string | null>(null);

  const bdaReps = [
    { name: 'Vikram Sharma', email: 'vikram@leadfollowup.com' },
    { name: 'Priya Patel', email: 'priya@leadfollowup.com' },
    { name: 'Rahul Varma', email: 'rahul@leadfollowup.com' },
  ];

  const handleRunDetector = async () => {
    setIsDetecting(true);
    setDetectMsg(null);
    try {
      const res = await followUpsApi.triggerDetection();
      setDetectMsg(`${res.leadsFlaggedForAlert} leads updated`);
      setTimeout(() => setDetectMsg(null), 4000);
      window.dispatchEvent(new CustomEvent('lead-status-updated'));
    } catch (err: any) {
      setDetectMsg('Failed');
      setTimeout(() => setDetectMsg(null), 3000);
    } finally {
      setIsDetecting(false);
    }
  };

  const handleSwitchBda = async (email: string) => {
    try {
      await login(email, 'password123');
      window.location.reload();
    } catch (err) {
      console.error('Failed to switch BDA rep:', err);
    }
  };

  return (
    <header className="h-16 border-b border-slate-200 bg-white/95 backdrop-blur-md px-6 flex items-center justify-between sticky top-0 z-40 shadow-xs">
      <div className="flex items-center gap-3">
        <div className="flex items-center gap-2.5">
          <div className="w-9 h-9 rounded-xl bg-gradient-to-tr from-indigo-600 via-indigo-500 to-violet-500 flex items-center justify-center font-black text-white shadow-md shadow-indigo-500/20">
            <PhoneCall className="w-5 h-5" />
          </div>
          <div>
            <h1 className="text-sm font-bold tracking-tight text-slate-900 flex items-center gap-2">
              Sales BDA Calling Station
              <span className="text-[10px] uppercase font-semibold px-2 py-0.5 rounded-full bg-emerald-50 text-emerald-700 border border-emerald-200">
                Live
              </span>
            </h1>
            <p className="text-[11px] text-slate-500">1-Call & 2-Call Outreach & Callback Tracker</p>
          </div>
        </div>
      </div>

      <div className="flex items-center gap-4">
        {/* Run Overdue Scanner Button */}
        <button
          onClick={handleRunDetector}
          disabled={isDetecting}
          className="flex items-center gap-1.5 px-3 py-1.5 rounded-xl bg-indigo-50 hover:bg-indigo-100 border border-indigo-200 text-indigo-700 text-xs font-semibold transition-all"
          title="Scan unattended and overdue calls"
        >
          {isDetecting ? (
            <RefreshCw className="w-3.5 h-3.5 animate-spin text-indigo-600" />
          ) : (
            <Play className="w-3.5 h-3.5 text-indigo-600" />
          )}
          <span>{isDetecting ? 'Scanning...' : detectMsg ? detectMsg : 'Scan Overdue'}</span>
        </button>

        {/* BDA Rep Switcher */}
        <div className="hidden md:flex items-center gap-1.5 bg-slate-50 border border-slate-200 rounded-xl px-2.5 py-1 text-xs text-slate-700">
          <UserCheck className="w-3.5 h-3.5 text-emerald-600" />
          <span className="text-[11px] text-slate-500">BDA:</span>
          <select
            value={user?.email}
            onChange={(e) => handleSwitchBda(e.target.value)}
            className="bg-transparent text-xs text-slate-800 font-semibold focus:outline-none cursor-pointer"
          >
            {bdaReps.map((rep) => (
              <option key={rep.email} value={rep.email} className="bg-white text-slate-900">
                {rep.name}
              </option>
            ))}
          </select>
        </div>

        {/* Notifications */}
        <NotificationCenter />

        {/* Separator */}
        <div className="h-6 w-px bg-slate-200" />

        {/* User Info & Logout */}
        <div className="flex items-center gap-3">
          <div className="text-right hidden sm:block">
            <div className="text-xs font-bold text-slate-900">{user?.name}</div>
            <div className="text-[10px] text-indigo-600 font-semibold uppercase">Sales BDA</div>
          </div>
          <div className="w-8 h-8 rounded-xl bg-gradient-to-br from-indigo-500 to-purple-600 flex items-center justify-center text-xs font-bold text-white shadow-xs">
            {user?.name?.charAt(0) || 'B'}
          </div>
          <button
            onClick={logout}
            className="p-1.5 rounded-lg hover:bg-slate-100 text-slate-400 hover:text-rose-600 transition-colors"
            title="Log out"
          >
            <LogOut className="w-4 h-4" />
          </button>
        </div>
      </div>
    </header>
  );
};

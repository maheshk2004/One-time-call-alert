import React, { useState, useEffect } from 'react';
import { useAuth } from '../context/AuthContext';
import { dashboardApi, leadsApi } from '../services/api';
import { DashboardMetrics, Lead } from '../types';
import {
  PhoneCall,
  Clock,
  CheckCircle2,
  AlertTriangle,
  ArrowRight,
  TrendingUp,
  Sparkles,
  Phone,
  Calendar,
  Users,
  Award,
} from 'lucide-react';
import { Link } from 'react-router-dom';
import { Badge } from '../components/common/Badge';

export const DashboardPage: React.FC = () => {
  const { user } = useAuth();
  const [metrics, setMetrics] = useState<DashboardMetrics | null>(null);
  const [urgentLeads, setUrgentLeads] = useState<Lead[]>([]);
  const [loading, setLoading] = useState(true);

  const fetchMetrics = async () => {
    try {
      const [summaryData, queueData] = await Promise.all([
        dashboardApi.getSummary(),
        leadsApi.getBdaQueue({ category: 'all', page: 1, pageSize: 6 }),
      ]);
      setMetrics(summaryData);
      setUrgentLeads(queueData.items || []);
    } catch (err) {
      console.error('Failed to load BDA dashboard metrics:', err);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchMetrics();
    const handleStatusUpdate = () => fetchMetrics();
    window.addEventListener('lead-status-updated', handleStatusUpdate);
    return () => window.removeEventListener('lead-status-updated', handleStatusUpdate);
  }, []);

  if (loading || !metrics) {
    return (
      <div className="py-24 flex flex-col items-center justify-center gap-3">
        <div className="w-8 h-8 border-4 border-indigo-500/30 border-t-indigo-500 rounded-full animate-spin" />
        <p className="text-xs text-slate-400">Loading BDA Daily Station...</p>
      </div>
    );
  }

  const callsNeeded = metrics.callsNeededToday ?? (metrics.followUpRequired + metrics.oneCallLeads);
  const unattended1 = metrics.unattended1stCount ?? metrics.oneCallLeads;
  const unattended2 = metrics.unattended2ndCount ?? 0;
  const callbacksDue = metrics.callbacksDueCount ?? metrics.overdueFollowUps;
  const callsToday = metrics.callsToday || 0;
  const interested = metrics.interested || 0;

  return (
    <div className="space-y-6">
      {/* Top Banner: Greeting & Direct Action CTA */}
      <div className="bg-gradient-to-r from-slate-900 via-indigo-950 to-slate-900 border border-indigo-500/30 rounded-2xl p-6 sm:p-8 shadow-2xl relative overflow-hidden">
        <div className="flex flex-col lg:flex-row lg:items-center justify-between gap-6 relative z-10">
          <div>
            <div className="inline-flex items-center gap-2 px-3 py-1 rounded-full bg-indigo-500/20 border border-indigo-500/30 text-indigo-300 text-xs font-semibold mb-3">
              <Sparkles className="w-3.5 h-3.5 text-indigo-400" />
              Sales BDA Workstation Active
            </div>
            <h1 className="text-2xl sm:text-3xl font-extrabold text-white tracking-tight">
              Hello, {user?.name || 'Sales BDA'} 👋
            </h1>
            <p className="text-sm text-slate-300 mt-2 max-w-xl">
              You have <span className="text-indigo-400 font-bold">{callsNeeded} student calls</span> waiting for action today. Re-engage unattended leads and lock in scheduled callbacks.
            </p>
          </div>

          <div className="flex items-center gap-3 shrink-0">
            <Link
              to="/call-queue"
              className="flex items-center gap-2.5 px-6 py-3.5 rounded-xl bg-gradient-to-r from-indigo-600 via-indigo-500 to-indigo-600 hover:from-indigo-500 hover:to-indigo-400 text-white font-bold text-sm shadow-xl shadow-indigo-600/30 transition transform active:scale-95"
            >
              <PhoneCall className="w-4 h-4" />
              Start Calling Queue
              <ArrowRight className="w-4 h-4" />
            </Link>
          </div>
        </div>
      </div>

      {/* 6 Core KPI Cards for BDA Daily Success */}
      <div className="grid grid-cols-2 sm:grid-cols-3 lg:grid-cols-6 gap-3.5">
        {/* Calls Needed Today */}
        <div className="bg-slate-900/90 border border-indigo-500/30 rounded-xl p-4 shadow-lg hover:border-indigo-500/50 transition">
          <div className="flex items-center justify-between">
            <span className="text-xs text-slate-400 font-medium">Calls Needed</span>
            <PhoneCall className="w-4 h-4 text-indigo-400" />
          </div>
          <div className="text-2xl sm:text-3xl font-black text-indigo-400 mt-1">{callsNeeded}</div>
          <span className="text-[10px] text-slate-500 mt-1 block">Total pending outreach</span>
        </div>

        {/* 1st Call Missed */}
        <div className="bg-slate-900/90 border border-amber-500/30 rounded-xl p-4 shadow-lg hover:border-amber-500/50 transition">
          <div className="flex items-center justify-between">
            <span className="text-xs text-slate-400 font-medium">1st Call Missed</span>
            <AlertTriangle className="w-4 h-4 text-amber-400" />
          </div>
          <div className="text-2xl sm:text-3xl font-black text-amber-400 mt-1">{unattended1}</div>
          <span className="text-[10px] text-slate-500 mt-1 block">Unattended attempt #1</span>
        </div>

        {/* 2nd Call Missed */}
        <div className="bg-slate-900/90 border border-rose-500/30 rounded-xl p-4 shadow-lg hover:border-rose-500/50 transition">
          <div className="flex items-center justify-between">
            <span className="text-xs text-slate-400 font-medium">2nd Call Missed</span>
            <PhoneCall className="w-4 h-4 text-rose-400" />
          </div>
          <div className="text-2xl sm:text-3xl font-black text-rose-400 mt-1">{unattended2}</div>
          <span className="text-[10px] text-slate-500 mt-1 block">Unattended attempt #2</span>
        </div>

        {/* Callbacks Due */}
        <div className="bg-slate-900/90 border border-cyan-500/30 rounded-xl p-4 shadow-lg hover:border-cyan-500/50 transition">
          <div className="flex items-center justify-between">
            <span className="text-xs text-slate-400 font-medium">Call Me Later</span>
            <Clock className="w-4 h-4 text-cyan-400" />
          </div>
          <div className="text-2xl sm:text-3xl font-black text-cyan-400 mt-1">{callbacksDue}</div>
          <span className="text-[10px] text-slate-500 mt-1 block">Scheduled follow-ups</span>
        </div>

        {/* Calls Done Today */}
        <div className="bg-slate-900/90 border border-slate-800 rounded-xl p-4 shadow-lg hover:border-slate-700 transition">
          <div className="flex items-center justify-between">
            <span className="text-xs text-slate-400 font-medium">Calls Today</span>
            <CheckCircle2 className="w-4 h-4 text-emerald-400" />
          </div>
          <div className="text-2xl sm:text-3xl font-black text-emerald-400 mt-1">{callsToday}</div>
          <span className="text-[10px] text-slate-500 mt-1 block">Completed dials</span>
        </div>

        {/* Interested / Enrolled */}
        <div className="bg-slate-900/90 border border-slate-800 rounded-xl p-4 shadow-lg hover:border-slate-700 transition">
          <div className="flex items-center justify-between">
            <span className="text-xs text-slate-400 font-medium">Interested</span>
            <Award className="w-4 h-4 text-teal-400" />
          </div>
          <div className="text-2xl sm:text-3xl font-black text-teal-400 mt-1">{interested}</div>
          <span className="text-[10px] text-slate-500 mt-1 block">Positive candidates</span>
        </div>
      </div>

      {/* Immediate Call Queue Preview */}
      <div className="bg-slate-900/80 border border-slate-800 rounded-2xl p-6 shadow-xl space-y-4">
        <div className="flex items-center justify-between">
          <div>
            <h3 className="text-base font-bold text-white flex items-center gap-2">
              <PhoneCall className="w-4 h-4 text-indigo-400" />
              Immediate Outreach Priority List
            </h3>
            <p className="text-xs text-slate-400 mt-0.5">
              Top students who missed previous calls or asked for a callback.
            </p>
          </div>
          <Link
            to="/call-queue"
            className="text-xs font-semibold text-indigo-400 hover:text-indigo-300 flex items-center gap-1 transition"
          >
            Open Full Queue ({callsNeeded})
            <ArrowRight className="w-3.5 h-3.5" />
          </Link>
        </div>

        {urgentLeads.length === 0 ? (
          <div className="py-8 text-center text-slate-400 text-xs">
            No pending calls in queue right now!
          </div>
        ) : (
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-3.5">
            {urgentLeads.map((lead) => (
              <div
                key={lead.id}
                className="p-4 rounded-xl bg-slate-800/60 border border-slate-750 hover:border-indigo-500/40 transition flex flex-col justify-between gap-3 shadow-md"
              >
                <div>
                  <div className="flex items-start justify-between gap-2">
                    <span className="font-bold text-white text-sm truncate">{lead.name}</span>
                    {lead.callAttemptCount === 1 && (
                      <span className="px-1.5 py-0.5 rounded text-[10px] font-semibold bg-amber-500/20 text-amber-300 border border-amber-500/30 shrink-0">
                        1 Missed
                      </span>
                    )}
                    {lead.callAttemptCount >= 2 && (
                      <span className="px-1.5 py-0.5 rounded text-[10px] font-semibold bg-rose-500/20 text-rose-300 border border-rose-500/30 shrink-0">
                        2 Missed
                      </span>
                    )}
                    {lead.nextFollowUpAt && (
                      <span className="px-1.5 py-0.5 rounded text-[10px] font-semibold bg-cyan-500/20 text-cyan-300 border border-cyan-500/30 shrink-0">
                        Callback
                      </span>
                    )}
                  </div>

                  <p className="text-xs text-slate-400 mt-1 truncate">
                    Course: <span className="text-slate-200">{lead.course || 'Admissions Lead'}</span>
                  </p>
                  <p className="text-xs font-mono text-indigo-400 mt-1">{lead.phone}</p>
                </div>

                <div className="flex items-center justify-between pt-2 border-t border-slate-700/60">
                  <span className="text-[11px] text-slate-400">
                    Source: {lead.source || 'Website'}
                  </span>
                  <Link
                    to="/call-queue"
                    className="px-3 py-1.5 rounded-lg bg-indigo-600 hover:bg-indigo-500 text-white text-xs font-bold transition flex items-center gap-1 shadow-sm"
                  >
                    <Phone className="w-3 h-3" />
                    Call Now
                  </Link>
                </div>
              </div>
            ))}
          </div>
        )}
      </div>

      {/* Recent Calls Completed Today */}
      <div className="bg-slate-900/80 border border-slate-800 rounded-2xl p-6 shadow-xl space-y-4">
        <h3 className="text-base font-bold text-white flex items-center gap-2">
          <Clock className="w-4 h-4 text-emerald-400" />
          Recent Calls Logged Today
        </h3>

        {!metrics.recentCalls || metrics.recentCalls.length === 0 ? (
          <p className="text-xs text-slate-400 py-4 text-center">No calls logged yet today. Hit the Call Queue to start dialing!</p>
        ) : (
          <div className="divide-y divide-slate-800">
            {metrics.recentCalls.slice(0, 5).map((call) => (
              <div key={call.id} className="py-3 flex items-center justify-between gap-4">
                <div className="flex items-center gap-3">
                  <div className="w-8 h-8 rounded-lg bg-slate-800 border border-slate-700 flex items-center justify-center text-slate-300">
                    <PhoneCall className="w-4 h-4 text-indigo-400" />
                  </div>
                  <div>
                    <div className="text-xs font-semibold text-white">
                      Call Attempt #{call.attemptNumber}
                    </div>
                    <div className="text-[11px] text-slate-400 mt-0.5">
                      Duration: {call.durationSeconds}s • Status: {call.callStatus}
                      {call.notes && <span className="italic ml-2 text-slate-300">"{call.notes}"</span>}
                    </div>
                  </div>
                </div>

                <Badge status={call.callOutcome} size="sm" />
              </div>
            ))}
          </div>
        )}
      </div>
    </div>
  );
};

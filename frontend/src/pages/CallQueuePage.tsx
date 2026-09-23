import React, { useState } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import {
  Phone,
  PhoneCall,
  Clock,
  CheckCircle2,
  XCircle,
  AlertCircle,
  Search,
  RefreshCw,
  Sparkles,
  BookOpen,
  ArrowRight,
  X,
  History,
  User as UserIcon,
  Headphones,
  Bot,
  MessageSquare,
  Volume2,
  Calendar,
  AlertTriangle,
} from 'lucide-react';
import { leadsApi, followUpsApi } from '../services/api';
import { Lead, BdaQueueCategory } from '../types';
import { LiveCallModal } from '../components/calling/LiveCallModal';

export const CallQueuePage: React.FC = () => {
  const queryClient = useQueryClient();
  const [category, setCategory] = useState<BdaQueueCategory>('all');
  const [search, setSearch] = useState('');
  const [activeCallLead, setActiveCallLead] = useState<Lead | null>(null);

  // Fetch BDA Calling Queue
  const { data, isLoading, refetch, isFetching } = useQuery({
    queryKey: ['bdaQueue', category, search],
    queryFn: () => leadsApi.getBdaQueue({ category, search, page: 1, pageSize: 50 }),
    refetchInterval: 30000,
  });

  // Scan Overdue Leads Mutation
  const detectorMutation = useMutation({
    mutationFn: () => followUpsApi.triggerDetection(),
    onSuccess: (res) => {
      queryClient.invalidateQueries({ queryKey: ['bdaQueue'] });
      queryClient.invalidateQueries({ queryKey: ['dashboardMetrics'] });
      alert(`Overdue check complete: ${res.leadsFlaggedForAlert} leads updated.`);
    },
  });

  const counts = data?.categoryCounts || {
    all: 0,
    unattended_1: 0,
    unattended_2: 0,
    callback_later: 0,
    fresh: 0,
    interested: 0,
  };

  const tabs: { id: BdaQueueCategory; label: string; count: number }[] = [
    { id: 'all', label: 'All Calls Needed', count: counts.all },
    { id: 'unattended_1', label: '📞 1st Call Unattended', count: counts.unattended_1 },
    { id: 'unattended_2', label: '🔁 2nd Call Unattended', count: counts.unattended_2 },
    { id: 'callback_later', label: '⏰ Call Me Later', count: counts.callback_later },
    { id: 'fresh', label: '🌟 Fresh Leads', count: counts.fresh },
    { id: 'interested', label: '🎉 Interested', count: counts.interested },
  ];

  return (
    <div className="space-y-6 bg-white min-h-screen">
      {/* Top Banner - Clean Crisp Enterprise Theme */}
      <div className="bg-white border border-slate-200 rounded-2xl p-6 shadow-sm">
        <div className="flex flex-col md:flex-row md:items-center justify-between gap-4">
          <div>
            <div className="flex items-center gap-3">
              <div className="w-10 h-10 rounded-xl bg-indigo-600 flex items-center justify-center text-white shadow-sm">
                <PhoneCall className="w-5 h-5" />
              </div>
              <div>
                <h1 className="text-2xl font-bold text-slate-900 tracking-tight flex items-center gap-2.5">
                  Sales BDA Calling Station
                  <span className="text-xs px-2.5 py-0.5 rounded-full bg-emerald-50 text-emerald-700 font-semibold border border-emerald-200">
                    Live Speech STT Active
                  </span>
                </h1>
                <p className="text-xs text-slate-500 mt-0.5">
                  Calls are transcribed live while speaking. AI processes dialogue automatically to set reminders and alerts — zero manual typing.
                </p>
              </div>
            </div>
          </div>

          <div className="flex items-center gap-3">
            <button
              onClick={() => detectorMutation.mutate()}
              disabled={detectorMutation.isPending}
              className="flex items-center gap-2 px-3.5 py-2 text-xs font-semibold rounded-xl bg-indigo-50 border border-indigo-200 text-indigo-700 hover:bg-indigo-100 transition shadow-xs"
              title="Run background check for unattended and overdue calls"
            >
              <RefreshCw className={`w-3.5 h-3.5 ${detectorMutation.isPending ? 'animate-spin' : ''}`} />
              Scan Overdue Leads
            </button>
            <button
              onClick={() => refetch()}
              className="p-2 rounded-xl bg-slate-50 border border-slate-200 text-slate-600 hover:text-slate-900 hover:bg-slate-100 transition"
              title="Refresh queue"
            >
              <RefreshCw className={`w-4 h-4 ${isFetching ? 'animate-spin' : ''}`} />
            </button>
          </div>
        </div>

        {/* Quick KPI stats strip */}
        <div className="grid grid-cols-2 sm:grid-cols-3 md:grid-cols-6 gap-3 mt-6 pt-5 border-t border-slate-100">
          <div className="bg-slate-50 rounded-xl p-3.5 border border-slate-200">
            <div className="text-[11px] text-slate-500 font-medium">Pending Follow-Ups</div>
            <div className="text-2xl font-bold text-indigo-600 mt-0.5">{counts.all}</div>
          </div>
          <div className="bg-slate-50 rounded-xl p-3.5 border border-slate-200">
            <div className="text-[11px] text-slate-500 font-medium">1st Call Unattended</div>
            <div className="text-2xl font-bold text-amber-600 mt-0.5">{counts.unattended_1}</div>
          </div>
          <div className="bg-slate-50 rounded-xl p-3.5 border border-slate-200">
            <div className="text-[11px] text-slate-500 font-medium">2nd Call Unattended</div>
            <div className="text-2xl font-bold text-rose-600 mt-0.5">{counts.unattended_2}</div>
          </div>
          <div className="bg-slate-50 rounded-xl p-3.5 border border-slate-200">
            <div className="text-[11px] text-slate-500 font-medium">AI Callback Scheduled</div>
            <div className="text-2xl font-bold text-cyan-600 mt-0.5">{counts.callback_later}</div>
          </div>
          <div className="bg-slate-50 rounded-xl p-3.5 border border-slate-200">
            <div className="text-[11px] text-slate-500 font-medium">Fresh Leads</div>
            <div className="text-2xl font-bold text-blue-600 mt-0.5">{counts.fresh}</div>
          </div>
          <div className="bg-slate-50 rounded-xl p-3.5 border border-slate-200">
            <div className="text-[11px] text-slate-500 font-medium">Positive / Interested</div>
            <div className="text-2xl font-bold text-emerald-600 mt-0.5">{counts.interested}</div>
          </div>
        </div>
      </div>

      {/* Tabs & Search Bar */}
      <div className="flex flex-col sm:flex-row items-stretch sm:items-center justify-between gap-4">
        {/* Category Tabs */}
        <div className="flex items-center gap-1.5 overflow-x-auto pb-2 sm:pb-0 scrollbar-none">
          {tabs.map((tab) => {
            const isActive = category === tab.id;
            return (
              <button
                key={tab.id}
                onClick={() => setCategory(tab.id)}
                className={`px-3.5 py-2 rounded-xl text-xs font-semibold whitespace-nowrap transition flex items-center gap-2 border ${
                  isActive
                    ? 'bg-indigo-600 text-white border-indigo-600 shadow-xs'
                    : 'bg-white text-slate-600 border-slate-200 hover:text-slate-900 hover:bg-slate-50'
                }`}
              >
                <span>{tab.label}</span>
                <span
                  className={`px-1.5 py-0.5 rounded-full text-[10px] font-bold ${
                    isActive ? 'bg-white/20 text-white' : 'bg-slate-100 text-slate-700'
                  }`}
                >
                  {tab.count}
                </span>
              </button>
            );
          })}
        </div>

        {/* Search Input */}
        <div className="relative min-w-[240px]">
          <Search className="w-4 h-4 text-slate-400 absolute left-3.5 top-1/2 -translate-y-1/2" />
          <input
            type="text"
            placeholder="Search student or phone..."
            value={search}
            onChange={(e) => setSearch(e.target.value)}
            className="w-full pl-9 pr-4 py-2 bg-white border border-slate-200 rounded-xl text-xs text-slate-800 placeholder-slate-400 focus:outline-none focus:border-indigo-600 transition shadow-xs"
          />
        </div>
      </div>

      {/* Calling Station Cards List */}
      <div className="bg-white border border-slate-200 rounded-2xl overflow-hidden shadow-xs">
        {isLoading ? (
          <div className="py-20 flex flex-col items-center justify-center gap-3">
            <RefreshCw className="w-6 h-6 text-indigo-600 animate-spin" />
            <p className="text-sm text-slate-500 font-medium">Loading calling queue...</p>
          </div>
        ) : !data?.items || data.items.length === 0 ? (
          <div className="py-16 text-center px-4">
            <div className="w-12 h-12 rounded-2xl bg-emerald-50 border border-emerald-200 mx-auto flex items-center justify-center text-emerald-600 mb-3">
              <CheckCircle2 className="w-6 h-6" />
            </div>
            <h3 className="text-base font-semibold text-slate-900">All Follow-Ups Cleared!</h3>
            <p className="text-xs text-slate-500 mt-1 max-w-md mx-auto">
              No students pending in this category. You have dialed all 1-call and 2-call unattended leads!
            </p>
          </div>
        ) : (
          <div className="divide-y divide-slate-100">
            {data.items.map((lead: Lead) => {
              const attemptCount = lead.callAttemptCount || 0;
              const hasCallback = lead.nextFollowUpAt;
              const isOverdueCallback = hasCallback && new Date(hasCallback).getTime() <= Date.now();

              return (
                <div
                  key={lead.id}
                  className="p-5 hover:bg-slate-50/70 transition flex flex-col lg:flex-row lg:items-center justify-between gap-4"
                >
                  {/* Student Details & AI Context */}
                  <div className="flex items-start gap-4 flex-1">
                    <div className="w-10 h-10 rounded-xl bg-indigo-50 border border-indigo-100 flex items-center justify-center text-indigo-700 font-bold text-sm shrink-0 mt-0.5">
                      {lead.name.charAt(0).toUpperCase()}
                    </div>
                    <div className="flex-1 min-w-0">
                      <div className="flex items-center gap-2.5 flex-wrap">
                        <span className="font-bold text-slate-900 text-sm">{lead.name}</span>
                        {lead.course && (
                          <span className="px-2 py-0.5 rounded-md text-[11px] font-semibold bg-slate-100 text-slate-700 border border-slate-200">
                            {lead.course}
                          </span>
                        )}
                        {lead.priority === 'HIGH' || lead.priority === 'URGENT' ? (
                          <span className="px-2 py-0.5 rounded-md text-[10px] font-bold bg-amber-50 text-amber-700 border border-amber-200">
                            HIGH PRIORITY
                          </span>
                        ) : null}
                      </div>

                      <div className="flex items-center gap-3 mt-1.5 text-xs text-slate-500 flex-wrap">
                        <span className="font-mono text-indigo-600 font-semibold flex items-center gap-1">
                          <Phone className="w-3.5 h-3.5" />
                          {lead.phone}
                        </span>
                        <span>•</span>
                        <span>Source: {lead.source || 'Website'}</span>
                        {lead.location && (
                          <>
                            <span>•</span>
                            <span>{lead.location}</span>
                          </>
                        )}
                      </div>

                      {/* Status Badges & Evidence Quotes */}
                      <div className="flex items-center gap-2 mt-2.5 flex-wrap">
                        {attemptCount === 0 && (
                          <span className="px-2.5 py-0.5 rounded-full text-[11px] font-semibold bg-blue-50 text-blue-700 border border-blue-200">
                            🌟 Fresh (Never Called)
                          </span>
                        )}
                        {attemptCount === 1 && (
                          <span className="px-2.5 py-0.5 rounded-full text-[11px] font-semibold bg-amber-50 text-amber-700 border border-amber-200">
                            📞 1st Call Unattended
                          </span>
                        )}
                        {attemptCount >= 2 && (
                          <span className="px-2.5 py-0.5 rounded-full text-[11px] font-semibold bg-rose-50 text-rose-700 border border-rose-200">
                            🔁 {attemptCount} Calls Unattended
                          </span>
                        )}
                        {hasCallback && (
                          <span
                            className={`px-2.5 py-0.5 rounded-full text-[11px] font-semibold flex items-center gap-1 ${
                              isOverdueCallback
                                ? 'bg-red-50 text-red-700 border border-red-200'
                                : 'bg-cyan-50 text-cyan-700 border border-cyan-200'
                            }`}
                          >
                            <Clock className="w-3 h-3" />
                            {isOverdueCallback ? 'Callback Overdue: ' : 'AI Callback: '}
                            {new Date(hasCallback).toLocaleTimeString([], {
                              hour: '2-digit',
                              minute: '2-digit',
                            })}
                          </span>
                        )}
                      </div>

                      {/* AI Extracted Evidence Quote (Why BDA must call) */}
                      {lead.notes && (
                        <div className="mt-2 text-xs bg-slate-50 border border-slate-200 rounded-lg p-2 text-slate-700 flex items-start gap-2">
                          <Bot className="w-4 h-4 text-indigo-600 shrink-0 mt-0.5" />
                          <div>
                            <span className="font-semibold text-slate-900">AI Reminder Note: </span>
                            <span className="italic">{lead.notes}</span>
                          </div>
                        </div>
                      )}
                    </div>
                  </div>

                  {/* Actions for BDA */}
                  <div className="flex items-center gap-2.5 shrink-0 self-start lg:self-center">
                    {/* Live Calling Button (Opens Live STT In-Call Studio) */}
                    <button
                      onClick={() => setActiveCallLead(lead)}
                      className="flex items-center gap-2 px-4 py-2.5 rounded-xl bg-emerald-600 hover:bg-emerald-700 text-white text-xs font-bold shadow-xs transition"
                    >
                      <PhoneCall className="w-4 h-4" />
                      <span>Start Call (Live AI STT)</span>
                    </button>

                    <a
                      href={`/leads/${lead.id}`}
                      className="px-3 py-2.5 rounded-xl bg-white hover:bg-slate-50 text-slate-600 hover:text-slate-900 text-xs font-semibold transition border border-slate-200 shadow-xs"
                      title="View student timeline"
                    >
                      History
                    </a>
                  </div>
                </div>
              );
            })}
          </div>
        )}
      </div>

      {/* LIVE CALL & REAL-TIME SPEECH TRANSCRIPTION MODAL (Zero manual typing) */}
      {activeCallLead && (
        <LiveCallModal
          lead={activeCallLead}
          onClose={() => setActiveCallLead(null)}
          onSuccess={() => {
            queryClient.invalidateQueries({ queryKey: ['bdaQueue'] });
            queryClient.invalidateQueries({ queryKey: ['dashboardMetrics'] });
          }}
        />
      )}
    </div>
  );
};

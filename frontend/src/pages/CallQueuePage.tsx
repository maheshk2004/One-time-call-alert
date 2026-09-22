import React, { useState } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import {
  Phone,
  PhoneCall,
  Clock,
  CheckCircle2,
  XCircle,
  Calendar,
  AlertCircle,
  Search,
  RefreshCw,
  Sparkles,
  BookOpen,
  ArrowRight,
  X,
  History,
  User as UserIcon,
} from 'lucide-react';
import { leadsApi, callsApi, followUpsApi } from '../services/api';
import { Lead, BdaQueueCategory } from '../types';

export const CallQueuePage: React.FC = () => {
  const queryClient = useQueryClient();
  const [category, setCategory] = useState<BdaQueueCategory>('all');
  const [search, setSearch] = useState('');
  const [activeCallLead, setActiveCallLead] = useState<Lead | null>(null);

  // Call Logger State
  const [callStatus, setCallStatus] = useState<string>('ANSWERED');
  const [callOutcome, setCallOutcome] = useState<string>('CALL_BACK_LATER');
  const [notes, setNotes] = useState<string>('');
  const [duration, setDuration] = useState<number>(60);
  const [callbackTime, setCallbackTime] = useState<string>('');
  const [customDate, setCustomDate] = useState<string>('');

  // Fetch BDA Queue
  const { data, isLoading, refetch, isFetching } = useQuery({
    queryKey: ['bdaQueue', category, search],
    queryFn: () => leadsApi.getBdaQueue({ category, search, page: 1, pageSize: 50 }),
    refetchInterval: 30000,
  });

  // Record Call Mutation
  const recordCallMutation = useMutation({
    mutationFn: async (payload: { leadId: string; data: any }) => {
      return callsApi.recordCall(payload.leadId, payload.data);
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['bdaQueue'] });
      queryClient.invalidateQueries({ queryKey: ['dashboardMetrics'] });
      setActiveCallLead(null);
      resetModalState();
    },
    onError: (err: any) => {
      alert(`Failed to record call: ${err.message}`);
    },
  });

  // Trigger Overdue Detector Mutation
  const detectorMutation = useMutation({
    mutationFn: () => followUpsApi.triggerDetection(),
    onSuccess: (res) => {
      queryClient.invalidateQueries({ queryKey: ['bdaQueue'] });
      queryClient.invalidateQueries({ queryKey: ['dashboardMetrics'] });
      alert(`Detector completed: ${res.leadsFlaggedForAlert} leads updated.`);
    },
  });

  const resetModalState = () => {
    setCallStatus('ANSWERED');
    setCallOutcome('CALL_BACK_LATER');
    setNotes('');
    setDuration(60);
    setCallbackTime('');
    setCustomDate('');
  };

  const handleOpenCallModal = (lead: Lead) => {
    setActiveCallLead(lead);
    resetModalState();
  };

  const handleSaveOutcome = (outcomeType: 'MISSED' | 'CALLBACK' | 'INTERESTED' | 'NOT_INTERESTED') => {
    if (!activeCallLead) return;

    let payload: any = {
      notes: notes.trim(),
      durationSeconds: duration,
    };

    if (outcomeType === 'MISSED') {
      payload.callStatus = 'NO_ANSWER';
      payload.callOutcome = 'FOLLOW_UP_REQUIRED';
      if (!payload.notes) payload.notes = 'Student did not attend call (No answer/busy).';
    } else if (outcomeType === 'CALLBACK') {
      payload.callStatus = 'ANSWERED';
      payload.callOutcome = 'CALL_BACK_LATER';

      let dueAt = new Date();
      if (callbackTime === '2h') {
        dueAt = new Date(Date.now() + 2 * 60 * 60 * 1000);
      } else if (callbackTime === 'evening') {
        dueAt.setHours(18, 0, 0, 0);
        if (dueAt.getTime() <= Date.now()) dueAt.setDate(dueAt.getDate() + 1);
      } else if (callbackTime === 'tomorrow_morning') {
        dueAt.setDate(dueAt.getDate() + 1);
        dueAt.setHours(10, 0, 0, 0);
      } else if (callbackTime === 'tomorrow_afternoon') {
        dueAt.setDate(dueAt.getDate() + 1);
        dueAt.setHours(14, 30, 0, 0);
      } else if (customDate) {
        dueAt = new Date(customDate);
      } else {
        dueAt = new Date(Date.now() + 24 * 60 * 60 * 1000);
      }

      payload.nextFollowUpAt = dueAt.toISOString();
      if (!payload.notes) payload.notes = 'Student requested a callback.';
    } else if (outcomeType === 'INTERESTED') {
      payload.callStatus = 'ANSWERED';
      payload.callOutcome = 'INTERESTED';
      if (!payload.notes) payload.notes = 'Student showed strong interest in enrollment.';
    } else if (outcomeType === 'NOT_INTERESTED') {
      payload.callStatus = 'ANSWERED';
      payload.callOutcome = 'NOT_INTERESTED';
      if (!payload.notes) payload.notes = 'Student is not interested. Dropped from call queue.';
    }

    recordCallMutation.mutate({ leadId: activeCallLead.id, data: payload });
  };

  const counts = data?.categoryCounts || {
    all: 0,
    unattended_1: 0,
    unattended_2: 0,
    callback_later: 0,
    fresh: 0,
    interested: 0,
  };

  const tabs: { id: BdaQueueCategory; label: string; count: number; color: string }[] = [
    { id: 'all', label: '⚡ All Calls Needed', count: counts.all, color: 'text-indigo-400' },
    { id: 'unattended_1', label: '📞 Missed Call #1', count: counts.unattended_1, color: 'text-amber-400' },
    { id: 'unattended_2', label: '🔁 Missed Call #2', count: counts.unattended_2, color: 'text-rose-400' },
    { id: 'callback_later', label: '⏰ Call Me Later', count: counts.callback_later, color: 'text-cyan-400' },
    { id: 'fresh', label: '🌟 Fresh Leads', count: counts.fresh, color: 'text-blue-400' },
    { id: 'interested', label: '🎉 Interested', count: counts.interested, color: 'text-emerald-400' },
  ];

  return (
    <div className="space-y-6">
      {/* Top Banner */}
      <div className="bg-gradient-to-r from-slate-900 via-indigo-950 to-slate-900 border border-indigo-500/20 rounded-2xl p-6 shadow-xl relative overflow-hidden">
        <div className="flex flex-col md:flex-row md:items-center justify-between gap-4 relative z-10">
          <div>
            <div className="flex items-center gap-3">
              <div className="w-10 h-10 rounded-xl bg-indigo-600/30 border border-indigo-500/40 flex items-center justify-center text-indigo-400 shadow-inner">
                <PhoneCall className="w-5 h-5" />
              </div>
              <div>
                <h1 className="text-2xl font-bold text-white tracking-tight">
                  Sales BDA Calling Station
                </h1>
                <p className="text-sm text-slate-400">
                  Target 1-call & 2-call unattended students, handle scheduled callbacks, and enroll interested candidates.
                </p>
              </div>
            </div>
          </div>

          <div className="flex items-center gap-3">
            <button
              onClick={() => detectorMutation.mutate()}
              disabled={detectorMutation.isPending}
              className="flex items-center gap-2 px-3.5 py-2 text-xs font-semibold rounded-lg bg-indigo-600/20 border border-indigo-500/30 text-indigo-300 hover:bg-indigo-600/30 transition shadow-sm"
              title="Run background check for unattended and overdue calls"
            >
              <RefreshCw className={`w-3.5 h-3.5 ${detectorMutation.isPending ? 'animate-spin' : ''}`} />
              Scan Overdue Leads
            </button>
            <button
              onClick={() => refetch()}
              className="p-2 rounded-lg bg-slate-800/80 border border-slate-700 text-slate-300 hover:text-white transition"
              title="Refresh queue"
            >
              <RefreshCw className={`w-4 h-4 ${isFetching ? 'animate-spin' : ''}`} />
            </button>
          </div>
        </div>

        {/* Quick KPI stats strip */}
        <div className="grid grid-cols-2 sm:grid-cols-3 md:grid-cols-6 gap-3 mt-6 pt-5 border-t border-slate-800/80">
          <div className="bg-slate-900/60 rounded-xl p-3 border border-slate-800">
            <div className="text-xs text-slate-400 font-medium">Calls Needed</div>
            <div className="text-2xl font-bold text-indigo-400 mt-0.5">{counts.all}</div>
          </div>
          <div className="bg-slate-900/60 rounded-xl p-3 border border-slate-800">
            <div className="text-xs text-slate-400 font-medium">1st Call Missed</div>
            <div className="text-2xl font-bold text-amber-400 mt-0.5">{counts.unattended_1}</div>
          </div>
          <div className="bg-slate-900/60 rounded-xl p-3 border border-slate-800">
            <div className="text-xs text-slate-400 font-medium">2nd Call Missed</div>
            <div className="text-2xl font-bold text-rose-400 mt-0.5">{counts.unattended_2}</div>
          </div>
          <div className="bg-slate-900/60 rounded-xl p-3 border border-slate-800">
            <div className="text-xs text-slate-400 font-medium">Call Me Later</div>
            <div className="text-2xl font-bold text-cyan-400 mt-0.5">{counts.callback_later}</div>
          </div>
          <div className="bg-slate-900/60 rounded-xl p-3 border border-slate-800">
            <div className="text-xs text-slate-400 font-medium">Fresh Leads</div>
            <div className="text-2xl font-bold text-blue-400 mt-0.5">{counts.fresh}</div>
          </div>
          <div className="bg-slate-900/60 rounded-xl p-3 border border-slate-800">
            <div className="text-xs text-slate-400 font-medium">Interested</div>
            <div className="text-2xl font-bold text-emerald-400 mt-0.5">{counts.interested}</div>
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
                className={`px-4 py-2 rounded-xl text-xs font-semibold whitespace-nowrap transition flex items-center gap-2 border ${
                  isActive
                    ? 'bg-indigo-600 text-white border-indigo-500 shadow-md shadow-indigo-600/20'
                    : 'bg-slate-900/80 text-slate-400 border-slate-800 hover:text-slate-200 hover:bg-slate-800/60'
                }`}
              >
                <span>{tab.label}</span>
                <span
                  className={`px-1.5 py-0.5 rounded-full text-[10px] font-bold ${
                    isActive ? 'bg-white/20 text-white' : 'bg-slate-800 text-slate-300'
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
            className="w-full pl-9 pr-4 py-2 bg-slate-900/90 border border-slate-800 rounded-xl text-xs text-slate-200 placeholder-slate-500 focus:outline-none focus:border-indigo-500 transition"
          />
        </div>
      </div>

      {/* Leads Table / Cards */}
      <div className="bg-slate-900/80 border border-slate-800/80 rounded-2xl overflow-hidden shadow-xl">
        {isLoading ? (
          <div className="py-20 flex flex-col items-center justify-center gap-3">
            <RefreshCw className="w-6 h-6 text-indigo-500 animate-spin" />
            <p className="text-sm text-slate-400 font-medium">Loading calling queue...</p>
          </div>
        ) : !data?.items || data.items.length === 0 ? (
          <div className="py-16 text-center px-4">
            <div className="w-12 h-12 rounded-2xl bg-slate-800/80 border border-slate-700 mx-auto flex items-center justify-center text-slate-400 mb-3">
              <CheckCircle2 className="w-6 h-6 text-emerald-400" />
            </div>
            <h3 className="text-base font-semibold text-white">All Caught Up!</h3>
            <p className="text-xs text-slate-400 mt-1 max-w-md mx-auto">
              No students pending in this category right now. Great job keeping your follow-up queue clear!
            </p>
          </div>
        ) : (
          <div className="divide-y divide-slate-800/80">
            {data.items.map((lead: Lead) => {
              const attemptCount = lead.callAttemptCount || 0;
              const hasCallback = lead.nextFollowUpAt;
              const isOverdueCallback = hasCallback && new Date(hasCallback).getTime() <= Date.now();

              return (
                <div
                  key={lead.id}
                  className="p-4 sm:p-5 hover:bg-slate-800/30 transition flex flex-col md:flex-row md:items-center justify-between gap-4"
                >
                  {/* Student Info */}
                  <div className="flex items-start gap-4">
                    <div className="w-10 h-10 rounded-xl bg-slate-800 border border-slate-700 flex items-center justify-center text-slate-300 font-semibold text-sm shrink-0 mt-0.5">
                      {lead.name.charAt(0).toUpperCase()}
                    </div>
                    <div>
                      <div className="flex items-center gap-2 flex-wrap">
                        <span className="font-semibold text-white text-sm">{lead.name}</span>
                        {lead.course && (
                          <span className="px-2 py-0.5 rounded-md text-[11px] font-medium bg-slate-800 text-slate-300 border border-slate-700">
                            {lead.course}
                          </span>
                        )}
                        {lead.priority === 'HIGH' || lead.priority === 'URGENT' ? (
                          <span className="px-2 py-0.5 rounded-md text-[10px] font-bold bg-amber-500/10 text-amber-400 border border-amber-500/20">
                            HIGH PRIORITY
                          </span>
                        ) : null}
                      </div>

                      <div className="flex items-center gap-4 mt-1.5 text-xs text-slate-400 flex-wrap">
                        <a
                          href={`tel:${lead.phone}`}
                          className="font-mono text-indigo-400 hover:underline flex items-center gap-1"
                        >
                          <Phone className="w-3 h-3" />
                          {lead.phone}
                        </a>
                        <span>•</span>
                        <span>Source: {lead.source || 'Website'}</span>
                        {lead.location && (
                          <>
                            <span>•</span>
                            <span>{lead.location}</span>
                          </>
                        )}
                      </div>

                      {/* Status & Last Call Notes */}
                      <div className="flex items-center gap-2 mt-2 flex-wrap">
                        {attemptCount === 0 && (
                          <span className="px-2 py-0.5 rounded-full text-[10px] font-semibold bg-blue-500/20 text-blue-300 border border-blue-500/30">
                            🌟 Fresh (Never Called)
                          </span>
                        )}
                        {attemptCount === 1 && (
                          <span className="px-2 py-0.5 rounded-full text-[10px] font-semibold bg-amber-500/20 text-amber-300 border border-amber-500/30">
                            📞 1st Call Unattended
                          </span>
                        )}
                        {attemptCount >= 2 && (
                          <span className="px-2 py-0.5 rounded-full text-[10px] font-semibold bg-rose-500/20 text-rose-300 border border-rose-500/30">
                            🔁 {attemptCount} Calls Unattended
                          </span>
                        )}
                        {hasCallback && (
                          <span
                            className={`px-2 py-0.5 rounded-full text-[10px] font-semibold flex items-center gap-1 ${
                              isOverdueCallback
                                ? 'bg-red-500/20 text-red-300 border border-red-500/40 animate-pulse'
                                : 'bg-cyan-500/20 text-cyan-300 border border-cyan-500/30'
                            }`}
                          >
                            <Clock className="w-3 h-3" />
                            {isOverdueCallback ? 'Callback Overdue: ' : 'Callback: '}
                            {new Date(hasCallback).toLocaleTimeString([], {
                              hour: '2-digit',
                              minute: '2-digit',
                            })}
                          </span>
                        )}
                        {lead.notes && (
                          <span className="text-[11px] text-slate-400 italic truncate max-w-xs">
                            "{lead.notes}"
                          </span>
                        )}
                      </div>
                    </div>
                  </div>

                  {/* Call Action Button */}
                  <div className="flex items-center gap-2.5 shrink-0">
                    <button
                      onClick={() => handleOpenCallModal(lead)}
                      className="flex items-center gap-2 px-4 py-2.5 rounded-xl bg-gradient-to-r from-emerald-600 to-teal-600 hover:from-emerald-500 hover:to-teal-500 text-white text-xs font-bold shadow-lg shadow-emerald-600/20 transition transform active:scale-95"
                    >
                      <PhoneCall className="w-3.5 h-3.5" />
                      Call & Log Outcome
                    </button>
                    <a
                      href={`/leads/${lead.id}`}
                      className="px-3 py-2.5 rounded-xl bg-slate-800 hover:bg-slate-700 text-slate-300 text-xs font-semibold transition border border-slate-700"
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

      {/* CALL OUTCOME MODAL */}
      {activeCallLead && (
        <div className="fixed inset-0 z-50 bg-black/80 backdrop-blur-sm flex items-center justify-center p-4">
          <div className="bg-slate-900 border border-slate-800 rounded-2xl max-w-lg w-full p-6 shadow-2xl space-y-5 animate-in fade-in zoom-in-95 duration-150">
            {/* Modal Header */}
            <div className="flex items-start justify-between">
              <div>
                <span className="text-[11px] font-bold tracking-wider uppercase text-indigo-400">
                  Calling Candidate
                </span>
                <h3 className="text-xl font-bold text-white mt-0.5">{activeCallLead.name}</h3>
                <div className="flex items-center gap-3 text-xs text-slate-400 mt-1">
                  <a
                    href={`tel:${activeCallLead.phone}`}
                    className="font-mono text-emerald-400 font-bold hover:underline flex items-center gap-1"
                  >
                    <Phone className="w-3.5 h-3.5" />
                    {activeCallLead.phone}
                  </a>
                  <span>•</span>
                  <span>{activeCallLead.course || 'Admissions Lead'}</span>
                  <span>•</span>
                  <span>Attempt #{activeCallLead.callAttemptCount + 1}</span>
                </div>
              </div>
              <button
                onClick={() => setActiveCallLead(null)}
                className="p-1.5 rounded-lg text-slate-400 hover:text-white hover:bg-slate-800 transition"
              >
                <X className="w-5 h-5" />
              </button>
            </div>

            {/* Quick Call Outcome Buttons */}
            <div className="space-y-3">
              <label className="text-xs font-semibold text-slate-300">
                Select Call Outcome (1-Click Action):
              </label>

              <div className="grid grid-cols-2 gap-2.5">
                {/* 1. Didn't Attend */}
                <button
                  type="button"
                  onClick={() => handleSaveOutcome('MISSED')}
                  disabled={recordCallMutation.isPending}
                  className="p-3.5 rounded-xl text-left border bg-rose-500/10 hover:bg-rose-500/20 border-rose-500/30 text-rose-300 transition group flex flex-col justify-between"
                >
                  <div className="flex items-center justify-between">
                    <span className="font-bold text-xs">🔴 Didn't Attend / Busy</span>
                    <PhoneCall className="w-4 h-4 text-rose-400 group-hover:scale-110 transition" />
                  </div>
                  <span className="text-[10px] text-slate-400 mt-2 block">
                    Increments to Attempt #{activeCallLead.callAttemptCount + 1} & queues for retry.
                  </span>
                </button>

                {/* 2. Interested */}
                <button
                  type="button"
                  onClick={() => handleSaveOutcome('INTERESTED')}
                  disabled={recordCallMutation.isPending}
                  className="p-3.5 rounded-xl text-left border bg-emerald-500/10 hover:bg-emerald-500/20 border-emerald-500/30 text-emerald-300 transition group flex flex-col justify-between"
                >
                  <div className="flex items-center justify-between">
                    <span className="font-bold text-xs">🟢 Interested / Enrolled</span>
                    <CheckCircle2 className="w-4 h-4 text-emerald-400 group-hover:scale-110 transition" />
                  </div>
                  <span className="text-[10px] text-slate-400 mt-2 block">
                    Responded positively! Moves to enrolled/interested track.
                  </span>
                </button>
              </div>

              {/* 3. Call Back Later Box */}
              <div className="p-3.5 rounded-xl border bg-slate-800/60 border-slate-700/80 space-y-3">
                <div className="flex items-center justify-between">
                  <span className="font-bold text-xs text-cyan-300 flex items-center gap-1.5">
                    <Clock className="w-4 h-4 text-cyan-400" />
                    ⏰ Student Asked to Call Later
                  </span>
                </div>

                <div className="grid grid-cols-2 sm:grid-cols-4 gap-1.5">
                  <button
                    type="button"
                    onClick={() => setCallbackTime('2h')}
                    className={`px-2.5 py-1.5 rounded-lg text-[11px] font-semibold border transition ${
                      callbackTime === '2h'
                        ? 'bg-cyan-600 text-white border-cyan-500'
                        : 'bg-slate-800 text-slate-300 border-slate-700 hover:bg-slate-750'
                    }`}
                  >
                    In 2 Hours
                  </button>
                  <button
                    type="button"
                    onClick={() => setCallbackTime('evening')}
                    className={`px-2.5 py-1.5 rounded-lg text-[11px] font-semibold border transition ${
                      callbackTime === 'evening'
                        ? 'bg-cyan-600 text-white border-cyan-500'
                        : 'bg-slate-800 text-slate-300 border-slate-700 hover:bg-slate-750'
                    }`}
                  >
                    Tonight (6 PM)
                  </button>
                  <button
                    type="button"
                    onClick={() => setCallbackTime('tomorrow_morning')}
                    className={`px-2.5 py-1.5 rounded-lg text-[11px] font-semibold border transition ${
                      callbackTime === 'tomorrow_morning'
                        ? 'bg-cyan-600 text-white border-cyan-500'
                        : 'bg-slate-800 text-slate-300 border-slate-700 hover:bg-slate-750'
                    }`}
                  >
                    Tomorrow 10 AM
                  </button>
                  <button
                    type="button"
                    onClick={() => setCallbackTime('tomorrow_afternoon')}
                    className={`px-2.5 py-1.5 rounded-lg text-[11px] font-semibold border transition ${
                      callbackTime === 'tomorrow_afternoon'
                        ? 'bg-cyan-600 text-white border-cyan-500'
                        : 'bg-slate-800 text-slate-300 border-slate-700 hover:bg-slate-750'
                    }`}
                  >
                    Tomorrow 2 PM
                  </button>
                </div>

                <div className="flex items-center justify-between pt-1">
                  <input
                    type="datetime-local"
                    value={customDate}
                    onChange={(e) => {
                      setCustomDate(e.target.value);
                      setCallbackTime('custom');
                    }}
                    className="px-2.5 py-1 bg-slate-900 border border-slate-700 rounded-lg text-[11px] text-slate-200 focus:outline-none focus:border-cyan-500"
                  />
                  <button
                    type="button"
                    onClick={() => handleSaveOutcome('CALLBACK')}
                    disabled={recordCallMutation.isPending}
                    className="px-4 py-1.5 rounded-lg bg-cyan-600 hover:bg-cyan-500 text-white text-xs font-bold transition shadow-sm"
                  >
                    Schedule Callback
                  </button>
                </div>
              </div>

              {/* 4. Not Interested Drop */}
              <div className="pt-2 border-t border-slate-800 flex items-center justify-between">
                <button
                  type="button"
                  onClick={() => {
                    if (confirm('Are you sure the student is NOT interested? This will remove them from your active calling queue.')) {
                      handleSaveOutcome('NOT_INTERESTED');
                    }
                  }}
                  disabled={recordCallMutation.isPending}
                  className="flex items-center gap-1.5 text-xs text-slate-400 hover:text-red-400 transition"
                >
                  <XCircle className="w-4 h-4" />
                  Student said: Not Interested / Do Not Call (Drop Lead)
                </button>
              </div>
            </div>

            {/* Optional Notes */}
            <div className="space-y-2 pt-2 border-t border-slate-800/80">
              <label className="text-xs text-slate-400 font-medium">Quick Notes (optional):</label>
              <input
                type="text"
                value={notes}
                onChange={(e) => setNotes(e.target.value)}
                placeholder="e.g. Asked for syllabus PDF, student working till 6pm..."
                className="w-full px-3.5 py-2 bg-slate-800/90 border border-slate-700 rounded-xl text-xs text-slate-200 focus:outline-none focus:border-indigo-500"
              />
            </div>
          </div>
        </div>
      )}
    </div>
  );
};

import React, { useState, useEffect } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { leadsApi, callsApi, transcriptsApi, aiApi, followUpsApi } from '../services/api';
import { Lead, CallAttempt, Transcript, AIAnalysis, CallStatus, CallOutcome, AIIntent } from '../types';
import { Badge } from '../components/common/Badge';
import { useAuth } from '../context/AuthContext';
import {
  Phone,
  PhoneCall,
  Calendar,
  Sparkles,
  ShieldBan,
  Clock,
  History,
  CheckCircle2,
  FileText,
  AlertTriangle,
  ArrowLeft,
  X,
  Volume2,
} from 'lucide-react';
import { LiveCallModal } from '../components/calling/LiveCallModal';

export const LeadDetailPage: React.FC = () => {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const { user } = useAuth();

  const [lead, setLead] = useState<Lead | null>(null);
  const [calls, setCalls] = useState<CallAttempt[]>([]);
  const [transcripts, setTranscripts] = useState<Record<string, Transcript>>({});
  const [analyses, setAnalyses] = useState<Record<string, AIAnalysis>>({});
  const [loading, setLoading] = useState(true);

  // Quick Call Modal state
  const [showCallModal, setShowCallModal] = useState(false);

  // AI Override Modal state
  const [overrideModalAnalysisId, setOverrideModalAnalysisId] = useState<string | null>(null);
  const [overrideIntent, setOverrideIntent] = useState<AIIntent>('CALL_BACK_LATER');
  const [overrideReason, setOverrideReason] = useState('');
  const [overriding, setOverriding] = useState(false);

  // Do Not Call toggle modal state
  const [showDNCModal, setShowDNCModal] = useState(false);
  const [dncReason, setDncReason] = useState('');

  const fetchLeadDetails = async () => {
    if (!id) return;
    try {
      const leadData = await leadsApi.getLead(id);
      setLead(leadData);

      const callsData = await callsApi.getCallsForLead(id);
      setCalls(callsData || []);

      // Fetch transcripts & AI analyses for calls
      const transMap: Record<string, Transcript> = {};
      const aiMap: Record<string, AIAnalysis> = {};

      for (const call of callsData || []) {
        if (call.transcriptAvailable) {
          try {
            const tr = await transcriptsApi.getTranscript(call.id);
            if (tr) transMap[call.id] = tr;
          } catch (e) {}
        }
        if (call.aiAnalysisId) {
          try {
            const an = await aiApi.getAnalysis(call.id);
            if (an) aiMap[call.id] = an;
          } catch (e) {}
        }
      }
      setTranscripts(transMap);
      setAnalyses(aiMap);
    } catch (err) {
      console.error('Failed to load lead details:', err);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchLeadDetails();
  }, [id]);

  const handleConfirmAI = async (analysisId: string) => {
    try {
      await aiApi.confirmReview(analysisId);
      fetchLeadDetails();
      window.dispatchEvent(new CustomEvent('lead-status-updated'));
    } catch (err: any) {
      alert(err.message || 'Failed to confirm review');
    }
  };

  const handleOverrideAI = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!overrideModalAnalysisId) return;
    setOverriding(true);
    try {
      await aiApi.correctReview(overrideModalAnalysisId, overrideIntent, overrideReason);
      setOverrideModalAnalysisId(null);
      setOverrideReason('');
      fetchLeadDetails();
      window.dispatchEvent(new CustomEvent('lead-status-updated'));
    } catch (err: any) {
      alert(err.message || 'Failed to update analysis');
    } finally {
      setOverriding(false);
    }
  };

  const handleToggleDNC = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!id || !lead) return;
    try {
      await leadsApi.setDoNotCall(id, !lead.doNotCall, dncReason);
      setShowDNCModal(false);
      setDncReason('');
      fetchLeadDetails();
      window.dispatchEvent(new CustomEvent('lead-status-updated'));
    } catch (err: any) {
      alert(err.message || 'Failed to update DNC status');
    }
  };

  if (loading || !lead) {
    return (
      <div className="py-20 flex flex-col items-center justify-center gap-2 bg-white">
        <div className="w-8 h-8 border-4 border-slate-200 border-t-indigo-600 rounded-full animate-spin" />
        <span className="text-xs text-slate-500">Loading lead details...</span>
      </div>
    );
  }

  return (
    <div className="space-y-6 bg-white min-h-screen">
      {/* Back Button & Header */}
      <div className="flex items-center gap-3">
        <button
          onClick={() => navigate(-1)}
          className="p-2 rounded-xl bg-slate-100 hover:bg-slate-200 text-slate-600 hover:text-slate-900 transition-colors"
        >
          <ArrowLeft className="w-4 h-4" />
        </button>
        <div>
          <div className="flex items-center gap-2">
            <h2 className="text-xl font-bold text-slate-900 tracking-tight">{lead.name}</h2>
            <Badge status={lead.status} size="md" />
            {lead.callAttemptCount === 1 && (
              <span className="text-[11px] font-bold px-2 py-0.5 rounded-full bg-amber-50 text-amber-700 border border-amber-200">
                Attempt #1
              </span>
            )}
            {lead.doNotCall && (
              <span className="text-[11px] font-bold px-2 py-0.5 rounded-full bg-purple-50 text-purple-700 border border-purple-200 flex items-center gap-1">
                <ShieldBan className="w-3.5 h-3.5" />
                Do Not Call
              </span>
            )}
          </div>
          <p className="text-xs text-slate-500 mt-0.5">
            Phone: <span className="font-mono text-slate-800 font-semibold">{lead.phone}</span> | Source:{' '}
            <span className="text-slate-700">{lead.source}</span> | Course:{' '}
            <span className="text-indigo-600 font-semibold">{lead.course || 'N/A'}</span>
          </p>
        </div>
      </div>

      {/* Action Bar */}
      <div className="bg-white p-4 rounded-xl border border-slate-200 shadow-xs flex flex-wrap items-center justify-between gap-3">
        <div className="flex items-center gap-2">
          {/* Quick Call Button */}
          <button
            onClick={() => setShowCallModal(true)}
            className="flex items-center gap-2 px-4 py-2 rounded-xl bg-emerald-600 hover:bg-emerald-700 text-white text-xs font-bold shadow-xs transition-all"
          >
            <PhoneCall className="w-4 h-4" />
            <span>Start Call (Live AI STT)</span>
          </button>

          {/* DNC Toggle */}
          <button
            onClick={() => {
              setDncReason(lead.doNotCall ? 'Reactivated by salesperson' : 'Customer stated Do Not Call');
              setShowDNCModal(true);
            }}
            className={`flex items-center gap-1.5 px-3 py-2 rounded-xl text-xs font-medium border transition-colors ${
              lead.doNotCall
                ? 'bg-purple-50 border-purple-200 text-purple-700 hover:bg-purple-100'
                : 'bg-slate-100 border-slate-200 text-slate-700 hover:text-purple-600 hover:border-purple-300'
            }`}
          >
            <ShieldBan className="w-4 h-4" />
            <span>{lead.doNotCall ? 'Reactivate (Remove DNC)' : 'Mark Do Not Call'}</span>
          </button>
        </div>

        <div className="text-xs text-slate-500 flex items-center gap-4">
          <div>
            Assigned to:{' '}
            <span className="font-semibold text-slate-800">
              {lead.assignedUser ? lead.assignedUser.name : 'Unassigned'}
            </span>
          </div>
          {lead.nextFollowUpAt && (
            <div className="flex items-center gap-1 text-cyan-700 font-semibold">
              <Clock className="w-3.5 h-3.5" />
              <span>Due: {new Date(lead.nextFollowUpAt).toLocaleString([], { dateStyle: 'medium', timeStyle: 'short' })}</span>
            </div>
          )}
        </div>
      </div>

      {/* Main Content Grid */}
      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
        {/* Left Column: Call Timeline & Transcripts & AI Analyses */}
        <div className="lg:col-span-2 space-y-6">
          <div className="flex items-center justify-between">
            <h3 className="text-sm font-bold text-slate-900 uppercase tracking-wider flex items-center gap-2">
              <Clock className="w-4 h-4 text-indigo-600" />
              <span>Call Attempts & Conversation Intelligence ({calls.length})</span>
            </h3>
          </div>

          {calls.length === 0 ? (
            <div className="bg-white p-8 rounded-2xl text-center text-slate-400 border border-slate-200 shadow-xs">
              <Phone className="w-8 h-8 text-slate-300 mx-auto mb-2" />
              <p className="text-sm font-medium text-slate-700">No calls logged yet.</p>
              <p className="text-xs text-slate-400 mt-1">
                Click "AI Process Call Transcript" above to run speech dialogue analysis.
              </p>
            </div>
          ) : (
            calls.map((call) => {
              const tr = transcripts[call.id];
              const ai = analyses[call.id];

              return (
                <div
                  key={call.id}
                  className="bg-white rounded-2xl p-5 border border-slate-200 shadow-xs space-y-4 transition-all"
                >
                  {/* Call Header */}
                  <div className="flex flex-wrap items-center justify-between gap-2 pb-3 border-b border-slate-100">
                    <div className="flex items-center gap-2">
                      <span className="text-xs font-bold px-2 py-0.5 rounded bg-indigo-50 text-indigo-700 border border-indigo-200">
                        Attempt #{call.attemptNumber}
                      </span>
                      <Badge status={call.callOutcome} size="sm" />
                      <span className="text-xs text-slate-500">({call.durationSeconds} seconds)</span>
                    </div>

                    <div className="text-xs text-slate-400 font-mono">
                      {new Date(call.callStartedAt).toLocaleString([], {
                        dateStyle: 'medium',
                        timeStyle: 'short',
                      })}
                    </div>
                  </div>

                  {/* Notes */}
                  {call.notes && (
                    <div className="text-xs text-slate-700 bg-slate-50 p-3 rounded-xl border border-slate-200">
                      <span className="text-slate-500 font-medium">Recorded Note: </span>
                      {call.notes}
                    </div>
                  )}

                  {/* AI Conversation Analysis Card */}
                  {ai && (
                    <div className="rounded-xl border border-indigo-200 bg-indigo-50/40 p-4 space-y-3">
                      <div className="flex items-center justify-between">
                        <div className="flex items-center gap-2">
                          <Sparkles className="w-4 h-4 text-indigo-600" />
                          <span className="text-xs font-bold text-slate-900">AI Conversation Analysis</span>
                          <span
                            className={`text-[10px] font-bold px-2 py-0.5 rounded-full ${
                              ai.confidence >= 0.9
                                ? 'bg-emerald-50 text-emerald-700 border border-emerald-200'
                                : 'bg-amber-50 text-amber-700 border border-amber-200'
                            }`}
                          >
                            {(ai.confidence * 100).toFixed(0)}% Confidence
                          </span>
                        </div>

                        <div className="flex items-center gap-1.5">
                          <Badge status={ai.intent} size="sm" />
                          <span
                            className={`text-[10px] uppercase font-semibold px-2 py-0.5 rounded ${
                              ai.status === 'CONFIRMED'
                                ? 'bg-emerald-50 text-emerald-700 border border-emerald-200'
                                : ai.status === 'OVERRIDDEN'
                                ? 'bg-purple-50 text-purple-700 border border-purple-200'
                                : ai.status === 'PENDING_REVIEW'
                                ? 'bg-amber-50 text-amber-700 border border-amber-200'
                                : 'bg-slate-100 text-slate-600'
                            }`}
                          >
                            {ai.status.replace(/_/g, ' ')}
                          </span>
                        </div>
                      </div>

                      <div className="text-xs text-slate-700">
                        <span className="text-slate-500 font-medium">Summary: </span>
                        {ai.summary}
                      </div>

                      {/* Evidence Quotes */}
                      {ai.evidence && ai.evidence.length > 0 && (
                        <div className="space-y-1">
                          <span className="text-[11px] text-slate-500 font-medium">Extracted Evidence:</span>
                          {ai.evidence.map((ev: any, idx: number) => (
                            <blockquote
                              key={idx}
                              className="text-xs border-l-2 border-indigo-600 pl-3 py-1 bg-white text-slate-800 italic rounded-r-lg border border-slate-200"
                            >
                              "{ev.text}"
                            </blockquote>
                          ))}
                        </div>
                      )}
                    </div>
                  )}

                  {/* Conversation Transcript */}
                  {tr && (
                    <div className="rounded-xl border border-slate-200 bg-slate-50/60 p-4 space-y-2">
                      <div className="flex items-center justify-between text-xs text-slate-500 mb-2">
                        <span className="font-semibold text-slate-900 flex items-center gap-1.5">
                          <FileText className="w-3.5 h-3.5 text-indigo-600" />
                          Dialogue Transcript ({tr.language.toUpperCase()})
                        </span>
                        <span className="text-[10px] text-slate-400">Engine: {tr.provider}</span>
                      </div>

                      <div className="space-y-2 max-h-56 overflow-y-auto pr-2">
                        {tr.speakerSegments?.map((seg: any, idx: number) => (
                          <div
                            key={idx}
                            className={`p-2.5 rounded-lg text-xs leading-relaxed ${
                              seg.speaker === 'salesperson'
                                ? 'bg-indigo-50 border border-indigo-200 ml-4 text-indigo-900'
                                : 'bg-white border border-slate-200 mr-4 text-slate-800'
                            }`}
                          >
                            <span
                              className={`text-[10px] font-bold uppercase tracking-wider block mb-0.5 ${
                                seg.speaker === 'salesperson' ? 'text-indigo-600' : 'text-emerald-600'
                              }`}
                            >
                              {seg.speaker}
                            </span>
                            {seg.text}
                          </div>
                        ))}
                      </div>
                    </div>
                  )}
                </div>
              );
            })
          )}
        </div>

        {/* Right Column: Lead Summary & Meta */}
        <div className="space-y-6">
          <div className="bg-white p-5 rounded-2xl border border-slate-200 shadow-xs space-y-4">
            <h3 className="text-sm font-bold text-slate-900 uppercase tracking-wider">Lead Profile</h3>

            <div className="space-y-3 text-xs">
              <div className="flex justify-between py-1.5 border-b border-slate-100">
                <span className="text-slate-500">Campaign</span>
                <span className="text-slate-800 font-medium">{lead.campaign || 'Admissions Campaign'}</span>
              </div>
              <div className="flex justify-between py-1.5 border-b border-slate-100">
                <span className="text-slate-500">Priority</span>
                <span className="text-slate-800 font-medium">{lead.priority}</span>
              </div>
              <div className="flex justify-between py-1.5 border-b border-slate-100">
                <span className="text-slate-500">Total Call Attempts</span>
                <span className="text-slate-900 font-bold">{lead.callAttemptCount}</span>
              </div>
              <div className="flex justify-between py-1.5 border-b border-slate-100">
                <span className="text-slate-500">Follow-Up State</span>
                <span className="text-slate-800 font-medium">{lead.followUpStatus}</span>
              </div>
              <div className="flex justify-between py-1.5 border-b border-slate-100">
                <span className="text-slate-500">Created At</span>
                <span className="text-slate-700">
                  {new Date(lead.createdAt).toLocaleDateString([], { dateStyle: 'medium' })}
                </span>
              </div>
            </div>

            {lead.doNotCallReason && (
              <div className="p-3 rounded-xl bg-purple-50 border border-purple-200 text-xs">
                <div className="font-semibold text-purple-700 mb-0.5">DNC Reason:</div>
                <div className="text-slate-700">{lead.doNotCallReason}</div>
                <div className="text-[10px] text-slate-400 mt-1">Flagged by: {lead.doNotCallDetectedBy}</div>
              </div>
            )}
          </div>
        </div>
      </div>

      {/* Live Call Modal with Real-Time STT */}
      {showCallModal && (
        <LiveCallModal
          lead={lead}
          onClose={() => setShowCallModal(false)}
          onSuccess={() => {
            fetchLeadDetails();
            window.dispatchEvent(new CustomEvent('lead-status-updated'));
          }}
        />
      )}

      {/* DNC Toggle Modal */}
      {showDNCModal && (
        <div className="fixed inset-0 z-50 bg-slate-900/60 backdrop-blur-sm flex items-center justify-center p-4">
          <div className="bg-white w-full max-w-md rounded-2xl p-6 shadow-2xl border border-slate-200 animate-in fade-in zoom-in-95">
            <h3 className="text-base font-bold text-slate-900 mb-2">
              {lead.doNotCall ? 'Reactivate Lead from DNC' : 'Flag as Do Not Call (DNC)'}
            </h3>
            <p className="text-xs text-slate-500 mb-4">
              {lead.doNotCall
                ? 'Reactivating this lead will allow future calls and follow-ups. Please provide justification.'
                : 'Marking Do Not Call will strictly cancel all follow-ups and block future alert notifications.'}
            </p>

            <form onSubmit={handleToggleDNC} className="space-y-4 text-xs">
              <div>
                <label className="block text-slate-700 font-medium mb-1">Reason *</label>
                <textarea
                  required
                  rows={3}
                  value={dncReason}
                  onChange={(e) => setDncReason(e.target.value)}
                  placeholder="Mandatory reason for audit trail..."
                  className="w-full p-2.5 bg-slate-50 border border-slate-200 rounded-lg text-slate-900 focus:outline-none focus:border-indigo-600 focus:bg-white"
                />
              </div>

              <div className="flex items-center justify-end gap-3 pt-3 border-t border-slate-200">
                <button
                  type="button"
                  onClick={() => setShowDNCModal(false)}
                  className="px-4 py-2 rounded-lg bg-slate-100 hover:bg-slate-200 text-slate-700"
                >
                  Cancel
                </button>
                <button
                  type="submit"
                  className={`px-4 py-2 rounded-lg text-white font-semibold ${
                    lead.doNotCall
                      ? 'bg-indigo-600 hover:bg-indigo-700'
                      : 'bg-purple-600 hover:bg-purple-700 shadow-xs'
                  }`}
                >
                  Confirm Change
                </button>
              </div>
            </form>
          </div>
        </div>
      )}
    </div>
  );
};

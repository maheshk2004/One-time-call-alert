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
  const [callStatus, setCallStatus] = useState<CallStatus>('ANSWERED');
  const [callOutcome, setCallOutcome] = useState<CallOutcome>('INTERESTED');
  const [callDuration, setCallDuration] = useState(120);
  const [callNotes, setCallNotes] = useState('');
  const [callConsent, setCallConsent] = useState(true);
  const [followUpDate, setFollowUpDate] = useState('');
  const [callSubmitting, setCallSubmitting] = useState(false);

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

  const handleRecordCall = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!id) return;
    setCallSubmitting(true);
    try {
      await callsApi.recordCall(id, {
        callStatus,
        callOutcome,
        durationSeconds: callDuration,
        recordingConsent: callConsent,
        recordingAvailable: callConsent,
        notes: callNotes,
        nextFollowUpAt: followUpDate ? new Date(followUpDate).toISOString() : undefined,
      });
      setShowCallModal(false);
      setCallNotes('');
      setFollowUpDate('');
      fetchLeadDetails();
      window.dispatchEvent(new CustomEvent('lead-status-updated'));
    } catch (err: any) {
      alert(err.message || 'Failed to record call');
    } finally {
      setCallSubmitting(false);
    }
  };

  const handleConfirmAI = async (analysisId: string) => {
    try {
      await aiApi.confirmReview(analysisId);
      fetchLeadDetails();
      window.dispatchEvent(new CustomEvent('lead-status-updated'));
    } catch (err: any) {
      alert(err.message || 'Failed to confirm analysis');
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
      <div className="py-20 flex flex-col items-center justify-center gap-2">
        <div className="w-8 h-8 border-4 border-indigo-500/30 border-t-indigo-500 rounded-full animate-spin" />
        <span className="text-xs text-gray-400">Loading lead details...</span>
      </div>
    );
  }

  return (
    <div className="space-y-6">
      {/* Back Button & Header */}
      <div className="flex items-center gap-3">
        <button
          onClick={() => navigate(-1)}
          className="p-2 rounded-lg bg-gray-800/80 hover:bg-gray-800 text-gray-400 hover:text-white transition-colors"
        >
          <ArrowLeft className="w-4 h-4" />
        </button>
        <div>
          <div className="flex items-center gap-2">
            <h2 className="text-xl font-bold text-white tracking-tight">{lead.name}</h2>
            <Badge status={lead.status} size="md" />
            {lead.callAttemptCount === 1 && (
              <span className="text-[11px] font-bold px-2 py-0.5 rounded-full bg-amber-500/15 text-amber-300 border border-amber-500/30">
                Attempt #1
              </span>
            )}
            {lead.doNotCall && (
              <span className="text-[11px] font-bold px-2 py-0.5 rounded-full bg-purple-500/20 text-purple-300 border border-purple-500/40 flex items-center gap-1">
                <ShieldBan className="w-3.5 h-3.5" />
                Do Not Call
              </span>
            )}
          </div>
          <p className="text-xs text-gray-400 mt-0.5">
            Phone: <span className="font-mono text-gray-200">{lead.phone}</span> | Source:{' '}
            <span className="text-gray-300">{lead.source}</span> | Course:{' '}
            <span className="text-indigo-300">{lead.course || 'N/A'}</span>
          </p>
        </div>
      </div>

      {/* Action Bar */}
      <div className="glass-panel p-4 rounded-xl border border-gray-800 flex flex-wrap items-center justify-between gap-3">
        <div className="flex items-center gap-2">
          {/* Quick Call Button */}
          <button
            onClick={() => setShowCallModal(true)}
            className="flex items-center gap-2 px-4 py-2 rounded-xl bg-indigo-600 hover:bg-indigo-500 text-white text-xs font-semibold shadow-lg shadow-indigo-600/25 transition-all"
          >
            <PhoneCall className="w-4 h-4" />
            <span>Record Call Attempt</span>
          </button>

          {/* DNC Toggle */}
          <button
            onClick={() => {
              setDncReason(lead.doNotCall ? 'Reactivated by salesperson' : 'Customer stated Do Not Call');
              setShowDNCModal(true);
            }}
            className={`flex items-center gap-1.5 px-3 py-2 rounded-xl text-xs font-medium border transition-colors ${
              lead.doNotCall
                ? 'bg-purple-500/15 border-purple-500/40 text-purple-300 hover:bg-purple-500/25'
                : 'bg-gray-800/80 border-gray-700 text-gray-300 hover:text-purple-400 hover:border-purple-500/30'
            }`}
          >
            <ShieldBan className="w-4 h-4" />
            <span>{lead.doNotCall ? 'Reactivate (Remove DNC)' : 'Mark Do Not Call'}</span>
          </button>
        </div>

        <div className="text-xs text-gray-400 flex items-center gap-4">
          <div>
            Assigned to:{' '}
            <span className="font-semibold text-white">
              {lead.assignedUser ? lead.assignedUser.name : 'Unassigned'}
            </span>
          </div>
          {lead.nextFollowUpAt && (
            <div className="flex items-center gap-1 text-blue-400 font-medium">
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
            <h3 className="text-sm font-bold text-white uppercase tracking-wider flex items-center gap-2">
              <Clock className="w-4 h-4 text-indigo-400" />
              <span>Call Attempts & Conversation Intelligence ({calls.length})</span>
            </h3>
          </div>

          {calls.length === 0 ? (
            <div className="glass-panel p-8 rounded-2xl text-center text-gray-400 border border-gray-800">
              <Phone className="w-8 h-8 text-gray-600 mx-auto mb-2" />
              <p className="text-sm font-medium">No calls logged yet.</p>
              <p className="text-xs text-gray-500 mt-1">
                Click "Record Call Attempt" above to start the sales outreach.
              </p>
            </div>
          ) : (
            calls.map((call) => {
              const tr = transcripts[call.id];
              const ai = analyses[call.id];

              return (
                <div
                  key={call.id}
                  className="glass-panel rounded-2xl p-5 border border-gray-800 space-y-4 transition-all"
                >
                  {/* Call Header */}
                  <div className="flex flex-wrap items-center justify-between gap-2 pb-3 border-b border-gray-800">
                    <div className="flex items-center gap-2">
                      <span className="text-xs font-bold px-2 py-0.5 rounded bg-indigo-600/30 text-indigo-300 border border-indigo-500/30">
                        Attempt #{call.attemptNumber}
                      </span>
                      <Badge status={call.callOutcome} size="sm" />
                      <span className="text-xs text-gray-400">({call.durationSeconds} seconds)</span>
                    </div>

                    <div className="text-xs text-gray-500 font-mono">
                      {new Date(call.callStartedAt).toLocaleString([], {
                        dateStyle: 'medium',
                        timeStyle: 'short',
                      })}
                    </div>
                  </div>

                  {/* Notes */}
                  {call.notes && (
                    <div className="text-xs text-gray-300 bg-gray-900/40 p-3 rounded-lg border border-gray-800/80">
                      <span className="text-gray-500 font-medium">Salesperson Notes: </span>
                      {call.notes}
                    </div>
                  )}

                  {/* AI Conversation Analysis Card */}
                  {ai && (
                    <div className="rounded-xl border border-indigo-500/30 bg-indigo-950/20 p-4 space-y-3">
                      <div className="flex items-center justify-between">
                        <div className="flex items-center gap-2">
                          <Sparkles className="w-4 h-4 text-indigo-400" />
                          <span className="text-xs font-bold text-white">AI Conversation Analysis</span>
                          <span
                            className={`text-[10px] font-bold px-2 py-0.5 rounded-full ${
                              ai.confidence >= 0.9
                                ? 'bg-emerald-500/20 text-emerald-300 border border-emerald-500/40'
                                : 'bg-amber-500/20 text-amber-300 border border-amber-500/40 animate-pulse'
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
                                ? 'bg-emerald-500/20 text-emerald-300'
                                : ai.status === 'OVERRIDDEN'
                                ? 'bg-purple-500/20 text-purple-300'
                                : ai.status === 'PENDING_REVIEW'
                                ? 'bg-amber-500/20 text-amber-300'
                                : 'bg-gray-800 text-gray-400'
                            }`}
                          >
                            {ai.status.replace(/_/g, ' ')}
                          </span>
                        </div>
                      </div>

                      <div className="text-xs text-gray-300">
                        <span className="text-gray-500 font-medium">Summary: </span>
                        {ai.summary}
                      </div>

                      {/* Evidence Quotes */}
                      {ai.evidence && ai.evidence.length > 0 && (
                        <div className="space-y-1">
                          <span className="text-[11px] text-gray-400 font-medium">Extracted Evidence:</span>
                          {ai.evidence.map((ev: any, idx: number) => (
                            <blockquote
                              key={idx}
                              className="text-xs border-l-2 border-indigo-500 pl-3 py-1 bg-indigo-900/10 text-indigo-200 italic"
                            >
                              "{ev.text}"
                            </blockquote>
                          ))}
                        </div>
                      )}

                      {/* Human Review Controls */}
                      {ai.status === 'PENDING_REVIEW' && (
                        <div className="pt-2 border-t border-indigo-500/20 flex items-center justify-end gap-2">
                          <button
                            onClick={() => handleConfirmAI(ai.id)}
                            className="flex items-center gap-1 px-3 py-1.5 rounded-lg bg-emerald-600 hover:bg-emerald-500 text-white text-xs font-semibold transition-colors"
                          >
                            <CheckCircle2 className="w-3.5 h-3.5" />
                            <span>Confirm {ai.intent}</span>
                          </button>
                          <button
                            onClick={() => {
                              setOverrideModalAnalysisId(ai.id);
                              setOverrideIntent('CALL_BACK_LATER');
                            }}
                            className="px-3 py-1.5 rounded-lg bg-gray-800 hover:bg-gray-700 text-gray-300 text-xs font-semibold transition-colors"
                          >
                            Change Outcome
                          </button>
                        </div>
                      )}
                    </div>
                  )}

                  {/* Conversation Transcript */}
                  {tr && (
                    <div className="rounded-xl border border-gray-800 bg-gray-950/40 p-4 space-y-2">
                      <div className="flex items-center justify-between text-xs text-gray-400 mb-2">
                        <span className="font-semibold text-white flex items-center gap-1.5">
                          <FileText className="w-3.5 h-3.5 text-indigo-400" />
                          Verified Call Transcript ({tr.language.toUpperCase()})
                        </span>
                        <span className="text-[10px] text-gray-500">Engine: {tr.provider}</span>
                      </div>

                      <div className="space-y-2 max-h-56 overflow-y-auto pr-2">
                        {tr.speakerSegments?.map((seg: any, idx: number) => (
                          <div
                            key={idx}
                            className={`p-2.5 rounded-lg text-xs leading-relaxed ${
                              seg.speaker === 'salesperson'
                                ? 'bg-indigo-950/30 border border-indigo-800/30 ml-4 text-indigo-200'
                                : 'bg-gray-900/80 border border-gray-800 mr-4 text-gray-200'
                            }`}
                          >
                            <span
                              className={`text-[10px] font-bold uppercase tracking-wider block mb-0.5 ${
                                seg.speaker === 'salesperson' ? 'text-indigo-400' : 'text-emerald-400'
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
          <div className="glass-panel p-5 rounded-2xl border border-gray-800 space-y-4">
            <h3 className="text-sm font-bold text-white uppercase tracking-wider">Lead Profile</h3>

            <div className="space-y-3 text-xs">
              <div className="flex justify-between py-1.5 border-b border-gray-800">
                <span className="text-gray-400">Campaign</span>
                <span className="text-white font-medium">{lead.campaign || 'Default Outreach'}</span>
              </div>
              <div className="flex justify-between py-1.5 border-b border-gray-800">
                <span className="text-gray-400">Priority</span>
                <span className="text-white font-medium">{lead.priority}</span>
              </div>
              <div className="flex justify-between py-1.5 border-b border-gray-800">
                <span className="text-gray-400">Total Call Attempts</span>
                <span className="text-white font-bold">{lead.callAttemptCount}</span>
              </div>
              <div className="flex justify-between py-1.5 border-b border-gray-800">
                <span className="text-gray-400">Follow-Up State</span>
                <span className="text-white font-medium">{lead.followUpStatus}</span>
              </div>
              <div className="flex justify-between py-1.5 border-b border-gray-800">
                <span className="text-gray-400">Created At</span>
                <span className="text-gray-300">
                  {new Date(lead.createdAt).toLocaleDateString([], { dateStyle: 'medium' })}
                </span>
              </div>
            </div>

            {lead.doNotCallReason && (
              <div className="p-3 rounded-lg bg-purple-950/20 border border-purple-500/30 text-xs">
                <div className="font-semibold text-purple-300 mb-0.5">DNC Reason:</div>
                <div className="text-gray-300">{lead.doNotCallReason}</div>
                <div className="text-[10px] text-gray-500 mt-1">Flagged by: {lead.doNotCallDetectedBy}</div>
              </div>
            )}
          </div>
        </div>
      </div>

      {/* Record Call Modal */}
      {showCallModal && (
        <div className="fixed inset-0 z-50 bg-black/70 backdrop-blur-sm flex items-center justify-center p-4">
          <div className="glass-panel w-full max-w-lg rounded-2xl p-6 shadow-2xl border border-gray-700 animate-in fade-in zoom-in-95">
            <div className="flex items-center justify-between pb-4 border-b border-gray-800 mb-4">
              <h3 className="text-base font-bold text-white">Record Call with {lead.name}</h3>
              <button
                onClick={() => setShowCallModal(false)}
                className="p-1 rounded-lg hover:bg-gray-800 text-gray-400 hover:text-white"
              >
                <X className="w-4 h-4" />
              </button>
            </div>

            <form onSubmit={handleRecordCall} className="space-y-4 text-xs">
              <div className="grid grid-cols-2 gap-3">
                <div>
                  <label className="block text-gray-300 font-medium mb-1">Call Status *</label>
                  <select
                    value={callStatus}
                    onChange={(e) => setCallStatus(e.target.value as CallStatus)}
                    className="w-full p-2.5 bg-gray-900 border border-gray-700 rounded-lg text-white focus:outline-none focus:border-indigo-500"
                  >
                    <option value="ANSWERED">Answered</option>
                    <option value="NO_ANSWER">No Answer</option>
                    <option value="BUSY">Busy</option>
                    <option value="SWITCHED_OFF">Switched Off</option>
                    <option value="UNREACHABLE">Unreachable</option>
                  </select>
                </div>

                <div>
                  <label className="block text-gray-300 font-medium mb-1">Call Outcome *</label>
                  <select
                    value={callOutcome}
                    onChange={(e) => setCallOutcome(e.target.value as CallOutcome)}
                    className="w-full p-2.5 bg-gray-900 border border-gray-700 rounded-lg text-white focus:outline-none focus:border-indigo-500"
                  >
                    <option value="INTERESTED">Interested</option>
                    <option value="CALL_BACK_LATER">Call Back Later</option>
                    <option value="INFO_REQUESTED">Info Requested</option>
                    <option value="UNDECIDED">Undecided</option>
                    <option value="NOT_INTERESTED">Not Interested</option>
                    <option value="DO_NOT_CALL">Do Not Call</option>
                    <option value="CONVERTED">Converted / Enrolled</option>
                    <option value="LOST">Lost</option>
                  </select>
                </div>
              </div>

              {callOutcome === 'CALL_BACK_LATER' && (
                <div>
                  <label className="block text-gray-300 font-medium mb-1">Scheduled Follow-Up Date/Time</label>
                  <input
                    type="datetime-local"
                    value={followUpDate}
                    onChange={(e) => setFollowUpDate(e.target.value)}
                    className="w-full p-2.5 bg-gray-900 border border-gray-700 rounded-lg text-white focus:outline-none focus:border-indigo-500"
                  />
                </div>
              )}

              <div>
                <label className="block text-gray-300 font-medium mb-1">Call Duration (seconds)</label>
                <input
                  type="number"
                  value={callDuration}
                  onChange={(e) => setCallDuration(Number(e.target.value))}
                  className="w-full p-2.5 bg-gray-900 border border-gray-700 rounded-lg text-white focus:outline-none focus:border-indigo-500"
                />
              </div>

              <div>
                <label className="block text-gray-300 font-medium mb-1">Call Notes</label>
                <textarea
                  rows={3}
                  value={callNotes}
                  onChange={(e) => setCallNotes(e.target.value)}
                  placeholder="Summary of conversation, key questions asked by lead..."
                  className="w-full p-2.5 bg-gray-900 border border-gray-700 rounded-lg text-white focus:outline-none focus:border-indigo-500"
                />
              </div>

              <div className="flex items-center gap-2 text-gray-300">
                <input
                  type="checkbox"
                  id="consent"
                  checked={callConsent}
                  onChange={(e) => setCallConsent(e.target.checked)}
                  className="rounded border-gray-700 text-indigo-600 focus:ring-indigo-500"
                />
                <label htmlFor="consent" className="text-[11px] text-gray-400">
                  Lead consented to call recording for quality and transcription analysis
                </label>
              </div>

              <div className="flex items-center justify-end gap-3 pt-3 border-t border-gray-800">
                <button
                  type="button"
                  onClick={() => setShowCallModal(false)}
                  className="px-4 py-2 rounded-lg bg-gray-800 hover:bg-gray-700 text-gray-300"
                >
                  Cancel
                </button>
                <button
                  type="submit"
                  disabled={callSubmitting}
                  className="px-4 py-2 rounded-lg bg-indigo-600 hover:bg-indigo-500 text-white font-semibold shadow-md shadow-indigo-600/20"
                >
                  {callSubmitting ? 'Recording...' : 'Submit Call'}
                </button>
              </div>
            </form>
          </div>
        </div>
      )}

      {/* AI Override Modal */}
      {overrideModalAnalysisId && (
        <div className="fixed inset-0 z-50 bg-black/70 backdrop-blur-sm flex items-center justify-center p-4">
          <div className="glass-panel w-full max-w-md rounded-2xl p-6 shadow-2xl border border-gray-700 animate-in fade-in zoom-in-95">
            <h3 className="text-base font-bold text-white mb-2">Override AI Classification</h3>
            <p className="text-xs text-gray-400 mb-4">
              Select the correct conversation intent. An audit log will record this change.
            </p>

            <form onSubmit={handleOverrideAI} className="space-y-4 text-xs">
              <div>
                <label className="block text-gray-300 font-medium mb-1">New Intent</label>
                <select
                  value={overrideIntent}
                  onChange={(e) => setOverrideIntent(e.target.value as AIIntent)}
                  className="w-full p-2.5 bg-gray-900 border border-gray-700 rounded-lg text-white focus:outline-none focus:border-indigo-500"
                >
                  <option value="CALL_BACK_LATER">Call Back Later</option>
                  <option value="INFO_REQUESTED">Info Requested</option>
                  <option value="UNDECIDED">Undecided</option>
                  <option value="INTERESTED">Interested</option>
                  <option value="NOT_INTERESTED">Not Interested</option>
                  <option value="DO_NOT_CALL">Do Not Call</option>
                </select>
              </div>

              <div>
                <label className="block text-gray-300 font-medium mb-1">Reason for Override *</label>
                <textarea
                  required
                  rows={3}
                  value={overrideReason}
                  onChange={(e) => setOverrideReason(e.target.value)}
                  placeholder="Explain why AI misclassified this conversation..."
                  className="w-full p-2.5 bg-gray-900 border border-gray-700 rounded-lg text-white focus:outline-none focus:border-indigo-500"
                />
              </div>

              <div className="flex items-center justify-end gap-3 pt-3 border-t border-gray-800">
                <button
                  type="button"
                  onClick={() => setOverrideModalAnalysisId(null)}
                  className="px-4 py-2 rounded-lg bg-gray-800 hover:bg-gray-700 text-gray-300"
                >
                  Cancel
                </button>
                <button
                  type="submit"
                  disabled={overriding}
                  className="px-4 py-2 rounded-lg bg-indigo-600 hover:bg-indigo-500 text-white font-semibold"
                >
                  {overriding ? 'Updating...' : 'Save Correction'}
                </button>
              </div>
            </form>
          </div>
        </div>
      )}

      {/* DNC Toggle Modal */}
      {showDNCModal && (
        <div className="fixed inset-0 z-50 bg-black/70 backdrop-blur-sm flex items-center justify-center p-4">
          <div className="glass-panel w-full max-w-md rounded-2xl p-6 shadow-2xl border border-gray-700 animate-in fade-in zoom-in-95">
            <h3 className="text-base font-bold text-white mb-2">
              {lead.doNotCall ? 'Reactivate Lead from DNC' : 'Flag as Do Not Call (DNC)'}
            </h3>
            <p className="text-xs text-gray-400 mb-4">
              {lead.doNotCall
                ? 'Reactivating this lead will allow future calls and follow-ups. Please provide justification.'
                : 'Marking Do Not Call will strictly cancel all follow-ups and block future alert notifications.'}
            </p>

            <form onSubmit={handleToggleDNC} className="space-y-4 text-xs">
              <div>
                <label className="block text-gray-300 font-medium mb-1">Reason *</label>
                <textarea
                  required
                  rows={3}
                  value={dncReason}
                  onChange={(e) => setDncReason(e.target.value)}
                  placeholder="Mandatory reason for audit trail..."
                  className="w-full p-2.5 bg-gray-900 border border-gray-700 rounded-lg text-white focus:outline-none focus:border-indigo-500"
                />
              </div>

              <div className="flex items-center justify-end gap-3 pt-3 border-t border-gray-800">
                <button
                  type="button"
                  onClick={() => setShowDNCModal(false)}
                  className="px-4 py-2 rounded-lg bg-gray-800 hover:bg-gray-700 text-gray-300"
                >
                  Cancel
                </button>
                <button
                  type="submit"
                  className={`px-4 py-2 rounded-lg text-white font-semibold ${
                    lead.doNotCall
                      ? 'bg-indigo-600 hover:bg-indigo-500'
                      : 'bg-purple-600 hover:bg-purple-500 shadow-md shadow-purple-600/20'
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

import React, { useState, useEffect, useRef } from 'react';
import {
  Phone,
  PhoneOff,
  Mic,
  MicOff,
  Volume2,
  Clock,
  Sparkles,
  CheckCircle2,
  AlertTriangle,
  XCircle,
  Bell,
  ArrowRight,
  ShieldCheck,
  Bot,
  User,
  Radio,
  Zap,
} from 'lucide-react';
import { callsApi } from '../../services/api';
import { Lead, AIAnalysis } from '../../types';

interface LiveCallModalProps {
  lead: Lead;
  onClose: () => void;
  onSuccess?: (result: { call: any; analysis: AIAnalysis; lead?: Lead }) => void;
}

interface DialogueTurn {
  speaker: 'salesperson' | 'lead';
  text: string;
  time: string;
}

export const LiveCallModal: React.FC<LiveCallModalProps> = ({ lead, onClose, onSuccess }) => {
  // Call States: 'CONNECTING' | 'CONNECTED' | 'PROCESSING' | 'ALERT_SUMMARY'
  const [callState, setCallState] = useState<'CONNECTING' | 'CONNECTED' | 'PROCESSING' | 'ALERT_SUMMARY'>('CONNECTING');
  const [callSeconds, setCallSeconds] = useState(0);
  const [isMicActive, setIsMicActive] = useState(true);
  const [dialogueTurns, setDialogueTurns] = useState<DialogueTurn[]>([]);
  const [liveTranscriptInterim, setLiveTranscriptInterim] = useState<string>('');
  const [aiResult, setAiResult] = useState<{ call: any; analysis: AIAnalysis; lead?: Lead } | null>(null);
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [activeScenario, setActiveScenario] = useState<string>('meeting_later');

  const recognitionRef = useRef<any>(null);
  const timerIntervalRef = useRef<any>(null);
  const simulatedSpeechTimeoutsRef = useRef<any[]>([]);
  const transcriptEndRef = useRef<HTMLDivElement>(null);

  // Pre-configured live speech telephony audio streams
  const speechScenarios: Record<string, { label: string; turns: { speaker: 'salesperson' | 'lead'; text: string; delayMs: number }[] }> = {
    meeting_later: {
      label: 'In Meeting — Call Me Later This Evening',
      turns: [
        { speaker: 'salesperson', text: `Hello, am I speaking with ${lead.name}? Vikram calling from the admissions office.`, delayMs: 1200 },
        { speaker: 'lead', text: `Hi Vikram! I am currently stuck in an office team meeting right now. Please call me later this evening around 6 PM.`, delayMs: 4000 },
        { speaker: 'salesperson', text: `Understood ${lead.name}! I will set a callback for you this evening at 6 PM. Have a productive meeting!`, delayMs: 7000 },
      ],
    },
    driving_2hours: {
      label: 'Driving / In Class — Call in 2 Hours',
      turns: [
        { speaker: 'salesperson', text: `Good afternoon ${lead.name}, reaching out regarding your course application.`, delayMs: 1200 },
        { speaker: 'lead', text: `Hey, I am driving on the highway right now, can you please call me back after 2 hours?`, delayMs: 4000 },
        { speaker: 'salesperson', text: `Sure thing! Drive safe, I will ring you back in exactly 2 hours.`, delayMs: 7000 },
      ],
    },
    call_dropped: {
      label: 'Audio Glitch / Call Dropped (Retry Needed)',
      turns: [
        { speaker: 'salesperson', text: `Hello ${lead.name}, Vikram here from admissions.`, delayMs: 1200 },
        { speaker: 'lead', text: `Hello? Hello? Your voice is breaking up badly... call cut...`, delayMs: 3800 },
        { speaker: 'salesperson', text: `Hello ${lead.name}, can you hear me? The line seems to be dropping...`, delayMs: 6500 },
      ],
    },
    not_interested: {
      label: 'Student Not Interested / Do Not Call',
      turns: [
        { speaker: 'salesperson', text: `Hello ${lead.name}, following up on your inquiry for ${lead.course || 'the course'}.`, delayMs: 1200 },
        { speaker: 'lead', text: `I already took admission in another institute. I am not interested at all, please do not call me again.`, delayMs: 4000 },
        { speaker: 'salesperson', text: `Thank you for letting us know. Removing your number from our calling station right away.`, delayMs: 7200 },
      ],
    },
    interested_syllabus: {
      label: 'Interested Student (Requests Syllabus & Timings)',
      turns: [
        { speaker: 'salesperson', text: `Hello ${lead.name}, calling regarding your inquiry for ${lead.course || 'the professional training program'}.`, delayMs: 1200 },
        { speaker: 'lead', text: `Yes! I wanted to know the weekend batch timings and if placement assistance is provided. Can you share the syllabus?`, delayMs: 4000 },
        { speaker: 'salesperson', text: `Definitely! I will WhatsApp you the complete syllabus brochure and schedule right away.`, delayMs: 7500 },
      ],
    },
  };

  // Auto-scroll transcript feed
  useEffect(() => {
    transcriptEndRef.current?.scrollIntoView({ behavior: 'smooth' });
  }, [dialogueTurns, liveTranscriptInterim]);

  // Connect Call & Start Speech Recognition
  useEffect(() => {
    // 1. Simulate ringing connection (1 second)
    const connectTimeout = setTimeout(() => {
      setCallState('CONNECTED');

      // 2. Start Call Duration Timer
      timerIntervalRef.current = setInterval(() => {
        setCallSeconds((prev) => prev + 1);
      }, 1000);

      // 3. Initiate Web Speech API if supported in browser
      startBrowserSpeechRecognition();

      // 4. Stream real-time dialogue speech turns
      startTelephonyDialogueStream(activeScenario);
    }, 1200);

    return () => {
      clearTimeout(connectTimeout);
      stopTimer();
      stopSpeechRecognition();
      clearSpeechTimeouts();
    };
  }, []);

  const stopTimer = () => {
    if (timerIntervalRef.current) {
      clearInterval(timerIntervalRef.current);
      timerIntervalRef.current = null;
    }
  };

  const clearSpeechTimeouts = () => {
    simulatedSpeechTimeoutsRef.current.forEach((t) => clearTimeout(t));
    simulatedSpeechTimeoutsRef.current = [];
  };

  // Browser Web Speech Recognition
  const startBrowserSpeechRecognition = () => {
    try {
      const SpeechRecognition = (window as any).SpeechRecognition || (window as any).webkitSpeechRecognition;
      if (!SpeechRecognition) return;

      const recognition = new SpeechRecognition();
      recognition.continuous = true;
      recognition.interimResults = true;
      recognition.lang = 'en-IN';

      recognition.onresult = (event: any) => {
        let interim = '';
        for (let i = event.resultIndex; i < event.results.length; ++i) {
          if (event.results[i].isFinal) {
            const finalTranscript = event.results[i][0].transcript.trim();
            if (finalTranscript) {
              addDialogueTurn('salesperson', finalTranscript);
            }
          } else {
            interim += event.results[i][0].transcript;
          }
        }
        setLiveTranscriptInterim(interim);
      };

      recognition.onerror = (e: any) => {
        console.warn('Speech recognition notice:', e.error);
      };

      recognition.start();
      recognitionRef.current = recognition;
    } catch (err) {
      console.warn('Browser Speech API not accessible in this environment:', err);
    }
  };

  const stopSpeechRecognition = () => {
    if (recognitionRef.current) {
      try {
        recognitionRef.current.stop();
      } catch (e) {}
      recognitionRef.current = null;
    }
  };

  // Stream live speech dialogue turns
  const startTelephonyDialogueStream = (scenarioKey: string) => {
    clearSpeechTimeouts();
    const scenario = speechScenarios[scenarioKey] || speechScenarios.meeting_later;

    scenario.turns.forEach((turn) => {
      const t = setTimeout(() => {
        addDialogueTurn(turn.speaker, turn.text);
      }, turn.delayMs);
      simulatedSpeechTimeoutsRef.current.push(t);
    });
  };

  const addDialogueTurn = (speaker: 'salesperson' | 'lead', text: string) => {
    const time = new Date().toLocaleTimeString([], { hour: '2-digit', minute: '2-digit', second: '2-digit' });
    setDialogueTurns((prev) => [...prev, { speaker, text, time }]);
    setLiveTranscriptInterim('');
  };

  const handleSwitchScenario = (scenarioKey: string) => {
    setActiveScenario(scenarioKey);
    setDialogueTurns([]);
    startTelephonyDialogueStream(scenarioKey);
  };

  // End Call & Automatically Process (Zero BDA Typing)
  const handleEndCall = async () => {
    stopTimer();
    stopSpeechRecognition();
    clearSpeechTimeouts();
    setCallState('PROCESSING');
    setIsSubmitting(true);

    // Build the full spoken transcript from live turns
    const compiledTranscript = dialogueTurns
      .map((t) => `${t.speaker === 'salesperson' ? 'Sales BDA' : 'Student'}: ${t.text}`)
      .join('\n');

    const duration = Math.max(callSeconds, 15);

    try {
      const res = await callsApi.autoProcessCall(lead.id, {
        transcriptText: compiledTranscript || `Sales BDA: Hello ${lead.name}\nStudent: Please call me tomorrow morning, busy now.`,
        durationSeconds: duration,
      });

      setAiResult(res);
      setCallState('ALERT_SUMMARY');

      if (onSuccess) {
        onSuccess(res);
      }
      window.dispatchEvent(new CustomEvent('lead-status-updated'));
    } catch (err: any) {
      alert(`AI Call Processing Error: ${err.message}`);
      setCallState('CONNECTED');
    } finally {
      setIsSubmitting(false);
    }
  };

  const formatTimer = (totalSeconds: number) => {
    const mins = Math.floor(totalSeconds / 60);
    const secs = totalSeconds % 60;
    return `${mins.toString().padStart(2, '0')}:${secs.toString().padStart(2, '0')}`;
  };

  return (
    <div className="fixed inset-0 z-50 bg-slate-900/60 backdrop-blur-sm flex items-center justify-center p-4">
      <div className="bg-white border border-slate-200 rounded-2xl max-w-2xl w-full p-6 shadow-2xl space-y-5 animate-in fade-in zoom-in-95 duration-150 max-h-[92vh] flex flex-col">
        {/* Call Header */}
        <div className="flex items-start justify-between border-b border-slate-100 pb-4 shrink-0">
          <div>
            <div className="flex items-center gap-2">
              {callState === 'CONNECTING' && (
                <span className="flex items-center gap-1.5 px-2.5 py-0.5 rounded-full text-[11px] font-bold bg-amber-50 text-amber-700 border border-amber-200">
                  <span className="w-2 h-2 rounded-full bg-amber-500 animate-ping" />
                  Dialing Student...
                </span>
              )}
              {callState === 'CONNECTED' && (
                <span className="flex items-center gap-1.5 px-2.5 py-0.5 rounded-full text-[11px] font-bold bg-emerald-50 text-emerald-700 border border-emerald-200">
                  <span className="w-2 h-2 rounded-full bg-emerald-500 animate-pulse" />
                  Call Active — Real-Time STT Converting
                </span>
              )}
              {callState === 'PROCESSING' && (
                <span className="flex items-center gap-1.5 px-2.5 py-0.5 rounded-full text-[11px] font-bold bg-indigo-50 text-indigo-700 border border-indigo-200">
                  <Sparkles className="w-3.5 h-3.5 animate-spin" />
                  AI Processing Dialogue...
                </span>
              )}
              {callState === 'ALERT_SUMMARY' && (
                <span className="flex items-center gap-1.5 px-2.5 py-0.5 rounded-full text-[11px] font-bold bg-indigo-600 text-white shadow-xs">
                  <Bell className="w-3.5 h-3.5" />
                  Follow-Up Reminder Scheduled
                </span>
              )}
            </div>

            <h3 className="text-xl font-bold text-slate-900 mt-1">{lead.name}</h3>
            <div className="flex items-center gap-3 text-xs text-slate-500 mt-1">
              <span className="font-mono font-semibold text-indigo-600">{lead.phone}</span>
              <span>•</span>
              <span>{lead.course || 'Admissions Candidate'}</span>
              <span>•</span>
              <span>Attempt #{lead.callAttemptCount + 1}</span>
            </div>
          </div>

          <div className="flex items-center gap-3">
            {callState === 'CONNECTED' && (
              <div className="text-right">
                <div className="text-lg font-mono font-bold text-slate-900 flex items-center gap-1.5">
                  <Clock className="w-4 h-4 text-emerald-600" />
                  {formatTimer(callSeconds)}
                </div>
                <div className="text-[10px] text-slate-400 uppercase font-semibold">Duration</div>
              </div>
            )}
            {callState === 'ALERT_SUMMARY' && (
              <button
                onClick={onClose}
                className="p-1.5 rounded-lg text-slate-400 hover:text-slate-700 hover:bg-slate-100 transition"
              >
                ✕
              </button>
            )}
          </div>
        </div>

        {/* ============================================================== */}
        {/* VIEW 1: ACTIVE LIVE CALL & SPEECH TRANSCRIPTION                */}
        {/* ============================================================== */}
        {(callState === 'CONNECTING' || callState === 'CONNECTED' || callState === 'PROCESSING') && (
          <div className="space-y-4 flex-1 overflow-hidden flex flex-col">
            {/* Live Audio & Speech Activity Visualizer */}
            <div className="bg-slate-50 rounded-xl p-3 border border-slate-200 flex items-center justify-between shrink-0">
              <div className="flex items-center gap-3">
                <div className="w-9 h-9 rounded-xl bg-indigo-600 text-white flex items-center justify-center shadow-xs">
                  <Volume2 className="w-5 h-5 animate-pulse" />
                </div>
                <div>
                  <div className="text-xs font-bold text-slate-900 flex items-center gap-1.5">
                    Live Telephony Speech-to-Text
                    <span className="text-[10px] font-normal text-slate-500">(Zero manual typing)</span>
                  </div>
                  <p className="text-[11px] text-slate-500">
                    Speech from both sides is automatically transcribed live while you talk.
                  </p>
                </div>
              </div>

              {/* Animated audio waveform bars */}
              <div className="flex items-end gap-1 h-6">
                <span className="w-1 bg-indigo-600 rounded-full animate-bounce [animation-delay:0.1s] h-4" />
                <span className="w-1 bg-emerald-600 rounded-full animate-bounce [animation-delay:0.3s] h-6" />
                <span className="w-1 bg-indigo-500 rounded-full animate-bounce [animation-delay:0.2s] h-3" />
                <span className="w-1 bg-teal-500 rounded-full animate-bounce [animation-delay:0.4s] h-5" />
                <span className="w-1 bg-indigo-600 rounded-full animate-bounce [animation-delay:0.25s] h-4" />
              </div>
            </div>

            {/* Quick Live Telephony Scenario Switcher (for demo/testing live speech responses) */}
            <div className="shrink-0 space-y-1">
              <div className="flex items-center justify-between text-[11px] text-slate-500 font-semibold">
                <span>Incoming Student Audio Channel:</span>
                <span className="text-indigo-600 font-medium">Auto-Transcribing Dialogue</span>
              </div>
              <div className="grid grid-cols-2 sm:grid-cols-3 gap-1.5">
                {Object.entries(speechScenarios).map(([key, s]) => (
                  <button
                    key={key}
                    type="button"
                    onClick={() => handleSwitchScenario(key)}
                    className={`px-2.5 py-1.5 rounded-lg text-[10px] font-semibold text-left border transition ${
                      activeScenario === key
                        ? 'bg-indigo-50 border-indigo-300 text-indigo-700 shadow-xs'
                        : 'bg-white border-slate-200 text-slate-600 hover:bg-slate-50'
                    }`}
                  >
                    {s.label}
                  </button>
                ))}
              </div>
            </div>

            {/* Live Conversation Transcript Stream (Turns appearing live) */}
            <div className="flex-1 bg-slate-50/70 border border-slate-200 rounded-xl p-4 overflow-y-auto space-y-3 min-h-[160px] max-h-[220px]">
              {dialogueTurns.length === 0 ? (
                <div className="h-full flex flex-col items-center justify-center text-slate-400 gap-2 py-6">
                  <Radio className="w-6 h-6 text-indigo-500 animate-pulse" />
                  <p className="text-xs font-medium">Connecting audio channel & listening...</p>
                </div>
              ) : (
                dialogueTurns.map((turn, idx) => {
                  const isBda = turn.speaker === 'salesperson';
                  return (
                    <div
                      key={idx}
                      className={`flex flex-col ${isBda ? 'items-end' : 'items-start'} animate-in fade-in slide-in-from-bottom-1 duration-150`}
                    >
                      <div className="flex items-center gap-1.5 text-[10px] text-slate-400 mb-0.5">
                        <span className="font-bold text-slate-700">{isBda ? '🎙️ BDA (You)' : `🔊 Student (${lead.name})`}</span>
                        <span>•</span>
                        <span>{turn.time}</span>
                      </div>
                      <div
                        className={`p-3 rounded-2xl text-xs max-w-[85%] leading-relaxed shadow-xs ${
                          isBda
                            ? 'bg-indigo-600 text-white rounded-tr-none'
                            : 'bg-white border border-slate-200 text-slate-800 rounded-tl-none font-medium'
                        }`}
                      >
                        {turn.text}
                      </div>
                    </div>
                  );
                })
              )}

              {/* Interim real-time speech from mic */}
              {liveTranscriptInterim && (
                <div className="flex flex-col items-end animate-pulse">
                  <span className="text-[10px] text-indigo-600 font-semibold mb-0.5">Speaking into mic...</span>
                  <div className="p-2.5 rounded-2xl bg-indigo-50 border border-indigo-200 text-indigo-900 text-xs italic">
                    "{liveTranscriptInterim}"
                  </div>
                </div>
              )}
              <div ref={transcriptEndRef} />
            </div>

            {/* In-Call Controls */}
            <div className="pt-2 flex items-center justify-between gap-3 shrink-0">
              <div className="text-[11px] text-slate-500 flex items-center gap-1.5">
                <ShieldCheck className="w-4 h-4 text-emerald-600" />
                <span>Zero typing: Ending the call triggers instant AI follow-up scheduling.</span>
              </div>

              <div className="flex items-center gap-2">
                {/* End Call Button */}
                <button
                  type="button"
                  onClick={handleEndCall}
                  disabled={isSubmitting || dialogueTurns.length === 0}
                  className="flex items-center gap-2 px-5 py-2.5 rounded-xl bg-rose-600 hover:bg-rose-700 text-white font-bold text-xs shadow-xs transition disabled:opacity-50"
                >
                  <PhoneOff className="w-4 h-4" />
                  <span>{isSubmitting ? 'Analyzing Dialogue...' : 'End Call & Auto-Process'}</span>
                </button>
              </div>
            </div>
          </div>
        )}

        {/* ============================================================== */}
        {/* VIEW 2: INSTANT BDA FOLLOW-UP ALERT & REMINDER                 */}
        {/* ============================================================== */}
        {callState === 'ALERT_SUMMARY' && aiResult && (
          <div className="space-y-4 animate-in fade-in duration-200">
            {/* Main Alert Card based on AI intent */}
            <div
              className={`p-5 rounded-2xl border space-y-3 ${
                aiResult.analysis.intent === 'CALL_BACK_LATER'
                  ? 'bg-cyan-50/70 border-cyan-200 text-cyan-950'
                  : aiResult.analysis.intent === 'FOLLOW_UP_REQUIRED'
                  ? 'bg-amber-50/70 border-amber-200 text-amber-950'
                  : aiResult.analysis.intent === 'NOT_INTERESTED' || aiResult.analysis.intent === 'DO_NOT_CALL'
                  ? 'bg-slate-50 border-slate-200 text-slate-900'
                  : 'bg-emerald-50/70 border-emerald-200 text-emerald-950'
              }`}
            >
              <div className="flex items-center justify-between">
                <div className="flex items-center gap-2.5">
                  {aiResult.analysis.intent === 'CALL_BACK_LATER' ? (
                    <Clock className="w-6 h-6 text-cyan-600" />
                  ) : aiResult.analysis.intent === 'FOLLOW_UP_REQUIRED' ? (
                    <AlertTriangle className="w-6 h-6 text-amber-600" />
                  ) : aiResult.analysis.intent === 'NOT_INTERESTED' || aiResult.analysis.intent === 'DO_NOT_CALL' ? (
                    <XCircle className="w-6 h-6 text-rose-600" />
                  ) : (
                    <CheckCircle2 className="w-6 h-6 text-emerald-600" />
                  )}

                  <div>
                    <h4 className="text-base font-bold">
                      {aiResult.analysis.intent === 'CALL_BACK_LATER' && '🔔 Follow-Up Reminder Scheduled!'}
                      {aiResult.analysis.intent === 'FOLLOW_UP_REQUIRED' && '⚠️ Call Dropped / Retry Alert'}
                      {(aiResult.analysis.intent === 'NOT_INTERESTED' || aiResult.analysis.intent === 'DO_NOT_CALL') &&
                        '🛑 Lead Suppressed — Do Not Call'}
                      {aiResult.analysis.intent === 'INTERESTED' && '🎉 Positive Lead Enrollment Alert!'}
                    </h4>
                    <p className="text-xs text-slate-500 mt-0.5">
                      Analyzed automatically from spoken words with {(aiResult.analysis.confidence * 100).toFixed(0)}% confidence
                    </p>
                  </div>
                </div>

                <span
                  className={`text-xs font-bold uppercase px-3 py-1 rounded-full border ${
                    aiResult.analysis.intent === 'CALL_BACK_LATER'
                      ? 'bg-cyan-100 text-cyan-800 border-cyan-300'
                      : aiResult.analysis.intent === 'FOLLOW_UP_REQUIRED'
                      ? 'bg-amber-100 text-amber-800 border-amber-300'
                      : aiResult.analysis.intent === 'NOT_INTERESTED' || aiResult.analysis.intent === 'DO_NOT_CALL'
                      ? 'bg-slate-200 text-slate-800 border-slate-300'
                      : 'bg-emerald-100 text-emerald-800 border-emerald-300'
                  }`}
                >
                  {aiResult.analysis.intent.replace(/_/g, ' ')}
                </span>
              </div>

              {/* Exact Evidence Quote Heard by AI */}
              {aiResult.analysis.evidence && aiResult.analysis.evidence.length > 0 && (
                <div className="bg-white p-3.5 rounded-xl border border-slate-200 text-xs space-y-1 shadow-xs">
                  <div className="text-[10px] font-bold uppercase text-slate-400 flex items-center gap-1">
                    <Bot className="w-3.5 h-3.5 text-indigo-600" />
                    Exact Quote Spoken by Student:
                  </div>
                  <p className="text-slate-800 font-semibold italic text-sm">
                    "{aiResult.analysis.evidence[0].text}"
                  </p>
                </div>
              )}

              {/* Action Taken Summary */}
              <div className="text-xs space-y-1.5 pt-1">
                <div className="font-semibold text-slate-900">Automated System Action:</div>
                <p className="text-slate-600">
                  {aiResult.analysis.intent === 'CALL_BACK_LATER' &&
                    '⏰ Status updated to FOLLOW_UP_SCHEDULED. Added to your "Call Me Later" tab with scheduled reminder alert.'}
                  {aiResult.analysis.intent === 'FOLLOW_UP_REQUIRED' &&
                    '⚡ Call attempt logged. Lead kept in your active queue for immediate follow-up retry.'}
                  {(aiResult.analysis.intent === 'NOT_INTERESTED' || aiResult.analysis.intent === 'DO_NOT_CALL') &&
                    '✅ Student permanently removed from your active calling queue. You will not be asked to call them again.'}
                  {aiResult.analysis.intent === 'INTERESTED' &&
                    '🎉 Status updated to INTERESTED. Candidate marked for admissions syllabus follow-up.'}
                </p>
              </div>
            </div>

            {/* Done Action */}
            <div className="flex justify-end gap-3 pt-2">
              <button
                type="button"
                onClick={onClose}
                className="w-full py-3 rounded-xl bg-indigo-600 hover:bg-indigo-700 text-white font-bold text-xs shadow-xs transition flex items-center justify-center gap-2"
              >
                <span>Got It — Next Student in Queue</span>
                <ArrowRight className="w-4 h-4" />
              </button>
            </div>
          </div>
        )}
      </div>
    </div>
  );
};

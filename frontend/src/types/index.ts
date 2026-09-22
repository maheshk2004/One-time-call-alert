export type Role = 'ADMIN' | 'MANAGER' | 'SALES';

export interface User {
  id: string;
  name: string;
  email: string;
  role: Role;
  teamId?: string;
  phone?: string;
  isActive: boolean;
}

export type LeadStatus =
  | 'NEW'
  | 'ASSIGNED'
  | 'CALL_PENDING'
  | 'CALLED_ONCE'
  | 'FOLLOW_UP_SCHEDULED'
  | 'FOLLOW_UP_REQUIRED'
  | 'CONTACTED'
  | 'INTERESTED'
  | 'UNDECIDED'
  | 'INFO_REQUESTED'
  | 'NOT_INTERESTED'
  | 'DO_NOT_CALL'
  | 'CONVERTED'
  | 'LOST'
  | 'DISQUALIFIED';

export type Priority = 'LOW' | 'MEDIUM' | 'HIGH' | 'URGENT';
export type FollowUpStatus = 'NONE' | 'PENDING' | 'SCHEDULED' | 'OVERDUE' | 'COMPLETED';

export interface Lead {
  id: string;
  name: string;
  email: string;
  phone: string;
  alternatePhone?: string;
  source: string;
  campaign?: string;
  adName?: string;
  course?: string;
  location?: string;
  assignedTo?: string;
  assignedUser?: {
    id: string;
    name: string;
    email: string;
    role: Role;
  };
  status: LeadStatus;
  priority: Priority;
  callAttemptCount: number;
  firstCallAt?: string;
  lastCallAt?: string;
  nextFollowUpAt?: string;
  followUpStatus: FollowUpStatus;
  doNotCall: boolean;
  doNotCallReason?: string;
  doNotCallDetectedBy?: string;
  doNotCallAt?: string;
  notes?: string;
  createdAt: string;
  updatedAt: string;
}

export type CallStatus =
  | 'INITIATED'
  | 'RINGING'
  | 'ANSWERED'
  | 'NO_ANSWER'
  | 'BUSY'
  | 'SWITCHED_OFF'
  | 'UNREACHABLE'
  | 'FAILED';

export type CallOutcome =
  | 'INTERESTED'
  | 'NOT_INTERESTED'
  | 'CALL_BACK_LATER'
  | 'INFO_REQUESTED'
  | 'UNDECIDED'
  | 'FOLLOW_UP_REQUIRED'
  | 'FOLLOW_UP_SCHEDULED'
  | 'CONVERTED'
  | 'LOST'
  | 'WRONG_NUMBER'
  | 'DO_NOT_CALL';

export interface CallAttempt {
  id: string;
  leadId: string;
  salespersonId: string;
  salespersonUser?: {
    id: string;
    name: string;
    email: string;
    role: Role;
  };
  attemptNumber: number;
  callStartedAt: string;
  callEndedAt?: string;
  durationSeconds: number;
  callStatus: CallStatus;
  callOutcome: CallOutcome;
  recordingConsent: boolean;
  recordingAvailable: boolean;
  recordingUrl?: string;
  recordingProvider?: string;
  transcriptAvailable: boolean;
  transcriptId?: string;
  aiAnalysisId?: string;
  manuallyReviewed: boolean;
  notes?: string;
  createdAt: string;
  updatedAt: string;
}

export interface SpeakerSegment {
  speaker: string;
  text: string;
  timestampMs?: number;
}

export interface Transcript {
  id: string;
  leadId: string;
  callAttemptId: string;
  language: string;
  transcript: string;
  speakerSegments: SpeakerSegment[];
  durationSeconds: number;
  provider: string;
  createdAt: string;
}

export type AIIntent =
  | 'INTERESTED'
  | 'NOT_INTERESTED'
  | 'CALL_BACK_LATER'
  | 'INFO_REQUESTED'
  | 'UNDECIDED'
  | 'FOLLOW_UP_REQUIRED'
  | 'CONVERTED'
  | 'LOST'
  | 'DO_NOT_CALL'
  | 'WRONG_NUMBER';

export type Sentiment = 'POSITIVE' | 'NEUTRAL' | 'NEGATIVE';
export type AIReviewStatus = 'AUTO_APPLIED' | 'PENDING_REVIEW' | 'CONFIRMED' | 'OVERRIDDEN';

export interface EvidenceItem {
  speaker: string;
  text: string;
}

export interface AIAnalysis {
  id: string;
  callAttemptId: string;
  leadId: string;
  transcriptId: string;
  intent: AIIntent;
  confidence: number;
  sentiment: Sentiment;
  followUpRequired: boolean;
  doNotCall: boolean;
  suggestedFollowUpHours?: number;
  summary: string;
  reason: string;
  evidence: EvidenceItem[];
  status: AIReviewStatus;
  originalIntent?: AIIntent;
  reviewedBy?: string;
  reviewedByUser?: {
    id: string;
    name: string;
    email: string;
    role: Role;
  };
  reviewedAt?: string;
  overrideReason?: string;
  createdAt: string;
  updatedAt: string;
}

export type FollowUpTaskStatus = 'PENDING' | 'SCHEDULED' | 'OVERDUE' | 'COMPLETED' | 'CANCELLED';

export interface FollowUp {
  id: string;
  leadId: string;
  leadName?: string;
  leadPhone?: string;
  salespersonId: string;
  salespersonUser?: {
    id: string;
    name: string;
    email: string;
    role: Role;
  };
  dueAt: string;
  scheduledAt: string;
  status: FollowUpTaskStatus;
  notes?: string;
  completedAt?: string;
  completedByCallAttemptId?: string;
  createdAt: string;
  updatedAt: string;
}

export type NotificationType =
  | 'ONE_CALL_FOLLOWUP_REQUIRED'
  | 'FOLLOWUP_OVERDUE'
  | 'AI_REVIEW_NEEDED'
  | 'LEAD_ASSIGNED'
  | 'DO_NOT_CALL_ALERT';

export interface Notification {
  id: string;
  userId: string;
  leadId?: string;
  type: NotificationType;
  title: string;
  message: string;
  read: boolean;
  readAt?: string;
  metadata?: Record<string, any>;
  createdAt: string;
}

export interface AuditLog {
  id: string;
  userId: string;
  userName: string;
  action: string;
  entityType: string;
  entityId: string;
  oldValue?: Record<string, any>;
  newValue?: Record<string, any>;
  reason?: string;
  metadata?: Record<string, any>;
  createdAt: string;
}

export interface SystemSettings {
  id: string;
  followUpThresholdHours: number;
  aiConfidenceThreshold: number;
  autoClassifyEnabled: boolean;
  humanReviewRequiredDNC: boolean;
  escalationHours: number;
  supportedLanguages: string[];
  duplicateCheckFields: string[];
  updatedAt: string;
}

export interface DashboardMetrics {
  totalLeads: number;
  totalCalls: number;
  callsToday: number;
  averageCallsPerLead: number;
  oneCallLeads: number;
  followUpRequired: number;
  followUpCompleted: number;
  overdueFollowUps: number;
  notInterested: number;
  doNotContact: number;
  interested: number;
  converted: number;
  lost: number;
  oneCallFollowUpRate: number;
  followUpCompletionRate: number;
  callsNeededToday?: number;
  unattended1stCount?: number;
  unattended2ndCount?: number;
  callbacksDueCount?: number;
  freshLeadsCount?: number;
  statusCounts: Record<string, number>;
  outcomeCounts: Record<string, number>;
  AICounts?: Record<string, number>;
  aiCounts: Record<string, number>;
  recentCalls: CallAttempt[];
}

export type BdaQueueCategory =
  | 'all'
  | 'unattended_1'
  | 'unattended_2'
  | 'callback_later'
  | 'fresh'
  | 'interested';

export interface BdaQueueResponse {
  items: Lead[];
  totalCount: number;
  page: number;
  pageSize: number;
  totalPages: number;
  categoryCounts: {
    all: number;
    unattended_1: number;
    unattended_2: number;
    callback_later: number;
    fresh: number;
    interested: number;
  };
}

export interface PaginatedResponse<T> {
  items: T[];
  totalCount: number;
  page: number;
  pageSize: number;
  totalPages: number;
}


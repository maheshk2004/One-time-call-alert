import axios from 'axios';
import {
  User,
  Lead,
  CallAttempt,
  Transcript,
  AIAnalysis,
  FollowUp,
  Notification,
  AuditLog,
  SystemSettings,
  DashboardMetrics,
  PaginatedResponse,
  BdaQueueResponse,
} from '../types';


const API_BASE_URL = 'http://localhost:8085/api';

export const api = axios.create({
  baseURL: API_BASE_URL,
  headers: {
    'Content-Type': 'application/json',
  },
});

// Attach Authorization token
api.interceptors.request.use((config) => {
  const token = localStorage.getItem('token');
  if (token && config.headers) {
    config.headers.Authorization = `Bearer ${token}`;
  }
  return config;
});

// Response interceptor for unwrapping { success: true, data: ... }
api.interceptors.response.use(
  (response) => {
    if (response.data && response.data.success) {
      return response.data.data;
    }
    return response.data;
  },
  (error) => {
    if (error.response?.status === 401 && window.location.pathname !== '/login') {
      localStorage.removeItem('token');
      localStorage.removeItem('user');
      window.location.href = '/login';
    }
    const message = error.response?.data?.message || error.message || 'Something went wrong';
    return Promise.reject(new Error(message));
  }
);

export const authApi = {
  login: (credentials: { email: string; password: string }): Promise<{ token: string; user: User }> =>
    api.post('/auth/login', credentials),
  getProfile: (): Promise<User> => api.get('/auth/me'),
  getUsers: (role?: string): Promise<User[]> => api.get('/users', { params: { role } }),
};

export const leadsApi = {
  getLeads: (params?: any): Promise<PaginatedResponse<Lead>> => api.get('/leads', { params }),
  getBdaQueue: (params?: any): Promise<BdaQueueResponse> => api.get('/bda/queue', { params }),
  getLead: (id: string): Promise<Lead> => api.get(`/leads/${id}`),
  createLead: (data: Partial<Lead>): Promise<Lead> => api.post('/leads', data),
  assignLead: (id: string, salespersonId: string): Promise<any> =>
    api.put(`/leads/${id}/assign`, { salespersonId }),
  updateStatus: (id: string, status: string, reason?: string): Promise<any> =>
    api.put(`/leads/${id}/status`, { status, reason }),
  setDoNotCall: (id: string, doNotCall: boolean, reason: string): Promise<any> =>
    api.put(`/leads/${id}/do-not-call`, { doNotCall, reason }),
  getOneCallLeads: (params?: any): Promise<PaginatedResponse<Lead>> =>
    api.get('/leads/one-call', { params }),
  getDoNotContact: (params?: any): Promise<PaginatedResponse<Lead>> =>
    api.get('/leads/do-not-contact', { params }),
};


export const callsApi = {
  recordCall: (leadId: string, data: any): Promise<CallAttempt> =>
    api.post(`/leads/${leadId}/calls`, data),
  getCallsForLead: (leadId: string): Promise<CallAttempt[]> =>
    api.get(`/leads/${leadId}/calls`),
  getCall: (id: string): Promise<CallAttempt> => api.get(`/calls/${id}`),
};

export const transcriptsApi = {
  saveTranscript: (callId: string, data: any): Promise<Transcript> =>
    api.post(`/calls/${callId}/transcript`, data),
  getTranscript: (callId: string): Promise<Transcript> =>
    api.get(`/calls/${callId}/transcript`),
};

export const aiApi = {
  analyzeCall: (callId: string): Promise<AIAnalysis> =>
    api.post(`/calls/${callId}/analyze`),
  getAnalysis: (callId: string): Promise<AIAnalysis> =>
    api.get(`/calls/${callId}/analysis`),
  getPendingReviews: (params?: any): Promise<PaginatedResponse<AIAnalysis>> =>
    api.get('/ai/reviews', { params }),
  confirmReview: (id: string): Promise<AIAnalysis> =>
    api.post(`/ai/reviews/${id}/confirm`),
  correctReview: (id: string, intent: string, reason: string): Promise<AIAnalysis> =>
    api.post(`/ai/reviews/${id}/correct`, { intent, reason }),
};

export const followUpsApi = {
  getFollowUps: (params?: any): Promise<PaginatedResponse<FollowUp>> =>
    api.get('/follow-ups', { params }),
  scheduleFollowUp: (leadId: string, data: { dueAt: string; notes?: string }): Promise<FollowUp> =>
    api.post(`/leads/${leadId}/follow-ups`, data),
  completeFollowUp: (id: string, callAttemptId?: string): Promise<any> =>
    api.post(`/follow-ups/${id}/complete`, { callAttemptId }),
  triggerDetection: (): Promise<{ message: string; leadsFlaggedForAlert: number }> =>
    api.post('/follow-ups/detect'),
};

export const notificationsApi = {
  getNotifications: (params?: any): Promise<{ notifications: Notification[]; unreadCount: number }> =>
    api.get('/notifications', { params }),
  markRead: (id: string): Promise<any> => api.put(`/notifications/${id}/read`),
  markAllRead: (): Promise<any> => api.put('/notifications/read-all'),
};

export const dashboardApi = {
  getSummary: (): Promise<DashboardMetrics> => api.get('/dashboard/summary'),
};

export const adminApi = {
  getSettings: (): Promise<SystemSettings> => api.get('/admin/settings'),
  updateSettings: (data: Partial<SystemSettings>): Promise<SystemSettings> =>
    api.put('/admin/settings', data),
  getAuditLogs: (params?: any): Promise<AuditLog[]> => api.get('/admin/audit-logs', { params }),
};

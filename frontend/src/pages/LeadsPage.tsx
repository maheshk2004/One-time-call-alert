import React, { useState, useEffect } from 'react';
import { leadsApi } from '../services/api';
import { Lead, LeadStatus, Priority, PaginatedResponse } from '../types';
import { Badge } from '../components/common/Badge';
import { Link } from 'react-router-dom';
import {
  Search,
  Plus,
  Filter,
  Phone,
  Mail,
  ChevronLeft,
  ChevronRight,
  ShieldBan,
  UserCheck,
  AlertCircle,
  X,
} from 'lucide-react';
import { useAuth } from '../context/AuthContext';
import { LiveCallModal } from '../components/calling/LiveCallModal';

export const LeadsPage: React.FC = () => {
  const [data, setData] = useState<PaginatedResponse<Lead> | null>(null);
  const [loading, setLoading] = useState(true);
  const [page, setPage] = useState(1);
  const [search, setSearch] = useState('');
  const [status, setStatus] = useState('');
  const [priority, setPriority] = useState('');
  const [source, setSource] = useState('');
  const [showCreateModal, setShowCreateModal] = useState(false);
  const [activeCallingLead, setActiveCallingLead] = useState<Lead | null>(null);

  // New lead form state
  const [newName, setNewName] = useState('');
  const [newPhone, setNewPhone] = useState('');
  const [newEmail, setNewEmail] = useState('');
  const [newCourse, setNewCourse] = useState('');
  const [newSource, setNewSource] = useState('Website Form');
  const [newPriority, setNewPriority] = useState<Priority>('MEDIUM');
  const [createError, setCreateError] = useState<string | null>(null);
  const [creating, setCreating] = useState(false);

  const fetchLeads = async () => {
    setLoading(true);
    try {
      const res = await leadsApi.getLeads({
        page,
        pageSize: 15,
        search,
        status,
        priority,
        source,
        doNotCall: false,
      });
      setData(res);
    } catch (err) {
      console.error(err);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchLeads();
  }, [page, status, priority, source]);

  const handleSearchSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    setPage(1);
    fetchLeads();
  };

  const handleCreateLead = async (e: React.FormEvent) => {
    e.preventDefault();
    setCreating(true);
    setCreateError(null);
    try {
      await leadsApi.createLead({
        name: newName,
        phone: newPhone,
        email: newEmail,
        course: newCourse,
        source: newSource,
        priority: newPriority,
      });
      setShowCreateModal(false);
      // Reset form
      setNewName('');
      setNewPhone('');
      setNewEmail('');
      setNewCourse('');
      fetchLeads();
    } catch (err: any) {
      setCreateError(err.message || 'Failed to create student lead');
    } finally {
      setCreating(false);
    }
  };

  return (
    <div className="space-y-6 bg-white min-h-screen">
      {/* Header */}
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
        <div>
          <h2 className="text-xl font-bold text-slate-900 tracking-tight">My Student Leads</h2>
          <p className="text-xs text-slate-500 mt-0.5">
            Active prospective students assigned for admissions outreach (DNC/uninterested leads filtered out)
          </p>
        </div>
        <div className="flex items-center gap-2.5">
          <Link
            to="/call-queue"
            className="flex items-center gap-2 px-4 py-2 rounded-xl bg-emerald-600 hover:bg-emerald-700 text-white text-xs font-bold shadow-xs transition-all"
          >
            <Phone className="w-3.5 h-3.5" />
            <span>Open Call Station</span>
          </Link>
          <button
            onClick={() => setShowCreateModal(true)}
            className="flex items-center gap-2 px-4 py-2 rounded-xl bg-indigo-600 hover:bg-indigo-700 text-white text-xs font-semibold shadow-xs transition-all"
          >
            <Plus className="w-4 h-4" />
            <span>Add Student</span>
          </button>
        </div>
      </div>

      {/* Filters Bar */}
      <div className="bg-white p-4 rounded-xl border border-slate-200 shadow-xs flex flex-wrap gap-3 items-center justify-between">
        <form onSubmit={handleSearchSubmit} className="flex-1 min-w-[240px] max-w-md relative">
          <Search className="w-4 h-4 text-slate-400 absolute left-3 top-2.5" />
          <input
            type="text"
            value={search}
            onChange={(e) => setSearch(e.target.value)}
            placeholder="Search by name, phone, email, course..."
            className="w-full pl-9 pr-4 py-2 bg-slate-50 border border-slate-200 rounded-lg text-xs text-slate-900 placeholder-slate-400 focus:outline-none focus:border-indigo-600 focus:bg-white transition"
          />
        </form>

        <div className="flex flex-wrap items-center gap-2">
          {/* Status filter */}
          <select
            value={status}
            onChange={(e) => {
              setStatus(e.target.value);
              setPage(1);
            }}
            className="bg-white border border-slate-200 rounded-lg text-xs text-slate-700 py-2 px-3 focus:outline-none focus:border-indigo-600 cursor-pointer shadow-xs"
          >
            <option value="">All Statuses</option>
            <option value="FOLLOW_UP_REQUIRED">Follow-Up Required (Urgent)</option>
            <option value="CALLED_ONCE">Called Once</option>
            <option value="FOLLOW_UP_SCHEDULED">Follow-Up Scheduled</option>
            <option value="INTERESTED">Interested</option>
            <option value="UNDECIDED">Undecided</option>
            <option value="NOT_INTERESTED">Not Interested</option>
            <option value="DO_NOT_CALL">Do Not Call</option>
            <option value="CONVERTED">Converted</option>
            <option value="NEW">New</option>
          </select>

          {/* Priority filter */}
          <select
            value={priority}
            onChange={(e) => {
              setPriority(e.target.value);
              setPage(1);
            }}
            className="bg-white border border-slate-200 rounded-lg text-xs text-slate-700 py-2 px-3 focus:outline-none focus:border-indigo-600 cursor-pointer shadow-xs"
          >
            <option value="">All Priorities</option>
            <option value="URGENT">Urgent</option>
            <option value="HIGH">High</option>
            <option value="MEDIUM">Medium</option>
            <option value="LOW">Low</option>
          </select>

          {/* Source filter */}
          <select
            value={source}
            onChange={(e) => {
              setSource(e.target.value);
              setPage(1);
            }}
            className="bg-white border border-slate-200 rounded-lg text-xs text-slate-700 py-2 px-3 focus:outline-none focus:border-indigo-600 cursor-pointer shadow-xs"
          >
            <option value="">All Sources</option>
            <option value="Facebook Ads">Facebook Ads</option>
            <option value="Google Search Ads">Google Ads</option>
            <option value="LinkedIn Ads">LinkedIn Ads</option>
            <option value="Website Form">Website Form</option>
            <option value="WhatsApp Inbound">WhatsApp</option>
          </select>

          {(search || status || priority || source) && (
            <button
              onClick={() => {
                setSearch('');
                setStatus('');
                setPriority('');
                setSource('');
                setPage(1);
              }}
              className="text-xs text-slate-500 hover:text-slate-900 px-2 py-1 font-medium"
            >
              Reset
            </button>
          )}
        </div>
      </div>

      {/* Leads Table */}
      <div className="bg-white rounded-2xl border border-slate-200 overflow-hidden shadow-xs">
        {loading ? (
          <div className="py-20 flex flex-col items-center justify-center gap-2">
            <div className="w-6 h-6 border-2 border-indigo-600 border-t-transparent rounded-full animate-spin" />
            <span className="text-xs text-slate-500">Loading leads...</span>
          </div>
        ) : !data || data.items.length === 0 ? (
          <div className="py-16 text-center text-slate-500">
            <AlertCircle className="w-8 h-8 text-slate-400 mx-auto mb-2" />
            <p className="text-sm font-medium">No leads found matching your criteria.</p>
          </div>
        ) : (
          <div className="overflow-x-auto">
            <table className="w-full text-left text-xs">
              <thead className="bg-slate-50 text-slate-500 uppercase text-[10px] tracking-wider border-b border-slate-200 font-semibold">
                <tr>
                  <th className="py-3 px-4">Lead</th>
                  <th className="py-3 px-4">Contact</th>
                  <th className="py-3 px-4">Source / Course</th>
                  <th className="py-3 px-4">Assigned Rep</th>
                  <th className="py-3 px-4 text-center">Attempts</th>
                  <th className="py-3 px-4">Status</th>
                  <th className="py-3 px-4 text-right">Actions</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-slate-100">
                {data.items.map((lead) => (
                  <tr key={lead.id} className="hover:bg-slate-50/80 transition-colors">
                    <td className="py-3 px-4">
                      <div className="font-semibold text-slate-900 flex items-center gap-1.5">
                        <Link to={`/leads/${lead.id}`} className="hover:text-indigo-600">
                          {lead.name}
                        </Link>
                        {lead.doNotCall && (
                          <span title="Do Not Call Lead">
                            <ShieldBan className="w-3.5 h-3.5 text-purple-600 inline" />
                          </span>
                        )}
                      </div>
                      <div className="text-[10px] text-slate-400">{lead.location || 'India'}</div>
                    </td>

                    <td className="py-3 px-4">
                      <div className="text-slate-800 font-mono text-[11px] flex items-center gap-1 font-semibold">
                        <Phone className="w-3 h-3 text-slate-400" />
                        {lead.phone}
                      </div>
                      {lead.email && (
                        <div className="text-slate-500 text-[10px] truncate max-w-[150px] flex items-center gap-1">
                          <Mail className="w-3 h-3 text-slate-400" />
                          {lead.email}
                        </div>
                      )}
                    </td>

                    <td className="py-3 px-4">
                      <span className="text-[10px] px-2 py-0.5 rounded-full bg-slate-100 text-slate-700 font-medium">
                        {lead.source}
                      </span>
                      <div className="text-slate-600 text-[11px] mt-0.5 truncate max-w-[180px]">
                        {lead.course || '—'}
                      </div>
                    </td>

                    <td className="py-3 px-4">
                      {lead.assignedUser ? (
                        <div className="flex items-center gap-1.5">
                          <UserCheck className="w-3.5 h-3.5 text-emerald-600" />
                          <span className="text-slate-800 font-medium">{lead.assignedUser.name}</span>
                        </div>
                      ) : (
                        <span className="text-slate-400 italic">Unassigned</span>
                      )}
                    </td>

                    <td className="py-3 px-4 text-center">
                      <span
                        className={`inline-block px-2 py-0.5 rounded-md font-mono text-[11px] font-bold ${
                          lead.callAttemptCount === 1
                            ? 'bg-amber-50 text-amber-700 border border-amber-200'
                            : lead.callAttemptCount > 1
                            ? 'bg-rose-50 text-rose-700 border border-rose-200'
                            : 'bg-slate-100 text-slate-500'
                        }`}
                      >
                        {lead.callAttemptCount}
                      </span>
                    </td>

                    <td className="py-3 px-4">
                      <Badge status={lead.status} size="sm" />
                    </td>

                    <td className="py-3 px-4 text-right whitespace-nowrap">
                      <div className="flex items-center justify-end gap-1.5">
                        <button
                          type="button"
                          onClick={() => setActiveCallingLead(lead)}
                          className="inline-flex items-center gap-1 px-2.5 py-1 rounded-lg bg-emerald-50 border border-emerald-200 hover:bg-emerald-100 text-emerald-700 text-xs font-bold transition-colors cursor-pointer"
                        >
                          <Phone className="w-3 h-3" />
                          Call
                        </button>
                        <Link
                          to={`/leads/${lead.id}`}
                          className="inline-flex items-center gap-1 px-2.5 py-1 rounded-lg bg-slate-100 hover:bg-slate-200 text-slate-700 text-xs font-semibold transition-colors"
                        >
                          Details →
                        </Link>
                      </div>
                    </td>

                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}

        {/* Pagination Bar */}
        {data && data.totalPages > 1 && (
          <div className="p-4 border-t border-slate-200 flex items-center justify-between text-xs text-slate-500">
            <span>
              Showing {(data.page - 1) * data.pageSize + 1} to{' '}
              {Math.min(data.page * data.pageSize, data.totalCount)} of {data.totalCount} leads
            </span>
            <div className="flex items-center gap-2">
              <button
                disabled={page <= 1}
                onClick={() => setPage(page - 1)}
                className="p-1.5 rounded-lg bg-slate-100 hover:bg-slate-200 text-slate-700 disabled:opacity-40 transition-colors"
              >
                <ChevronLeft className="w-4 h-4" />
              </button>
              <span className="font-semibold text-slate-800">
                Page {page} of {data.totalPages}
              </span>
              <button
                disabled={page >= data.totalPages}
                onClick={() => setPage(page + 1)}
                className="p-1.5 rounded-lg bg-slate-100 hover:bg-slate-200 text-slate-700 disabled:opacity-40 transition-colors"
              >
                <ChevronRight className="w-4 h-4" />
              </button>
            </div>
          </div>
        )}
      </div>

      {/* Quick Create Lead Modal */}
      {showCreateModal && (
        <div className="fixed inset-0 z-50 bg-slate-900/60 backdrop-blur-sm flex items-center justify-center p-4">
          <div className="bg-white w-full max-w-lg rounded-2xl p-6 shadow-2xl border border-slate-200 animate-in fade-in zoom-in-95">
            <div className="flex items-center justify-between pb-4 border-b border-slate-200 mb-4">
              <h3 className="text-base font-bold text-slate-900">Add New Student Lead</h3>
              <button
                onClick={() => setShowCreateModal(false)}
                className="p-1 rounded-lg hover:bg-slate-100 text-slate-400 hover:text-slate-700"
              >
                <X className="w-4 h-4" />
              </button>
            </div>

            {createError && (
              <div className="mb-4 p-3 rounded-lg bg-rose-50 border border-rose-200 text-xs text-rose-700 font-medium">
                {createError}
              </div>
            )}

            <form onSubmit={handleCreateLead} className="space-y-4 text-xs">
              <div>
                <label className="block text-slate-700 font-medium mb-1">Full Name *</label>
                <input
                  type="text"
                  required
                  value={newName}
                  onChange={(e) => setNewName(e.target.value)}
                  placeholder="e.g. Ramesh Kumar"
                  className="w-full p-2.5 bg-slate-50 border border-slate-200 rounded-lg text-slate-900 focus:outline-none focus:border-indigo-600 focus:bg-white"
                />
              </div>

              <div className="grid grid-cols-2 gap-3">
                <div>
                  <label className="block text-slate-700 font-medium mb-1">Phone Number *</label>
                  <input
                    type="tel"
                    required
                    value={newPhone}
                    onChange={(e) => setNewPhone(e.target.value)}
                    placeholder="+91 98765 43210"
                    className="w-full p-2.5 bg-slate-50 border border-slate-200 rounded-lg text-slate-900 focus:outline-none focus:border-indigo-600 focus:bg-white font-mono"
                  />
                  <span className="text-[10px] text-slate-400 mt-0.5 block">Duplicate check enforced</span>
                </div>
                <div>
                  <label className="block text-slate-700 font-medium mb-1">Email Address</label>
                  <input
                    type="email"
                    value={newEmail}
                    onChange={(e) => setNewEmail(e.target.value)}
                    placeholder="ramesh@example.com"
                    className="w-full p-2.5 bg-slate-50 border border-slate-200 rounded-lg text-slate-900 focus:outline-none focus:border-indigo-600 focus:bg-white"
                  />
                </div>
              </div>

              <div className="grid grid-cols-2 gap-3">
                <div>
                  <label className="block text-slate-700 font-medium mb-1">Course / Offering</label>
                  <input
                    type="text"
                    value={newCourse}
                    onChange={(e) => setNewCourse(e.target.value)}
                    placeholder="e.g. Full-Stack Web Dev"
                    className="w-full p-2.5 bg-slate-50 border border-slate-200 rounded-lg text-slate-900 focus:outline-none focus:border-indigo-600 focus:bg-white"
                  />
                </div>
                <div>
                  <label className="block text-slate-700 font-medium mb-1">Lead Source</label>
                  <select
                    value={newSource}
                    onChange={(e) => setNewSource(e.target.value)}
                    className="w-full p-2.5 bg-slate-50 border border-slate-200 rounded-lg text-slate-900 focus:outline-none focus:border-indigo-600 focus:bg-white"
                  >
                    <option value="Website Form">Website Form</option>
                    <option value="Facebook Ads">Facebook Ads</option>
                    <option value="Google Search Ads">Google Ads</option>
                    <option value="LinkedIn Ads">LinkedIn Ads</option>
                    <option value="WhatsApp Inbound">WhatsApp Inbound</option>
                    <option value="Manual Entry">Manual Entry</option>
                  </select>
                </div>
              </div>

              <div>
                <label className="block text-slate-700 font-medium mb-1">Priority</label>
                <select
                  value={newPriority}
                  onChange={(e) => setNewPriority(e.target.value as Priority)}
                  className="w-full p-2.5 bg-slate-50 border border-slate-200 rounded-lg text-slate-900 focus:outline-none focus:border-indigo-600 focus:bg-white"
                >
                  <option value="URGENT">Urgent</option>
                  <option value="HIGH">High</option>
                  <option value="MEDIUM">Medium</option>
                  <option value="LOW">Low</option>
                </select>
              </div>

              <div className="flex items-center justify-end gap-3 pt-3 border-t border-slate-200">
                <button
                  type="button"
                  onClick={() => setShowCreateModal(false)}
                  className="px-4 py-2 rounded-lg bg-slate-100 hover:bg-slate-200 text-slate-700 font-medium transition-colors"
                >
                  Cancel
                </button>
                <button
                  type="submit"
                  disabled={creating}
                  className="px-4 py-2 rounded-lg bg-indigo-600 hover:bg-indigo-700 text-white font-semibold transition-all shadow-xs disabled:opacity-50"
                >
                  {creating ? 'Registering...' : 'Create Lead'}
                </button>
              </div>
            </form>
          </div>
        </div>
      )}

      {/* Live Calling & Real-Time STT Modal */}
      {activeCallingLead && (
        <LiveCallModal
          lead={activeCallingLead}
          onClose={() => setActiveCallingLead(null)}
          onSuccess={() => fetchLeads()}
        />
      )}
    </div>
  );
};

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

export const LeadsPage: React.FC = () => {
  const [data, setData] = useState<PaginatedResponse<Lead> | null>(null);
  const [loading, setLoading] = useState(true);
  const [page, setPage] = useState(1);
  const [search, setSearch] = useState('');
  const [status, setStatus] = useState('');
  const [priority, setPriority] = useState('');
  const [source, setSource] = useState('');
  const [showCreateModal, setShowCreateModal] = useState(false);

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
    <div className="space-y-6">
      {/* Header */}
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
        <div>
          <h2 className="text-xl font-bold text-white tracking-tight">My Student Leads</h2>
          <p className="text-xs text-slate-400 mt-0.5">
            Active prospective students assigned for admissions outreach (DNC/uninterested leads filtered out)
          </p>
        </div>
        <div className="flex items-center gap-2.5">
          <Link
            to="/call-queue"
            className="flex items-center gap-2 px-4 py-2 rounded-xl bg-emerald-600 hover:bg-emerald-500 text-white text-xs font-bold shadow-lg shadow-emerald-600/20 transition-all"
          >
            <Phone className="w-3.5 h-3.5" />
            <span>Open Call Station</span>
          </Link>
          <button
            onClick={() => setShowCreateModal(true)}
            className="flex items-center gap-2 px-4 py-2 rounded-xl bg-indigo-600 hover:bg-indigo-500 text-white text-xs font-semibold shadow-lg shadow-indigo-600/20 transition-all"
          >
            <Plus className="w-4 h-4" />
            <span>Add Student</span>
          </button>
        </div>
      </div>

      {/* Filters Bar */}
      <div className="glass-panel p-4 rounded-xl border border-gray-800 flex flex-wrap gap-3 items-center justify-between">
        <form onSubmit={handleSearchSubmit} className="flex-1 min-w-[240px] max-w-md relative">
          <Search className="w-4 h-4 text-gray-400 absolute left-3 top-2.5" />
          <input
            type="text"
            value={search}
            onChange={(e) => setSearch(e.target.value)}
            placeholder="Search by name, phone, email, course..."
            className="w-full pl-9 pr-4 py-2 bg-gray-900/80 border border-gray-700/70 rounded-lg text-xs text-white placeholder-gray-500 focus:outline-none focus:border-indigo-500"
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
            className="bg-gray-900 border border-gray-700 rounded-lg text-xs text-gray-300 py-2 px-3 focus:outline-none focus:border-indigo-500"
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
            className="bg-gray-900 border border-gray-700 rounded-lg text-xs text-gray-300 py-2 px-3 focus:outline-none focus:border-indigo-500"
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
            className="bg-gray-900 border border-gray-700 rounded-lg text-xs text-gray-300 py-2 px-3 focus:outline-none focus:border-indigo-500"
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
              className="text-xs text-gray-400 hover:text-white px-2 py-1"
            >
              Reset
            </button>
          )}
        </div>
      </div>

      {/* Leads Table */}
      <div className="glass-panel rounded-2xl border border-gray-800 overflow-hidden">
        {loading ? (
          <div className="py-20 flex flex-col items-center justify-center gap-2">
            <div className="w-6 h-6 border-2 border-indigo-500 border-t-transparent rounded-full animate-spin" />
            <span className="text-xs text-gray-400">Loading leads...</span>
          </div>
        ) : !data || data.items.length === 0 ? (
          <div className="py-16 text-center text-gray-400">
            <AlertCircle className="w-8 h-8 text-gray-600 mx-auto mb-2" />
            <p className="text-sm font-medium">No leads found matching your criteria.</p>
          </div>
        ) : (
          <div className="overflow-x-auto">
            <table className="w-full text-left text-xs">
              <thead className="bg-gray-900/80 text-gray-400 uppercase text-[10px] tracking-wider border-b border-gray-800">
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
              <tbody className="divide-y divide-gray-800/60">
                {data.items.map((lead) => (
                  <tr key={lead.id} className="hover:bg-gray-800/30 transition-colors">
                    <td className="py-3 px-4">
                      <div className="font-semibold text-white flex items-center gap-1.5">
                        <Link to={`/leads/${lead.id}`} className="hover:text-indigo-300">
                          {lead.name}
                        </Link>
                        {lead.doNotCall && (
                          <span title="Do Not Call Lead">
                            <ShieldBan className="w-3.5 h-3.5 text-purple-400 inline" />
                          </span>
                        )}
                      </div>
                      <div className="text-[10px] text-gray-500">{lead.location || 'India'}</div>
                    </td>

                    <td className="py-3 px-4">
                      <div className="text-gray-300 font-mono text-[11px] flex items-center gap-1">
                        <Phone className="w-3 h-3 text-gray-500" />
                        {lead.phone}
                      </div>
                      {lead.email && (
                        <div className="text-gray-500 text-[10px] truncate max-w-[150px] flex items-center gap-1">
                          <Mail className="w-3 h-3 text-gray-600" />
                          {lead.email}
                        </div>
                      )}
                    </td>

                    <td className="py-3 px-4">
                      <span className="text-[10px] px-2 py-0.5 rounded-full bg-gray-800 text-gray-300 font-medium">
                        {lead.source}
                      </span>
                      <div className="text-gray-400 text-[11px] mt-0.5 truncate max-w-[180px]">
                        {lead.course || '—'}
                      </div>
                    </td>

                    <td className="py-3 px-4">
                      {lead.assignedUser ? (
                        <div className="flex items-center gap-1.5">
                          <UserCheck className="w-3.5 h-3.5 text-emerald-400" />
                          <span className="text-gray-200">{lead.assignedUser.name}</span>
                        </div>
                      ) : (
                        <span className="text-gray-500 italic">Unassigned</span>
                      )}
                    </td>

                    <td className="py-3 px-4 text-center">
                      <span
                        className={`inline-block px-2 py-0.5 rounded-md font-mono text-[11px] font-bold ${
                          lead.callAttemptCount === 1
                            ? 'bg-amber-500/15 text-amber-300 border border-amber-500/30'
                            : lead.callAttemptCount > 1
                            ? 'bg-indigo-500/15 text-indigo-300 border border-indigo-500/30'
                            : 'bg-gray-800 text-gray-500'
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
                        <Link
                          to="/call-queue"
                          className="inline-flex items-center gap-1 px-2.5 py-1 rounded-lg bg-emerald-600/20 border border-emerald-500/30 hover:bg-emerald-600/30 text-emerald-300 text-xs font-bold transition-colors"
                        >
                          <Phone className="w-3 h-3" />
                          Call
                        </Link>
                        <Link
                          to={`/leads/${lead.id}`}
                          className="inline-flex items-center gap-1 px-2.5 py-1 rounded-lg bg-gray-800 hover:bg-gray-700 text-gray-200 text-xs font-medium transition-colors"
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
          <div className="p-4 border-t border-gray-800 flex items-center justify-between text-xs text-gray-400">
            <span>
              Showing {(data.page - 1) * data.pageSize + 1} to{' '}
              {Math.min(data.page * data.pageSize, data.totalCount)} of {data.totalCount} leads
            </span>
            <div className="flex items-center gap-2">
              <button
                disabled={page <= 1}
                onClick={() => setPage(page - 1)}
                className="p-1.5 rounded-lg bg-gray-800 hover:bg-gray-700 disabled:opacity-40 transition-colors"
              >
                <ChevronLeft className="w-4 h-4" />
              </button>
              <span className="font-semibold text-white">
                Page {page} of {data.totalPages}
              </span>
              <button
                disabled={page >= data.totalPages}
                onClick={() => setPage(page + 1)}
                className="p-1.5 rounded-lg bg-gray-800 hover:bg-gray-700 disabled:opacity-40 transition-colors"
              >
                <ChevronRight className="w-4 h-4" />
              </button>
            </div>
          </div>
        )}
      </div>

      {/* Quick Create Lead Modal with Duplicate Detection Warning */}
      {showCreateModal && (
        <div className="fixed inset-0 z-50 bg-black/70 backdrop-blur-sm flex items-center justify-center p-4">
          <div className="glass-panel w-full max-w-lg rounded-2xl p-6 shadow-2xl border border-gray-700 animate-in fade-in zoom-in-95">
            <div className="flex items-center justify-between pb-4 border-b border-gray-800 mb-4">
              <h3 className="text-base font-bold text-white">Add New Sales Lead</h3>
              <button
                onClick={() => setShowCreateModal(false)}
                className="p-1 rounded-lg hover:bg-gray-800 text-gray-400 hover:text-white"
              >
                <X className="w-4 h-4" />
              </button>
            </div>

            {createError && (
              <div className="mb-4 p-3 rounded-lg bg-red-500/20 border border-red-500/40 text-xs text-red-300 font-medium">
                {createError}
              </div>
            )}

            <form onSubmit={handleCreateLead} className="space-y-4 text-xs">
              <div>
                <label className="block text-gray-300 font-medium mb-1">Full Name *</label>
                <input
                  type="text"
                  required
                  value={newName}
                  onChange={(e) => setNewName(e.target.value)}
                  placeholder="e.g. Ramesh Kumar"
                  className="w-full p-2.5 bg-gray-900 border border-gray-700 rounded-lg text-white focus:outline-none focus:border-indigo-500"
                />
              </div>

              <div className="grid grid-cols-2 gap-3">
                <div>
                  <label className="block text-gray-300 font-medium mb-1">Phone Number *</label>
                  <input
                    type="tel"
                    required
                    value={newPhone}
                    onChange={(e) => setNewPhone(e.target.value)}
                    placeholder="+91 98765 43210"
                    className="w-full p-2.5 bg-gray-900 border border-gray-700 rounded-lg text-white focus:outline-none focus:border-indigo-500"
                  />
                  <span className="text-[10px] text-gray-500 mt-0.5 block">Duplicate check enforced</span>
                </div>
                <div>
                  <label className="block text-gray-300 font-medium mb-1">Email Address</label>
                  <input
                    type="email"
                    value={newEmail}
                    onChange={(e) => setNewEmail(e.target.value)}
                    placeholder="ramesh@example.com"
                    className="w-full p-2.5 bg-gray-900 border border-gray-700 rounded-lg text-white focus:outline-none focus:border-indigo-500"
                  />
                </div>
              </div>

              <div className="grid grid-cols-2 gap-3">
                <div>
                  <label className="block text-gray-300 font-medium mb-1">Course / Offering</label>
                  <input
                    type="text"
                    value={newCourse}
                    onChange={(e) => setNewCourse(e.target.value)}
                    placeholder="e.g. Full-Stack Web Dev"
                    className="w-full p-2.5 bg-gray-900 border border-gray-700 rounded-lg text-white focus:outline-none focus:border-indigo-500"
                  />
                </div>
                <div>
                  <label className="block text-gray-300 font-medium mb-1">Lead Source</label>
                  <select
                    value={newSource}
                    onChange={(e) => setNewSource(e.target.value)}
                    className="w-full p-2.5 bg-gray-900 border border-gray-700 rounded-lg text-white focus:outline-none focus:border-indigo-500"
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
                <label className="block text-gray-300 font-medium mb-1">Priority</label>
                <select
                  value={newPriority}
                  onChange={(e) => setNewPriority(e.target.value as Priority)}
                  className="w-full p-2.5 bg-gray-900 border border-gray-700 rounded-lg text-white focus:outline-none focus:border-indigo-500"
                >
                  <option value="URGENT">Urgent</option>
                  <option value="HIGH">High</option>
                  <option value="MEDIUM">Medium</option>
                  <option value="LOW">Low</option>
                </select>
              </div>

              <div className="flex items-center justify-end gap-3 pt-3 border-t border-gray-800">
                <button
                  type="button"
                  onClick={() => setShowCreateModal(false)}
                  className="px-4 py-2 rounded-lg bg-gray-800 hover:bg-gray-700 text-gray-300 font-medium transition-colors"
                >
                  Cancel
                </button>
                <button
                  type="submit"
                  disabled={creating}
                  className="px-4 py-2 rounded-lg bg-indigo-600 hover:bg-indigo-500 text-white font-semibold transition-all shadow-md shadow-indigo-600/20 disabled:opacity-50"
                >
                  {creating ? 'Registering...' : 'Create Lead'}
                </button>
              </div>
            </form>
          </div>
        </div>
      )}
    </div>
  );
};

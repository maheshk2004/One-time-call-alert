import React, { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { useAuth } from '../context/AuthContext';
import { ShieldCheck, ArrowRight, Sparkles, User, Lock, PhoneCall } from 'lucide-react';

export const LoginPage: React.FC = () => {
  const [email, setEmail] = useState('vikram@leadfollowup.com');
  const [password, setPassword] = useState('password123');
  const [error, setError] = useState<string | null>(null);
  const [isLoading, setIsLoading] = useState(false);
  const { login } = useAuth();
  const navigate = useNavigate();

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError(null);
    setIsLoading(true);
    try {
      await login(email, password);
      navigate('/call-queue');
    } catch (err: any) {
      setError(err.message || 'Login failed');
    } finally {
      setIsLoading(false);
    }
  };

  const handleQuickLogin = (demoEmail: string) => {
    setEmail(demoEmail);
    setPassword('password123');
  };

  return (
    <div className="min-h-screen bg-[#0b0f19] flex items-center justify-center p-4 relative overflow-hidden">
      {/* Glow background effects */}
      <div className="absolute -top-40 -left-40 w-96 h-96 bg-indigo-600/15 rounded-full blur-3xl pointer-events-none" />
      <div className="absolute -bottom-40 -right-40 w-96 h-96 bg-emerald-600/15 rounded-full blur-3xl pointer-events-none" />

      <div className="w-full max-w-md relative z-10">
        <div className="bg-slate-900/90 rounded-2xl p-8 shadow-2xl border border-slate-800 backdrop-blur-xl">
          <div className="text-center mb-8">
            <div className="w-12 h-12 rounded-2xl bg-gradient-to-tr from-indigo-600 to-emerald-500 mx-auto flex items-center justify-center font-bold text-white text-xl shadow-lg shadow-indigo-500/25 mb-4">
              <PhoneCall className="w-6 h-6" />
            </div>
            <h2 className="text-2xl font-bold text-white tracking-tight">Sales BDA Calling Station</h2>
            <p className="text-xs text-slate-400 mt-1">1-Call & 2-Call Alert Monitoring & Admissions Outreach</p>
          </div>

          {error && (
            <div className="mb-6 p-3 rounded-lg bg-red-500/15 border border-red-500/30 text-xs text-red-400 text-center font-medium">
              {error}
            </div>
          )}

          <form onSubmit={handleSubmit} className="space-y-4">
            <div>
              <label className="block text-xs font-semibold text-slate-300 uppercase tracking-wider mb-1.5">
                BDA Work Email
              </label>
              <div className="relative">
                <User className="w-4 h-4 text-slate-500 absolute left-3.5 top-3" />
                <input
                  type="email"
                  value={email}
                  onChange={(e) => setEmail(e.target.value)}
                  required
                  placeholder="vikram@leadfollowup.com"
                  className="w-full pl-10 pr-4 py-2.5 bg-slate-800/80 border border-slate-700/70 rounded-xl text-sm text-white focus:outline-none focus:border-indigo-500 transition-colors"
                />
              </div>
            </div>

            <div>
              <label className="block text-xs font-semibold text-slate-300 uppercase tracking-wider mb-1.5">
                Password
              </label>
              <div className="relative">
                <Lock className="w-4 h-4 text-slate-500 absolute left-3.5 top-3" />
                <input
                  type="password"
                  value={password}
                  onChange={(e) => setPassword(e.target.value)}
                  required
                  placeholder="••••••••••••"
                  className="w-full pl-10 pr-4 py-2.5 bg-slate-800/80 border border-slate-700/70 rounded-xl text-sm text-white focus:outline-none focus:border-indigo-500 transition-colors"
                />
              </div>
            </div>

            <button
              type="submit"
              disabled={isLoading}
              className="w-full py-3 px-4 rounded-xl bg-gradient-to-r from-indigo-600 to-emerald-600 hover:from-indigo-500 hover:to-emerald-500 text-white text-sm font-bold shadow-lg shadow-indigo-500/25 flex items-center justify-center gap-2 transition-all disabled:opacity-50"
            >
              <span>{isLoading ? 'Authenticating...' : 'Enter Calling Station'}</span>
              <ArrowRight className="w-4 h-4" />
            </button>
          </form>

          {/* Quick Demo BDA Personas */}
          <div className="mt-8 pt-6 border-t border-slate-800">
            <div className="flex items-center gap-2 mb-3 text-xs text-slate-400 font-medium">
              <Sparkles className="w-3.5 h-3.5 text-amber-400" />
              <span>Select Sales BDA Persona (password: password123):</span>
            </div>
            <div className="grid grid-cols-3 gap-2">
              <button
                type="button"
                onClick={() => handleQuickLogin('vikram@leadfollowup.com')}
                className={`p-2.5 text-left rounded-xl border transition ${
                  email === 'vikram@leadfollowup.com'
                    ? 'bg-indigo-600/30 border-indigo-500 text-white'
                    : 'bg-slate-800/60 border-slate-700/60 text-slate-300 hover:bg-slate-800'
                }`}
              >
                <div className="text-xs font-bold">Vikram</div>
                <div className="text-[10px] text-slate-400">Senior BDA</div>
              </button>
              <button
                type="button"
                onClick={() => handleQuickLogin('priya@leadfollowup.com')}
                className={`p-2.5 text-left rounded-xl border transition ${
                  email === 'priya@leadfollowup.com'
                    ? 'bg-indigo-600/30 border-indigo-500 text-white'
                    : 'bg-slate-800/60 border-slate-700/60 text-slate-300 hover:bg-slate-800'
                }`}
              >
                <div className="text-xs font-bold">Priya</div>
                <div className="text-[10px] text-slate-400">Sales BDA</div>
              </button>
              <button
                type="button"
                onClick={() => handleQuickLogin('rahul@leadfollowup.com')}
                className={`p-2.5 text-left rounded-xl border transition ${
                  email === 'rahul@leadfollowup.com'
                    ? 'bg-indigo-600/30 border-indigo-500 text-white'
                    : 'bg-slate-800/60 border-slate-700/60 text-slate-300 hover:bg-slate-800'
                }`}
              >
                <div className="text-xs font-bold">Rahul</div>
                <div className="text-[10px] text-slate-400">Sales BDA</div>
              </button>
            </div>
          </div>
        </div>

        <div className="text-center mt-6 text-xs text-slate-500 flex items-center justify-center gap-1.5">
          <ShieldCheck className="w-3.5 h-3.5 text-emerald-400" />
          <span>Automatic DNC & Not-Interested Suppression Active</span>
        </div>
      </div>
    </div>
  );
};

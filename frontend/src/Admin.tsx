import { useState } from 'react';
import { Navigate } from 'react-router-dom';
import type { User } from 'firebase/auth';
import { Settings, Plus, Power, RefreshCcw, DollarSign } from 'lucide-react';

interface AdminProps {
    user: User | null;
}

function Admin({ user }: AdminProps) {
    const [activeTab, setActiveTab] = useState<'create' | 'manage'>('manage');
    const [compName, setCompName] = useState('');
    const [file, setFile] = useState<File | null>(null);
    const [creating, setCreating] = useState(false);
    const [createMsg, setCreateMsg] = useState<{ text: string, type: 'success' | 'error' } | null>(null);

    // Strict check - only this email is allowed
    const isAdmin = user?.email === 'drinkfood.exe@gmail.com';

    if (!isAdmin) {
        return <Navigate to="/" replace />;
    }

    return (
        <div className="min-h-screen bg-[#0a0a0a] text-gray-100 flex flex-col pt-4">
            <div className="max-w-7xl mx-auto w-full px-4 sm:px-6 lg:px-8 space-y-8 pb-12">
                {/* Header */}
                <div className="flex flex-col md:flex-row md:items-end justify-between gap-4">
                    <div>
                        <div className="flex items-center gap-3 mb-2">
                            <div className="p-2.5 bg-indigo-500/10 rounded-xl border border-indigo-500/20">
                                <Settings className="w-6 h-6 text-indigo-400" />
                            </div>
                            <h1 className="text-3xl font-bold">Admin Dashboard</h1>
                        </div>
                        <p className="text-gray-400">Configure events, resolve markets, and manage the trading engine.</p>
                    </div>

                    <div className="flex bg-gray-900 border border-gray-800 rounded-xl p-1">
                        <button
                            onClick={() => setActiveTab('manage')}
                            className={`px-6 py-2 rounded-lg text-sm font-medium transition-all ${activeTab === 'manage'
                                ? 'bg-gray-800 shadow-sm text-white'
                                : 'text-gray-400 hover:text-gray-200'
                                }`}
                        >
                            Manage Engine
                        </button>
                        <button
                            onClick={() => setActiveTab('create')}
                            className={`px-6 py-2 rounded-lg text-sm font-medium transition-all ${activeTab === 'create'
                                ? 'bg-gray-800 shadow-sm text-white'
                                : 'text-gray-400 hover:text-gray-200'
                                }`}
                        >
                            Create Data
                        </button>
                    </div>
                </div>

                {activeTab === 'manage' && (
                    <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
                        {/* Kill Switch Card */}
                        <div className="p-6 rounded-2xl bg-gray-900 border border-gray-800 flex flex-col justify-between">
                            <div className="mb-6">
                                <div className="flex items-center gap-3 mb-2">
                                    <div className="p-2 bg-red-500/10 rounded-lg">
                                        <Power className="w-5 h-5 text-red-500" />
                                    </div>
                                    <h3 className="font-bold text-lg">Kill Switch</h3>
                                </div>
                                <p className="text-sm text-gray-400">Instantly halt all trading for an active competition.</p>
                            </div>
                            <button className="w-full py-3 px-4 rounded-xl bg-red-950/50 hover:bg-red-900/50 border border-red-900/50 text-red-400 font-semibold transition-colors flex items-center justify-center gap-2">
                                <Power className="w-4 h-4" />
                                Pause Trading
                            </button>
                        </div>

                        {/* Resolve Markets Card */}
                        <div className="p-6 rounded-2xl bg-gray-900 border border-gray-800 flex flex-col justify-between">
                            <div className="mb-6">
                                <div className="flex items-center gap-3 mb-2">
                                    <div className="p-2 bg-emerald-500/10 rounded-lg">
                                        <DollarSign className="w-5 h-5 text-emerald-500" />
                                    </div>
                                    <h3 className="font-bold text-lg">Resolve & Payout</h3>
                                </div>
                                <p className="text-sm text-gray-400">Trigger the payout engine to distribute balances for a completed event.</p>
                            </div>
                            <button className="w-full py-3 px-4 rounded-xl bg-emerald-950/50 hover:bg-emerald-900/50 border border-emerald-900/50 text-emerald-400 font-semibold transition-colors flex items-center justify-center gap-2">
                                <DollarSign className="w-4 h-4" />
                                Execute Payouts
                            </button>
                        </div>

                        {/* Reset Competition Card */}
                        <div className="p-6 rounded-2xl bg-gray-900 border border-gray-800 flex flex-col justify-between">
                            <div className="mb-6">
                                <div className="flex items-center gap-3 mb-2">
                                    <div className="p-2 bg-amber-500/10 rounded-lg">
                                        <RefreshCcw className="w-5 h-5 text-amber-500" />
                                    </div>
                                    <h3 className="font-bold text-lg">Reset Competition</h3>
                                </div>
                                <p className="text-sm text-gray-400">Clear AMM pools and refund users. Usually used for testing.</p>
                            </div>
                            <button className="w-full py-3 px-4 rounded-xl bg-amber-950/50 hover:bg-amber-900/50 border border-amber-900/50 text-amber-400 font-semibold transition-colors flex items-center justify-center gap-2">
                                <RefreshCcw className="w-4 h-4" />
                                Reset Data
                            </button>
                        </div>
                    </div>
                )}

                {activeTab === 'create' && (
                    <div className="p-8 rounded-2xl bg-gray-900 border border-gray-800 max-w-2xl">
                        <h2 className="text-xl font-bold mb-6 flex items-center gap-2">
                            <Plus className="w-5 h-5 text-indigo-400" />
                            Initialize New Competition
                        </h2>

                        <form className="space-y-6" onSubmit={async (e) => {
                            e.preventDefault();
                            if (!compName || !file) {
                                setCreateMsg({ text: "Please provide a name and an Excel file.", type: "error" });
                                return;
                            }

                            setCreating(true);
                            setCreateMsg(null);

                            const formData = new FormData();
                            formData.append("name", compName);
                            formData.append("file", file);

                            try {
                                const res = await fetch("http://localhost:8080/api/v1/admin/competitions", {
                                    method: "POST",
                                    body: formData
                                });
                                const data = await res.json();
                                if (res.ok) {
                                    setCreateMsg({ text: `Success! Imported ${data.teams_imported} teams.`, type: "success" });
                                    setCompName('');
                                    setFile(null);
                                    // Reset file input
                                    const fileInput = document.getElementById('team-file') as HTMLInputElement;
                                    if (fileInput) fileInput.value = '';
                                } else {
                                    setCreateMsg({ text: data.error || "Failed to create.", type: "error" });
                                }
                            } catch (err) {
                                setCreateMsg({ text: "Server error occurred.", type: "error" });
                            } finally {
                                setCreating(false);
                            }
                        }}>
                            {createMsg && (
                                <div className={`p-4 rounded-xl border ${createMsg.type === 'success' ? 'bg-green-500/10 border-green-500/20 text-green-400' : 'bg-red-500/10 border-red-500/20 text-red-400'}`}>
                                    {createMsg.text}
                                </div>
                            )}

                            <div className="space-y-2">
                                <label className="text-sm font-medium text-gray-300">Competition Name / ID</label>
                                <input
                                    type="text"
                                    value={compName}
                                    onChange={(e) => setCompName(e.target.value)}
                                    placeholder="e.g. vex-worlds-2024"
                                    className="w-full bg-gray-950 border border-gray-800 rounded-xl px-4 py-3 text-white placeholder:text-gray-600 focus:outline-none focus:border-indigo-500 focus:ring-1 focus:ring-indigo-500 transition-all"
                                />
                            </div>

                            <div className="space-y-2">
                                <label className="text-sm font-medium text-gray-300">Team List (XLS)</label>
                                <div className="p-4 border border-dashed border-gray-700 bg-gray-950/50 rounded-xl text-center">
                                    <input
                                        id="team-file"
                                        type="file"
                                        accept=".xls"
                                        onChange={(e) => setFile(e.target.files?.[0] || null)}
                                        className="text-sm text-gray-400 file:mr-4 file:py-2 file:px-4 file:rounded-lg file:border-0 file:text-sm file:font-semibold file:bg-indigo-500/10 file:text-indigo-400 hover:file:bg-indigo-500/20"
                                    />
                                </div>
                                <p className="text-xs text-gray-500">Must include Team, Team Name, Division, Organization, Location columns.</p>
                            </div>

                            <button
                                type="submit"
                                disabled={creating}
                                className="w-full bg-indigo-600 hover:bg-indigo-500 text-white font-semibold py-3 px-4 rounded-xl transition-all shadow-lg shadow-indigo-500/20 active:scale-[0.98] disabled:opacity-50 disabled:cursor-not-allowed"
                            >
                                {creating ? "Processing..." : "Create Event Scaffold"}
                            </button>
                        </form>
                    </div>
                )}
            </div>
        </div>
    );
}

export default Admin;

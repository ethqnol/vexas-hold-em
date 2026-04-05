import { useState, useEffect } from 'react';
import { Link } from 'react-router-dom';
import { ArrowLeft, Trophy, TrendingUp } from 'lucide-react';
import type { User } from 'firebase/auth';
import { getRank, formatNumber } from './utils/ranks';

interface LeaderboardProps {
    user: User | null;
}

interface Entry {
    rank: number;
    userId: string;
    displayName: string;
    balance: number;
    totalLost: number;
    title: string;
}

const MEDAL: Record<number, string> = { 1: '🥇', 2: '🥈', 3: '🥉' };

export default function Leaderboard({ user }: LeaderboardProps) {
    const [entries, setEntries] = useState<Entry[]>([]);
    const [loading, setLoading] = useState(true);
    const [tab, setTab] = useState<'balance' | 'lost'>('balance');

    const API = import.meta.env.VITE_API_URL;

    useEffect(() => {
        fetch(`${API}/api/v1/leaderboard`)
            .then(r => r.json())
            .then(d => { setEntries(d.leaderboard ?? []); setLoading(false); });
    }, []);

    const sorted = [...entries].sort((a, b) =>
        tab === 'balance' ? b.balance - a.balance : b.totalLost - a.totalLost
    ).map((e, i) => ({ ...e, rank: i + 1 }));

    return (
        <div className="min-h-screen bg-[#08080f] text-gray-100 flex flex-col pt-4">
            <header className="max-w-4xl mx-auto w-full px-4 mb-8 flex items-center gap-4">
                <Link to="/" className="p-2 rounded-xl bg-gray-900 border border-gray-800 hover:bg-gray-800 transition-colors">
                    <ArrowLeft className="w-5 h-5 text-gray-400" />
                </Link>
                <div className="flex items-center gap-3">
                    <div className="p-2.5 bg-yellow-500/10 rounded-xl border border-yellow-500/20">
                        <Trophy className="w-6 h-6 text-yellow-400" />
                    </div>
                    <h1 className="text-2xl font-bold text-white">Leaderboard</h1>
                </div>
            </header>

            <main className="flex-1 max-w-4xl mx-auto px-4 w-full pb-12">
                {/* tabs */}
                <div className="flex gap-2 mb-6 p-1 bg-gray-900 rounded-xl border border-gray-800 w-fit">
                    <button
                        onClick={() => setTab('balance')}
                        className={`px-5 py-2 rounded-lg text-sm font-semibold transition-all ${tab === 'balance' ? 'bg-yellow-500/20 text-yellow-300 border border-yellow-500/30' : 'text-gray-500 hover:text-gray-300'}`}
                    >
                        <TrendingUp className="w-4 h-4 inline mr-1.5 -mt-0.5" />
                        Richest
                    </button>
                    <button
                        onClick={() => setTab('lost')}
                        className={`px-5 py-2 rounded-lg text-sm font-semibold transition-all ${tab === 'lost' ? 'bg-red-500/20 text-red-300 border border-red-500/30' : 'text-gray-500 hover:text-gray-300'}`}
                    >
                        💀 Most Lost
                    </button>
                </div>

                {loading ? (
                    <div className="flex items-center justify-center h-48">
                        <div className="w-8 h-8 border-2 border-yellow-500/30 border-t-yellow-400 rounded-full animate-spin" />
                    </div>
                ) : (
                    <div className="space-y-2">
                        {sorted.map(entry => {
                            const rank = getRank(entry.totalLost);
                            const isMe = user?.uid === entry.userId;
                            return (
                                <div
                                    key={entry.userId}
                                    className={`flex items-center gap-4 p-4 rounded-2xl border transition-all ${
                                        isMe
                                            ? 'border-indigo-400/50 bg-indigo-900/10 shadow-lg shadow-indigo-500/5'
                                            : 'border-gray-800 bg-gray-900/40 hover:border-gray-700'
                                    }`}
                                >
                                    {/* rank number */}
                                    <div className="w-8 text-center font-black text-lg">
                                        {MEDAL[entry.rank] ?? (
                                            <span className="text-gray-600 text-sm">#{entry.rank}</span>
                                        )}
                                    </div>

                                    {/* name + badges */}
                                    <div className="flex-1 min-w-0">
                                        <div className="flex items-center gap-2 flex-wrap">
                                            <span className={`font-bold truncate ${isMe ? 'text-indigo-300' : 'text-white'}`}>
                                                {entry.displayName || 'Anonymous'}
                                            </span>
                                            {isMe && (
                                                <span className="text-xs px-2 py-0.5 rounded-full bg-indigo-500/20 text-indigo-300 border border-indigo-500/30">you</span>
                                            )}
                                            {entry.title && (
                                                <span className="text-xs px-2 py-0.5 rounded-full bg-purple-500/10 text-purple-300 border border-purple-500/20">
                                                    {entry.title}
                                                </span>
                                            )}
                                        </div>
                                        {/* rank badge */}
                                        <div className={`inline-flex items-center gap-1 mt-1 text-xs px-2 py-0.5 rounded-full ${rank.bgClass} border ${rank.borderClass}`}>
                                            <span>{rank.emoji}</span>
                                            <span className={rank.colorClass}>{rank.name}</span>
                                        </div>
                                    </div>

                                    {/* stats */}
                                    <div className="text-right flex-shrink-0">
                                        <div className={`font-mono font-bold ${tab === 'balance' ? 'text-green-400' : 'text-gray-400'}`}>
                                            {formatNumber(entry.balance)}
                                        </div>
                                        <div className={`font-mono text-xs ${tab === 'lost' ? 'text-red-400' : 'text-gray-600'}`}>
                                            −{formatNumber(entry.totalLost)}
                                        </div>
                                    </div>
                                </div>
                            );
                        })}

                        {sorted.length === 0 && (
                            <div className="text-center py-16 text-gray-600">
                                <Trophy className="w-12 h-12 mx-auto mb-4 opacity-20" />
                                <p>No players yet</p>
                            </div>
                        )}
                    </div>
                )}
            </main>
        </div>
    );
}

import { useState, useEffect } from 'react';
import { Link } from 'react-router-dom';
import { ArrowLeft, Droplets, Trophy } from 'lucide-react';
import type { User } from 'firebase/auth';

interface LotteryProps {
    user: User | null;
    profile: any;
    refreshProfile: () => void;
}

export default function Lottery({ user, profile, refreshProfile }: LotteryProps) {
    const [jackpot, setJackpot] = useState<number>(100);
    const [lastWinner, setLastWinner] = useState<string>('');
    const [lastWonAt, setLastWonAt] = useState<number>(0);
    const [loading, setLoading] = useState(false);
    const [result, setResult] = useState<{ won: boolean; message: string; jackpot: number } | null>(null);
    const [ripples, setRipples] = useState<number[]>([]);

    const API = import.meta.env.VITE_API_URL;

    const fetchStatus = async () => {
        const res = await fetch(`${API}/api/v1/casino/lottery`);
        if (res.ok) {
            const data = await res.json();
            setJackpot(data.jackpot ?? 100);
            setLastWinner(data.lastWinnerName ?? '');
            setLastWonAt(data.lastWonAt ?? 0);
        }
    };

    useEffect(() => { fetchStatus(); }, []);

    const buyTicket = async () => {
        if (!user || loading) return;
        setLoading(true);
        setResult(null);

        // water ripple effect
        setRipples(r => [...r, Date.now()]);
        setTimeout(() => setRipples(r => r.slice(1)), 1500);

        try {
            const res = await fetch(`${API}/api/v1/casino/lottery`, {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ userId: user.uid }),
            });
            const data = await res.json();
            if (!res.ok) throw new Error(data.error);

            const won = data.result === 'WIN';
            setResult({ won, message: data.message, jackpot: data.jackpot });
            refreshProfile();
            fetchStatus();
        } catch (e: any) {
            setResult({ won: false, message: e.message, jackpot: jackpot });
        } finally {
            setLoading(false);
        }
    };

    const formatJackpot = (n: number) => {
        if (n >= 1_000_000) return (n / 1_000_000).toFixed(2) + 'M';
        if (n >= 1_000) return (n / 1_000).toFixed(1) + 'K';
        return n.toFixed(2);
    };

    return (
        <div className="min-h-screen bg-[#070b12] text-gray-100 flex flex-col pt-4 relative overflow-hidden">
            {/* animated water bg */}
            <div className="pointer-events-none absolute inset-0 overflow-hidden">
                <div className="absolute inset-0 bg-gradient-to-b from-blue-950/30 via-transparent to-cyan-950/20" />
                {[...Array(6)].map((_, i) => (
                    <div
                        key={i}
                        className="absolute left-0 right-0 opacity-5"
                        style={{
                            bottom: `${i * 18}%`,
                            height: '120px',
                            background: `radial-gradient(ellipse 120% 60px at 50% 50%, rgba(56,189,248,0.4), transparent)`,
                            animation: `wave ${3 + i * 0.7}s ease-in-out ${i * 0.4}s infinite alternate`,
                        }}
                    />
                ))}
            </div>

            {/* click ripples */}
            {ripples.map(key => (
                <div key={key} className="pointer-events-none absolute inset-0 flex items-center justify-center">
                    <div className="w-32 h-32 rounded-full border-2 border-cyan-400/60 animate-ping" />
                </div>
            ))}

            <style>{`
                @keyframes wave {
                    from { transform: translateX(-5%) scaleY(1); }
                    to   { transform: translateX(5%)  scaleY(1.3); }
                }
            `}</style>

            <header className="max-w-4xl mx-auto w-full px-4 mb-8 flex items-center gap-4 relative z-10">
                <Link to="/casino" className="p-2 rounded-xl bg-gray-900 border border-gray-800 hover:bg-gray-800 transition-colors">
                    <ArrowLeft className="w-5 h-5 text-gray-400" />
                </Link>
                <div className="flex items-center gap-3">
                    <div className="p-2.5 bg-cyan-500/10 rounded-xl border border-cyan-500/20">
                        <Droplets className="w-6 h-6 text-cyan-400" />
                    </div>
                    <div>
                        <h1 className="text-2xl font-bold text-white">Water Game Lottery</h1>
                        <p className="text-xs text-cyan-400/70">the water game is real. probably.</p>
                    </div>
                </div>
            </header>

            <main className="flex-1 max-w-4xl mx-auto px-4 w-full space-y-6 pb-12 relative z-10">
                {/* jackpot display */}
                <div className="rounded-3xl border border-cyan-500/20 bg-gray-900/60 backdrop-blur-sm p-10 text-center relative overflow-hidden">
                    <div className="absolute inset-0 bg-gradient-to-b from-cyan-500/5 to-transparent" />
                    <div className="text-sm font-semibold text-cyan-400/60 uppercase tracking-widest mb-2 relative">Current Jackpot</div>
                    <div className="text-7xl font-black font-mono text-cyan-300 mb-2 relative" style={{ textShadow: '0 0 40px rgba(34,211,238,0.5)' }}>
                        {formatJackpot(jackpot)}
                    </div>
                    <div className="text-lg text-cyan-400/50 relative">S.H.I.T. Coins</div>
                </div>

                {/* info row */}
                <div className="grid grid-cols-3 gap-4 text-center">
                    <div className="p-4 rounded-2xl bg-gray-900/60 border border-gray-800">
                        <div className="text-2xl font-bold text-white">10</div>
                        <div className="text-xs text-gray-400 mt-1">VEX per ticket</div>
                    </div>
                    <div className="p-4 rounded-2xl bg-gray-900/60 border border-gray-800">
                        <div className="text-2xl font-bold text-yellow-500">1:1,000,000</div>
                        <div className="text-xs text-gray-400 mt-1">odds per ticket</div>
                    </div>
                    <div className="p-4 rounded-2xl bg-gray-900/60 border border-gray-800">
                        <div className="text-2xl font-bold text-green-400">80%</div>
                        <div className="text-xs text-gray-400 mt-1">goes to jackpot</div>
                    </div>
                </div>

                {/* buy button */}
                {profile && (
                    <div className="text-center space-y-4">
                        <div className="text-sm text-gray-400">
                            Your balance: <span className="text-green-400 font-mono font-bold">{profile.Balance?.toFixed(2)} S.H.I.T.</span>
                        </div>
                        <button
                            onClick={buyTicket}
                            disabled={loading || !profile || profile.Balance < 10}
                            className="px-12 py-5 bg-cyan-600 hover:bg-cyan-500 disabled:bg-gray-800 disabled:text-gray-600 text-white font-black text-xl rounded-2xl transition-all hover:scale-105 active:scale-95 shadow-2xl shadow-cyan-500/20 disabled:cursor-not-allowed"
                        >
                            {loading ? (
                                <span className="flex items-center gap-2 justify-center">
                                    <span className="w-5 h-5 border-2 border-white/30 border-t-white rounded-full animate-spin" />
                                    Drawing…
                                </span>
                            ) : 'Buy Ticket — 10 S.H.I.T.'}
                        </button>
                        {profile.Balance < 10 && (
                            <p className="text-red-400 text-sm">insufficient balance</p>
                        )}
                    </div>
                )}

                {/* result */}
                {result && (
                    <div className={`p-8 rounded-3xl border text-center transition-all ${result.won
                        ? 'bg-cyan-900/30 border-cyan-400/50 shadow-2xl shadow-cyan-400/20'
                        : 'bg-gray-900/60 border-gray-700'
                        }`}>
                        {result.won && (
                            <div className="text-6xl mb-4 animate-bounce">🌊</div>
                        )}
                        <p className={`font-bold text-xl mb-2 ${result.won ? 'text-cyan-300' : 'text-gray-400'}`}>
                            {result.won ? `YOU WON ${formatJackpot(result.jackpot)} S.H.I.T.!` : 'not a winner'}
                        </p>
                        <p className="text-gray-500 text-sm italic">{result.message}</p>
                    </div>
                )}

                {/* last winner */}
                {lastWinner && (
                    <div className="p-5 rounded-2xl bg-gray-900/40 border border-yellow-500/20 flex items-center gap-4">
                        <Trophy className="w-5 h-5 text-yellow-400 flex-shrink-0" />
                        <div>
                            <div className="text-sm font-semibold text-yellow-300">{lastWinner} won the jackpot</div>
                            <div className="text-xs text-gray-500">
                                {lastWonAt ? new Date(lastWonAt).toLocaleDateString() : ''}
                            </div>
                        </div>
                    </div>
                )}

                <p className="text-center text-xs text-gray-600 italic">
                    "The Water Game" refers to an aquatic VEX challenge that has been rumored for years but never materialized.
                    Like the lottery jackpot, you can get very close but never quite there.
                </p>
            </main>
        </div>
    );
}

import { useState } from 'react';
import { Link } from 'react-router-dom';
import { ArrowLeft, Dices, Loader2 } from 'lucide-react';
import type { User } from 'firebase/auth';

interface RouletteProps {
    user: User | null;
    profile: any;
    refreshProfile: () => void;
}

export default function Roulette({ user, profile, refreshProfile }: RouletteProps) {
    const [betType, setBetType] = useState('red');
    const [amount, setAmount] = useState('');
    const [spinning, setSpinning] = useState(false);
    const [result, setResult] = useState<{ spin: number, color: string, payout: number } | null>(null);
    const [error, setError] = useState<string | null>(null);

    const handleSpin = async () => {
        if (!user || !amount) return;
        setSpinning(true);
        setResult(null);
        setError(null);

        try {
            const res = await fetch(`${import.meta.env.VITE_API_URL}/api/v1/casino/roulette`, {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({
                    userId: user.uid,
                    betType,
                    amount: parseFloat(amount)
                })
            });
            const data = await res.json();
            if (!res.ok) throw new Error(data.error || 'Spin failed');

            await new Promise(resolve => setTimeout(resolve, 1500));
            setResult(data);
            refreshProfile();
        } catch (e: any) {
            setError(e.message);
        } finally {
            setSpinning(false);
        }
    };

    if (!user) return null;

    const isWin = result && result.payout > 0 && !spinning;

    return (
        <div className="min-h-screen bg-[#0a0a0a] text-gray-100 flex flex-col pt-4">
            <header className="max-w-3xl mx-auto w-full px-4 sm:px-6 lg:px-8 mb-8 flex items-center gap-4">
                <Link to="/casino" className="p-2 rounded-xl bg-white/[0.03] border border-white/[0.06] hover:bg-white/[0.06] transition-colors">
                    <ArrowLeft className="w-5 h-5 text-gray-400" />
                </Link>
                <div className="flex items-center gap-3">
                    <div className="p-2.5 bg-red-500/10 rounded-xl border border-red-500/20">
                        <Dices className="w-6 h-6 text-red-500" />
                    </div>
                    <h1 className="text-3xl font-bold bg-gradient-to-r from-red-400 to-orange-500 bg-clip-text text-transparent">Roulette</h1>
                </div>
                <div className="ml-auto flex items-center gap-2">
                    <span className="text-[10px] text-gray-500 font-bold uppercase tracking-widest">Balance</span>
                    <span className="text-lg font-mono text-green-400 font-bold">{profile?.Balance?.toFixed(2) || "0.00"}</span>
                </div>
            </header>

            <main className="flex-1 max-w-3xl mx-auto px-4 sm:px-6 lg:px-8 w-full pb-12">

                {/* ═══ the machine ═══ */}
                <div className={`relative rounded-[2rem] overflow-hidden transition-shadow duration-700 ${isWin ? 'shadow-[0_0_80px_rgba(239,68,68,0.12)]' : 'shadow-2xl shadow-black/50'}`}>
                    {/* red trim */}
                    <div className="absolute inset-0 rounded-[2rem] bg-gradient-to-b from-red-600/25 via-red-900/10 to-red-900/25 p-[1.5px]">
                        <div className="w-full h-full rounded-[2rem] bg-[#0e0e0e]"></div>
                    </div>

                    <div className="relative z-10 p-6 sm:p-10">
                        {/* title */}
                        <div className="text-center mb-8">
                            <h2 className="text-[10px] font-black uppercase tracking-[0.4em] text-red-500/40">♠ ♥ ♦ ♣  ROULETTE  ♣ ♦ ♥ ♠</h2>
                        </div>

                        {/* ── wheel result ── */}
                        <div className="mx-auto max-w-sm mb-8">
                            <div className={`rounded-2xl p-[1.5px] transition-all duration-700 ${isWin ? 'bg-gradient-to-r from-green-500/50 via-emerald-400/60 to-green-500/50' : 'bg-white/[0.06]'}`}>
                                <div className="bg-[#080808] rounded-[14px] p-8 sm:p-10 flex flex-col items-center justify-center min-h-[220px] relative overflow-hidden">
                                    {spinning ? (
                                        <div className="flex flex-col items-center">
                                            <Loader2 className="w-16 h-16 text-red-500 animate-spin mb-3" />
                                            <span className="text-xs text-gray-600 uppercase tracking-widest font-bold animate-pulse">Spinning...</span>
                                        </div>
                                    ) : result ? (
                                        <div className="flex flex-col items-center text-center">
                                            <div className="text-[10px] text-gray-500 uppercase tracking-widest font-bold mb-4">The ball landed on</div>
                                            <div className={`text-7xl sm:text-8xl font-black font-mono w-32 h-32 sm:w-36 sm:h-36 rounded-full flex items-center justify-center shadow-2xl mb-4 ${result.color === 'red' ? 'bg-red-500 text-white shadow-red-500/30' :
                                                    result.color === 'black' ? 'bg-[#1a1a1a] text-white border-2 border-white/10 shadow-black/50' :
                                                        'bg-green-500 text-white shadow-green-500/30'
                                                }`}>
                                                {result.spin}
                                            </div>
                                        </div>
                                    ) : (
                                        <div className="flex flex-col items-center opacity-30">
                                            <div className="w-28 h-28 rounded-full border-4 border-dashed border-gray-700 flex items-center justify-center mb-4">
                                                <Dices className="w-10 h-10 text-gray-600" />
                                            </div>
                                            <span className="text-xs text-gray-600 font-bold uppercase tracking-widest">Place your bet</span>
                                        </div>
                                    )}
                                </div>
                            </div>
                        </div>

                        {/* ── outcome ── */}
                        <div className="h-16 max-w-sm mx-auto flex items-center justify-center mb-6">
                            {isWin && (
                                <div className="w-full text-center">
                                    <h3 className="text-green-400 font-black text-xl tracking-[0.25em] uppercase">💰 WINNER</h3>
                                    <p className="font-mono text-2xl font-bold text-white mt-0.5">+{result!.payout.toFixed(2)} <span className="text-green-500/50 text-sm">S.H.I.T.</span></p>
                                </div>
                            )}
                            {result && result.payout === 0 && !spinning && (
                                <span className="text-gray-600 text-xs font-bold uppercase tracking-[0.15em]">Better Luck Next Spin</span>
                            )}
                        </div>

                        {/* ── bet selector ── */}
                        <div className="max-w-sm mx-auto mb-6">
                            <div className="text-[10px] font-bold text-gray-600 uppercase tracking-widest mb-3">Choose Bet</div>
                            <div className="grid grid-cols-3 gap-2 mb-3">
                                <button onClick={() => setBetType('red')} className={`py-3.5 rounded-xl font-bold text-sm transition-all ${betType === 'red' ? 'bg-red-500 text-white shadow-lg shadow-red-500/20 ring-1 ring-white/20' : 'bg-white/[0.03] text-red-500/70 border border-white/[0.04] hover:bg-white/[0.06]'}`}>
                                    RED<br /><span className="text-[10px] opacity-60 font-normal">2×</span>
                                </button>
                                <button onClick={() => setBetType('black')} className={`py-3.5 rounded-xl font-bold text-sm transition-all ${betType === 'black' ? 'bg-[#1a1a1a] text-white shadow-lg ring-1 ring-white/20' : 'bg-white/[0.03] text-gray-400 border border-white/[0.04] hover:bg-white/[0.06]'}`}>
                                    BLACK<br /><span className="text-[10px] opacity-60 font-normal">2×</span>
                                </button>
                                <button onClick={() => setBetType('green')} className={`py-3.5 rounded-xl font-bold text-sm transition-all ${betType === 'green' ? 'bg-green-500 text-white shadow-lg shadow-green-500/20 ring-1 ring-white/20' : 'bg-white/[0.03] text-green-500/70 border border-white/[0.04] hover:bg-white/[0.06]'}`}>
                                    GREEN<br /><span className="text-[10px] opacity-60 font-normal">36×</span>
                                </button>
                            </div>
                            <div className="grid grid-cols-2 gap-2">
                                <button onClick={() => setBetType('even')} className={`py-2.5 rounded-xl font-bold text-sm transition-all ${betType === 'even' ? 'bg-blue-600 text-white shadow-lg ring-1 ring-white/20' : 'bg-white/[0.03] text-gray-500 border border-white/[0.04] hover:bg-white/[0.06]'}`}>
                                    EVEN <span className="text-[10px] opacity-60 font-normal">2×</span>
                                </button>
                                <button onClick={() => setBetType('odd')} className={`py-2.5 rounded-xl font-bold text-sm transition-all ${betType === 'odd' ? 'bg-blue-600 text-white shadow-lg ring-1 ring-white/20' : 'bg-white/[0.03] text-gray-500 border border-white/[0.04] hover:bg-white/[0.06]'}`}>
                                    ODD <span className="text-[10px] opacity-60 font-normal">2×</span>
                                </button>
                            </div>
                        </div>

                        {/* ── controls ── */}
                        <div className="max-w-sm mx-auto flex gap-3 items-end">
                            <div className="flex-1">
                                <div className="flex items-center gap-2 mb-2">
                                    <span className="text-[10px] font-bold text-gray-600 uppercase tracking-widest">Wager</span>
                                    <div className="flex gap-1 ml-auto">
                                        {[5, 10, 25, 50, 67, 100].map(v => (
                                            <button key={v} onClick={() => setAmount(String(v))} className={`px-2 py-0.5 rounded text-[10px] font-bold transition-all ${amount === String(v) ? 'bg-red-500/20 text-red-400 border border-red-500/30' : 'bg-white/[0.03] text-gray-600 hover:text-gray-400 border border-white/[0.04]'}`}>{v}</button>
                                        ))}
                                    </div>
                                </div>
                                <input
                                    type="number"
                                    value={amount}
                                    onChange={(e) => setAmount(e.target.value)}
                                    placeholder="0"
                                    className="w-full bg-white/[0.02] border border-white/[0.06] rounded-xl px-4 py-3.5 text-white font-mono text-xl focus:outline-none focus:border-red-500/30 transition-colors text-center placeholder:text-gray-700"
                                />
                            </div>
                            <button
                                onClick={handleSpin}
                                disabled={spinning || !amount || parseFloat(amount) <= 0}
                                className="h-[58px] px-8 bg-gradient-to-b from-red-500 to-red-700 hover:from-red-400 hover:to-red-600 text-white font-black text-base tracking-widest uppercase rounded-xl transition-all shadow-lg shadow-red-600/20 hover:shadow-red-500/30 hover:scale-[1.03] active:scale-[0.97] disabled:opacity-20 disabled:grayscale disabled:hover:scale-100 disabled:cursor-not-allowed shrink-0"
                            >
                                {spinning ? '...' : 'SPIN'}
                            </button>
                        </div>

                        {error && <div className="text-red-400 text-xs font-medium px-3 py-2 bg-red-500/10 border border-red-500/20 rounded-lg max-w-sm mx-auto text-center mt-4">{error}</div>}
                    </div>
                </div>

            </main>
        </div>
    );
}

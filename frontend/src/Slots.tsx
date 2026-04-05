import { useState } from 'react';
import { Link } from 'react-router-dom';
import { ArrowLeft, Cherry } from 'lucide-react';
import type { User } from 'firebase/auth';
import jasonImg from './assets/jason.png';
import charlesImg from './assets/charles.png';

const SYMBOL_MAP: Record<string, React.ReactNode> = {
    'jason': <img src={jasonImg} alt="Jason" className="w-16 h-16 sm:w-20 sm:h-20 object-cover rounded-xl border border-white/10" />,
    'charles': <img src={charlesImg} alt="Charles" className="w-16 h-16 sm:w-20 sm:h-20 object-cover rounded-xl border border-white/10" />,
};

interface SlotsProps {
    user: User | null;
    profile: any;
    refreshProfile: () => void;
}

export default function Slots({ user, profile, refreshProfile }: SlotsProps) {
    const [amount, setAmount] = useState('');
    const [spinning, setSpinning] = useState(false);
    const [result, setResult] = useState<{ reels: string[], payout: number } | null>(null);
    const [error, setError] = useState<string | null>(null);

    const handleSpin = async () => {
        if (!user || !amount) return;
        setSpinning(true);
        setResult(null);
        setError(null);

        try {
            const res = await fetch(`${import.meta.env.VITE_API_URL}/api/v1/casino/slots`, {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({
                    userId: user.uid,
                    amount: parseFloat(amount)
                })
            });
            const data = await res.json();
            if (!res.ok) throw new Error(data.error || 'Spin failed');

            // artificial delay for tension
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
    const isJackpot = result && result.reels[0] === result.reels[1] && result.reels[1] === result.reels[2];

    return (
        <div className="min-h-screen bg-[#0a0a0a] text-gray-100 flex flex-col pt-4">
            <header className="max-w-3xl mx-auto w-full px-4 sm:px-6 lg:px-8 mb-8 flex items-center gap-4">
                <Link to="/casino" className="p-2 rounded-xl bg-white/[0.03] border border-white/[0.06] hover:bg-white/[0.06] transition-colors">
                    <ArrowLeft className="w-5 h-5 text-gray-400" />
                </Link>
                <div className="flex items-center gap-3">
                    <div className="p-2.5 bg-yellow-500/10 rounded-xl border border-yellow-500/20">
                        <Cherry className="w-6 h-6 text-yellow-500" />
                    </div>
                    <h1 className="text-3xl font-bold bg-gradient-to-r from-yellow-400 to-amber-500 bg-clip-text text-transparent">VEX Slots</h1>
                </div>
                <div className="ml-auto flex items-center gap-2">
                    <span className="text-[10px] text-gray-500 font-bold uppercase tracking-widest">Balance</span>
                    <span className="text-lg font-mono text-green-400 font-bold">{profile?.Balance?.toFixed(2) || "0.00"}</span>
                </div>
            </header>

            <main className="flex-1 max-w-3xl mx-auto px-4 sm:px-6 lg:px-8 w-full pb-12">

                {/* ═══ machine body ═══ */}
                <div className={`relative rounded-[2rem] overflow-hidden transition-shadow duration-700 ${isWin ? 'shadow-[0_0_80px_rgba(234,179,8,0.15)]' : 'shadow-2xl shadow-black/50'}`}>
                    {/* gold trim */}
                    <div className="absolute inset-0 rounded-[2rem] bg-gradient-to-b from-yellow-600/30 via-yellow-800/15 to-yellow-900/30 p-[1.5px]">
                        <div className="w-full h-full rounded-[2rem] bg-[#0e0e0e]"></div>
                    </div>

                    <div className="relative z-10 p-6 sm:p-10">
                        {/* title */}
                        <div className="text-center mb-8">
                            <h2 className="text-[10px] font-black uppercase tracking-[0.4em] text-yellow-600/50">★ ★ ★  VEX SLOTS  ★ ★ ★</h2>
                        </div>

                        {/* ── reel window ── */}
                        <div className={`mx-auto max-w-md rounded-2xl p-[1.5px] mb-6 transition-all duration-700 ${isWin ? 'bg-gradient-to-r from-yellow-500/50 via-amber-400/70 to-yellow-500/50' : 'bg-white/[0.06]'}`}>
                            <div className="bg-[#080808] rounded-[14px] p-5 sm:p-6 relative overflow-hidden">
                                {/* scanlines */}
                                <div className="absolute inset-0 pointer-events-none opacity-[0.02]" style={{ backgroundImage: 'repeating-linear-gradient(0deg, transparent, transparent 2px, white 2px, white 3px)' }}></div>

                                <div className="flex items-stretch justify-center h-28 sm:h-36 relative">
                                    {spinning ? (
                                        [0, 1, 2].map(i => (
                                            <div key={i} className="flex-1 flex items-center justify-center relative">
                                                {i > 0 && <div className="absolute left-0 top-3 bottom-3 w-px bg-white/[0.04]"></div>}
                                                <div className="text-6xl sm:text-7xl animate-bounce opacity-40 blur-[3px]" style={{ animationDelay: `${i * 100}ms` }}>❓</div>
                                            </div>
                                        ))
                                    ) : result ? (
                                        result.reels.map((symbol: string, i: number) => (
                                            <div key={i} className="flex-1 flex items-center justify-center relative">
                                                {i > 0 && <div className="absolute left-0 top-3 bottom-3 w-px bg-white/[0.04]"></div>}
                                                <div className="drop-shadow-[0_0_12px_rgba(255,255,255,0.1)]">
                                                    {SYMBOL_MAP[symbol] || <span className="text-6xl sm:text-7xl">{symbol}</span>}
                                                </div>
                                            </div>
                                        ))
                                    ) : (
                                        [0, 1, 2].map(i => (
                                            <div key={i} className="flex-1 flex items-center justify-center relative">
                                                {i > 0 && <div className="absolute left-0 top-3 bottom-3 w-px bg-white/[0.04]"></div>}
                                                <div className="text-6xl sm:text-7xl opacity-15 grayscale">🍒</div>
                                            </div>
                                        ))
                                    )}
                                </div>
                            </div>
                        </div>

                        {/* ── outcome ── */}
                        <div className="h-16 max-w-md mx-auto flex items-center justify-center mb-6">
                            {isWin && (
                                <div className="w-full text-center">
                                    <h3 className="text-yellow-400 font-black text-xl tracking-[0.25em] uppercase">
                                        {isJackpot ? '🎉 JACKPOT' : '💰 WINNER'}
                                    </h3>
                                    <p className="font-mono text-2xl font-bold text-white mt-0.5">+{result!.payout.toFixed(2)} <span className="text-yellow-500/50 text-sm">S.H.I.T.</span></p>
                                </div>
                            )}
                            {result && result.payout === 0 && !spinning && (
                                <span className="text-gray-600 text-xs font-bold uppercase tracking-[0.15em]">No Match — Try Again</span>
                            )}
                        </div>

                        {/* ── controls ── */}
                        <div className="max-w-md mx-auto flex gap-3 items-end mb-8">
                            <div className="flex-1">
                                <div className="flex items-center gap-2 mb-2">
                                    <span className="text-[10px] font-bold text-gray-600 uppercase tracking-widest">Wager</span>
                                    <div className="flex gap-1 ml-auto">
                                        {[5, 10, 25, 50, 67, 100].map(v => (
                                            <button key={v} onClick={() => setAmount(String(v))} className={`px-2 py-0.5 rounded text-[10px] font-bold transition-all ${amount === String(v) ? 'bg-yellow-500/20 text-yellow-400 border border-yellow-500/30' : 'bg-white/[0.03] text-gray-600 hover:text-gray-400 border border-white/[0.04]'}`}>{v}</button>
                                        ))}
                                    </div>
                                </div>
                                <input
                                    type="number"
                                    value={amount}
                                    onChange={(e) => setAmount(e.target.value)}
                                    placeholder="0"
                                    className="w-full bg-white/[0.02] border border-white/[0.06] rounded-xl px-4 py-3.5 text-white font-mono text-xl focus:outline-none focus:border-yellow-500/30 transition-colors text-center placeholder:text-gray-700"
                                />
                            </div>
                            <button
                                onClick={handleSpin}
                                disabled={spinning || !amount || parseFloat(amount) <= 0}
                                className="h-[58px] px-8 bg-gradient-to-b from-yellow-500 to-amber-600 hover:from-yellow-400 hover:to-amber-500 text-black font-black text-base tracking-widest uppercase rounded-xl transition-all shadow-lg shadow-yellow-600/20 hover:shadow-yellow-500/30 hover:scale-[1.03] active:scale-[0.97] disabled:opacity-20 disabled:grayscale disabled:hover:scale-100 disabled:cursor-not-allowed shrink-0"
                            >
                                {spinning ? '...' : 'SPIN'}
                            </button>
                        </div>

                        {error && <div className="text-red-400 text-xs font-medium px-3 py-2 bg-red-500/10 border border-red-500/20 rounded-lg max-w-md mx-auto text-center mb-4">{error}</div>}

                        {/* ── payout table (inside machine) ── */}
                        <div className="max-w-md mx-auto border-t border-white/[0.04] pt-6">
                            <div className="text-[10px] font-bold text-gray-600 uppercase tracking-widest mb-3">Payouts</div>
                            <div className="space-y-1.5 text-xs">
                                <div className="flex justify-between items-center px-3 py-2 rounded-lg bg-white/[0.02]">
                                    <span className="flex gap-1.5 items-center">
                                        <img src={jasonImg} alt="Jason" className="w-6 h-6 object-cover rounded" />
                                        <img src={jasonImg} alt="Jason" className="w-6 h-6 object-cover rounded" />
                                        <img src={jasonImg} alt="Jason" className="w-6 h-6 object-cover rounded" />
                                        <span className="text-gray-500 ml-2">Jason</span>
                                    </span>
                                    <span className="text-yellow-500 font-bold font-mono">50×</span>
                                </div>
                                <div className="flex justify-between items-center px-3 py-2 rounded-lg bg-white/[0.02]">
                                    <span className="flex gap-1.5 items-center">
                                        <img src={charlesImg} alt="Charles" className="w-6 h-6 object-cover rounded" />
                                        <img src={charlesImg} alt="Charles" className="w-6 h-6 object-cover rounded" />
                                        <img src={charlesImg} alt="Charles" className="w-6 h-6 object-cover rounded" />
                                        <span className="text-gray-500 ml-2">Charles</span>
                                    </span>
                                    <span className="text-yellow-500 font-bold font-mono">25×</span>
                                </div>
                                <div className="flex justify-between items-center px-3 py-2 rounded-lg bg-white/[0.02]">
                                    <span className="flex gap-1.5 items-center">
                                        <span className="text-base">⭐ ⭐ ⭐</span>
                                        <span className="text-gray-500 ml-2">Stars</span>
                                    </span>
                                    <span className="text-yellow-500/70 font-bold font-mono">15×</span>
                                </div>
                                <div className="flex justify-between items-center px-3 py-2 rounded-lg bg-white/[0.02]">
                                    <span className="flex gap-1.5 items-center">
                                        <span className="text-base">🍒 🍒 🍒</span>
                                        <span className="text-gray-500 ml-2">Any Fruit</span>
                                    </span>
                                    <span className="text-gray-500 font-bold font-mono">10×</span>
                                </div>
                                <div className="flex justify-between items-center px-3 py-2">
                                    <span className="text-gray-600">Any 2 Matching</span>
                                    <span className="text-gray-600 font-bold font-mono">2×</span>
                                </div>
                            </div>
                        </div>
                    </div>
                </div>

            </main>
        </div>
    );
}

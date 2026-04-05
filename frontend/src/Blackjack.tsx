import { useState } from 'react';
import { Link } from 'react-router-dom';
import { ArrowLeft } from 'lucide-react';
import type { User } from 'firebase/auth';

interface BlackjackProps {
    user: User | null;
    profile: any;
    refreshProfile: () => void;
}

type GameStatus = 'idle' | 'playing' | 'done';

interface GameState {
    playerCards: string[];
    dealerVisibleCard: string;   // always set during play
    dealerAllCards: string[];    // only set when done
    playerTotal: number;
    dealerVisibleTotal: number;
    dealerFinalTotal: number;
    result: string;
    payout: number;
}

// ── card helpers ──────────────────────────────────────────────────────────────

function parseCard(card: string) {
    const suits = ['♠', '♥', '♦', '♣'];
    let suit = '';
    for (const s of suits) {
        if (card.endsWith(s)) { suit = s; break; }
    }
    const rank = suit ? card.slice(0, -1) : card;
    const isRed = suit === '♥' || suit === '♦';
    return { rank, suit, isRed };
}

function PlayingCard({ card, faceDown = false }: { card: string; faceDown?: boolean; small?: boolean }) {
    if (faceDown) {
        return (
            <div className="w-14 h-20 rounded-xl border-2 border-gray-600 bg-gradient-to-br from-indigo-900 to-purple-900 flex items-center justify-center shadow-lg flex-shrink-0">
                <div className="w-8 h-12 rounded-lg border border-indigo-700/50 bg-indigo-800/30" />
            </div>
        );
    }
    const { rank, suit, isRed } = parseCard(card);
    return (
        <div className={`w-14 h-20 rounded-xl border-2 ${isRed ? 'border-red-300' : 'border-gray-400'} bg-white shadow-xl flex flex-col justify-between p-1.5 flex-shrink-0`}>
            <div className={`text-xs font-black leading-none ${isRed ? 'text-red-600' : 'text-gray-900'}`}>{rank}</div>
            <div className={`text-xl text-center leading-none ${isRed ? 'text-red-500' : 'text-gray-800'}`}>{suit}</div>
            <div className={`text-xs font-black rotate-180 leading-none ${isRed ? 'text-red-600' : 'text-gray-900'}`}>{rank}</div>
        </div>
    );
}

function HandRow({ label, cards, isHidden = false, total, showBust = false }: {
    label: string;
    cards: string[];
    isHidden?: boolean;
    total: number;
    showBust?: boolean;
}) {
    return (
        <div className="space-y-2">
            <div className="flex items-center gap-2">
                <span className="text-xs font-bold text-gray-400 uppercase tracking-widest">{label}</span>
                <span className={`px-2 py-0.5 rounded-md font-mono font-bold text-sm border ${showBust && total > 21 ? 'bg-red-900/40 border-red-500/40 text-red-300' : 'bg-black/30 border-gray-700 text-white'}`}>
                    {showBust && total > 21 ? `BUST (${total})` : total}
                </span>
            </div>
            <div className="flex gap-2 flex-wrap">
                {cards.map((card, i) => (
                    <PlayingCard
                        key={i}
                        card={card}
                        faceDown={isHidden && i === 1}
                    />
                ))}
            </div>
        </div>
    );
}

// ── component ──────────────────────────────────────────────────────────────────

export default function Blackjack({ user, profile, refreshProfile }: BlackjackProps) {
    const [bet, setBet] = useState('10');
    const [status, setStatus] = useState<GameStatus>('idle');
    const [game, setGame] = useState<Partial<GameState>>({});
    const [loading, setLoading] = useState(false);
    const [error, setError] = useState<string | null>(null);

    const API = import.meta.env.VITE_API_URL;
    const uid = user?.uid;

    const deal = async () => {
        const betAmt = parseFloat(bet);
        if (!uid || isNaN(betAmt) || betAmt <= 0) return;
        setLoading(true);
        setError(null);
        try {
            const res = await fetch(`${API}/api/v1/casino/blackjack/deal`, {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ userId: uid, betAmount: betAmt }),
            });
            const data = await res.json();
            if (!res.ok) throw new Error(data.error ?? 'Deal failed');

            if (data.status === 'done') {
                // natural blackjack
                setGame({
                    playerCards: data.playerCards,
                    dealerAllCards: data.dealerCards,
                    playerTotal: data.playerTotal,
                    dealerFinalTotal: data.dealerTotal,
                    result: data.result,
                    payout: data.payout,
                });
                setStatus('done');
            } else {
                setGame({
                    playerCards: data.playerCards,
                    dealerVisibleCard: data.dealerVisibleCard,
                    playerTotal: data.playerTotal,
                    dealerVisibleTotal: data.dealerVisibleTotal ?? 0,
                });
                setStatus('playing');
            }
            refreshProfile();
        } catch (e: any) {
            setError(e.message);
        } finally {
            setLoading(false);
        }
    };

    const doAction = async (act: 'hit' | 'stand' | 'double') => {
        if (!uid || status !== 'playing') return;
        setError(null);
        // set loading per-action so buttons show feedback
        setLoading(true);
        try {
            const res = await fetch(`${API}/api/v1/casino/blackjack/action`, {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ userId: uid, action: act }),
            });
            const data = await res.json();
            if (!res.ok) throw new Error(data.error ?? 'Action failed');

            if (data.status === 'done') {
                setGame({
                    playerCards: data.playerCards,
                    dealerAllCards: data.dealerCards,
                    playerTotal: data.playerTotal,
                    dealerFinalTotal: data.dealerTotal,
                    result: data.result,
                    payout: data.payout,
                });
                setStatus('done');
            } else {
                // still playing (only happens on hit)
                setGame(prev => ({
                    ...prev,
                    playerCards: data.playerCards,
                    playerTotal: data.playerTotal,
                }));
            }
            refreshProfile();
        } catch (e: any) {
            setError(e.message);
        } finally {
            setLoading(false);
        }
    };

    const reset = () => {
        setStatus('idle');
        setGame({});
        setError(null);
    };

    const CHIPS = [5, 10, 25, 50, 67, 100, 500];

    const resultLabel: Record<string, string> = {
        win: '🏆 You Win!',
        blackjack: '🃏 Blackjack! (2.5×)',
        push: '🤝 Push',
        bust: '💥 Bust',
        lose: '😔 Dealer Wins',
    };

    const resultColor = (r: string) => {
        if (r === 'win' || r === 'blackjack') return 'border-green-500/50 bg-green-900/30 text-green-300';
        if (r === 'push') return 'border-yellow-500/40 bg-yellow-900/20 text-yellow-300';
        return 'border-red-500/40 bg-red-900/20 text-red-300';
    };

    const balance = profile?.Balance ?? 0;
    const betNum = parseFloat(bet) || 0;
    const canDeal = !loading && !!uid && betNum > 0 && (balance === 0 || betNum <= balance);

    return (
        <div className="min-h-screen text-gray-100 flex flex-col pt-4"
            style={{ background: 'radial-gradient(ellipse at top, #0d2818 0%, #050f0a 60%, #000 100%)' }}>

            <header className="max-w-5xl mx-auto w-full px-4 mb-6 flex items-center gap-4">
                <Link to="/casino" className="p-2 rounded-xl bg-gray-900/80 border border-gray-800 hover:bg-gray-800 transition-colors">
                    <ArrowLeft className="w-5 h-5 text-gray-400" />
                </Link>
                <h1 className="text-3xl font-black text-white">♠ Blackjack</h1>
                <div className="ml-auto text-sm text-gray-400">
                    Balance: <span className="text-green-400 font-mono font-bold">{balance.toFixed(2)}</span>
                </div>
            </header>

            <main className="flex-1 max-w-5xl mx-auto px-4 w-full pb-12">
                {/* felt table */}
                <div
                    className="rounded-3xl border border-green-900/40 p-8 space-y-6 relative overflow-hidden"
                    style={{
                        background: 'radial-gradient(ellipse at center, #1a3a2a 0%, #0d2018 70%, #081510 100%)',
                        boxShadow: 'inset 0 2px 20px rgba(0,0,0,0.5), 0 0 40px rgba(0,80,40,0.15)',
                    }}
                >
                    {/* felt texture overlay */}
                    <div className="absolute inset-0 rounded-3xl pointer-events-none"
                        style={{ background: 'repeating-linear-gradient(45deg, transparent, transparent 40px, rgba(0,0,0,0.04) 40px, rgba(0,0,0,0.04) 80px)' }} />

                    {/* waiting state */}
                    {status === 'idle' && (
                        <div className="text-center py-6 text-green-900/50 text-lg font-black tracking-widest select-none">
                            ♠ &nbsp; VEXAS BLACKJACK &nbsp; ♠
                        </div>
                    )}

                    {/* dealer hand */}
                    {status !== 'idle' && (
                        <HandRow
                            label="Dealer"
                            cards={status === 'done'
                                ? (game.dealerAllCards ?? [])
                                : [game.dealerVisibleCard ?? '', '??']}
                            isHidden={status === 'playing'}
                            total={status === 'done' ? (game.dealerFinalTotal ?? 0) : (game.dealerVisibleTotal ?? 0)}
                            showBust={status === 'done'}
                        />
                    )}

                    {/* result banner */}
                    {status === 'done' && game.result && (
                        <div className={`rounded-2xl border px-6 py-4 text-center font-black text-xl relative z-10 ${resultColor(game.result)}`}>
                            {resultLabel[game.result] ?? game.result}
                            {(game.payout ?? 0) > 0 && (
                                <span className="block text-base font-semibold mt-1 opacity-80">
                                    +{(game.payout ?? 0).toFixed(2)} S.H.I.T. returned
                                </span>
                            )}
                        </div>
                    )}

                    {/* player hand */}
                    {status !== 'idle' && (
                        <HandRow
                            label="You"
                            cards={game.playerCards ?? []}
                            isHidden={false}
                            total={game.playerTotal ?? 0}
                            showBust
                        />
                    )}

                    {/* error */}
                    {error && (
                        <div className="px-4 py-3 rounded-xl bg-red-900/40 border border-red-500/40 text-red-300 text-sm text-center font-medium">
                            ⚠ {error}
                        </div>
                    )}

                    <div className="border-t border-green-900/30" />

                    {/* ── idle controls ── */}
                    {status === 'idle' && (
                        <div className="space-y-4 relative z-10">
                            {/* chip picker */}
                            <div className="flex flex-wrap gap-2 justify-center">
                                {CHIPS.map(c => (
                                    <button
                                        key={c}
                                        onClick={() => setBet(String(c))}
                                        className={`px-4 py-2 rounded-xl font-bold text-sm transition-all border ${bet === String(c)
                                            ? 'bg-yellow-500 border-yellow-400 text-black scale-105 shadow-lg shadow-yellow-500/30'
                                            : 'bg-black/20 border-green-900/40 text-yellow-400 hover:bg-black/40'}`}
                                    >
                                        {c}
                                    </button>
                                ))}
                                <input
                                    type="number"
                                    value={bet}
                                    onChange={e => setBet(e.target.value)}
                                    className="w-24 px-3 py-2 rounded-xl bg-black/30 border border-green-900/40 text-white font-mono text-sm text-center focus:outline-none focus:border-yellow-500/50"
                                    min="1"
                                />
                            </div>
                            <div className="text-center">
                                <button
                                    id="blackjack-deal-btn"
                                    onClick={deal}
                                    disabled={!canDeal}
                                    className="px-12 py-4 bg-yellow-500 hover:bg-yellow-400 disabled:bg-gray-700 disabled:text-gray-500 disabled:cursor-not-allowed text-black font-black text-lg rounded-2xl transition-all hover:scale-105 active:scale-95 shadow-xl shadow-yellow-500/20"
                                >
                                    {loading ? 'Dealing…' : `Deal — ${bet} S.H.I.T.`}
                                </button>
                                {!uid && <p className="text-red-400 text-xs mt-2">log in to play</p>}
                                {uid && betNum > 0 && betNum > balance && balance > 0 && (
                                    <p className="text-red-400 text-xs mt-2">insufficient balance</p>
                                )}
                            </div>
                        </div>
                    )}

                    {/* ── playing controls ── */}
                    {status === 'playing' && (
                        <div className="flex gap-3 justify-center flex-wrap relative z-10">
                            <button
                                id="blackjack-hit-btn"
                                onClick={() => doAction('hit')}
                                disabled={loading}
                                className="px-10 py-3 bg-green-600 hover:bg-green-500 disabled:bg-gray-700 disabled:text-gray-500 text-white font-black rounded-xl transition-all hover:scale-105 active:scale-95 text-lg"
                            >
                                {loading ? '…' : 'Hit'}
                            </button>
                            <button
                                id="blackjack-stand-btn"
                                onClick={() => doAction('stand')}
                                disabled={loading}
                                className="px-10 py-3 bg-red-700 hover:bg-red-600 disabled:bg-gray-700 disabled:text-gray-500 text-white font-black rounded-xl transition-all hover:scale-105 active:scale-95 text-lg"
                            >
                                {loading ? '…' : 'Stand'}
                            </button>
                            <button
                                id="blackjack-double-btn"
                                onClick={() => doAction('double')}
                                disabled={loading || betNum > balance}
                                className="px-10 py-3 bg-yellow-600 hover:bg-yellow-500 disabled:bg-gray-700 disabled:text-gray-500 text-white font-black rounded-xl transition-all hover:scale-105 active:scale-95 text-lg"
                                title={betNum > balance ? 'not enough balance to double' : 'Double Down'}
                            >
                                {loading ? '…' : 'Double'}
                            </button>
                        </div>
                    )}

                    {/* ── done controls ── */}
                    {status === 'done' && (
                        <div className="text-center relative z-10">
                            <button
                                id="blackjack-new-hand-btn"
                                onClick={reset}
                                className="px-10 py-3 bg-indigo-600 hover:bg-indigo-500 text-white font-black rounded-xl transition-all hover:scale-105 active:scale-95"
                            >
                                New Hand
                            </button>
                        </div>
                    )}
                </div>
            </main>
        </div>
    );
}

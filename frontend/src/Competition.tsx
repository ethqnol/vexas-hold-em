import { useState, useEffect, useRef } from 'react';
import { useParams, Link } from 'react-router-dom';
import { Loader2, ArrowLeft, TrendingUp, TrendingDown } from 'lucide-react';
import {
    LineChart, Line, XAxis, YAxis, Tooltip, ResponsiveContainer, ReferenceLine
} from 'recharts';
import { auth } from './firebase';

interface Market {
    id: string;
    data: {
        teamId: string;
        teamName: string;
        division: string;
        organization: string;
        location: string;
        yesPool: number;
        noPool: number;
        initialYesOdds: number;
    };
    yesOdds: number;
}

interface HistoryPoint {
    t: number;
    y: number;
}

function oddsToPrice(odds: number) {
    return Math.round(odds * 100);
}

const CHART_COLORS = [
    '#ef4444', '#f97316', '#f59e0b', '#84cc16', '#22c55e',
    '#10b981', '#14b8a6', '#06b6d4', '#0ea5e9', '#3b82f6',
    '#6366f1', '#8b5cf6', '#a855f7', '#d946ef', '#ec4899', '#f43f5e'
];

function Competition() {
    const { id } = useParams();
    const [markets, setMarkets] = useState<Market[]>([]);
    const [loading, setLoading] = useState(true);
    const [selected, setSelected] = useState<Market | null>(null);
    const [history, setHistory] = useState<HistoryPoint[]>([]);
    const [historyLoading, setHistoryLoading] = useState(false);

    const [globalHistory, setGlobalHistory] = useState<any[]>([]);
    const [globalLoading, setGlobalLoading] = useState(false);
    const [hoveredLine, setHoveredLine] = useState<string | null>(null);

    const [tradeType, setTradeType] = useState<'YES' | 'NO'>('YES');
    const [amount, setAmount] = useState('');
    const [tradeMode, setTradeMode] = useState<'buy' | 'sell'>('buy');
    const [submitting, setSubmitting] = useState(false);
    const [tradeResult, setTradeResult] = useState<{ shares?: number; payout?: number } | null>(null);
    const [tradeError, setTradeError] = useState<string | null>(null);
    const panelRef = useRef<HTMLDivElement>(null);

    useEffect(() => {
        const loadData = async () => {
            let loadedMarkets: Market[] = [];
            try {
                const res = await fetch(`${import.meta.env.VITE_API_URL}/api/v1/competitions/${id}/markets`);
                if (res.ok) {
                    const data = await res.json();
                    loadedMarkets = data.markets || [];
                    setMarkets(loadedMarkets);
                }
            } catch (e) {
                console.error(e);
            } finally {
                setLoading(false);
            }

            if (loadedMarkets.length > 0) {
                setGlobalLoading(true);
                try {
                    const res = await fetch(`${import.meta.env.VITE_API_URL}/api/v1/competitions/${id}/history`);
                    if (res.ok) {
                        const data = await res.json();
                        const points = data.history || [];
                        const sortedPoints = [...points].sort((a, b) => a.t - b.t);

                        // Seed every market with its initial odds (stored at creation time).
                        // This is correct for markets with no history (current = initial)
                        // and gives the correct pre-history baseline for markets with trades.
                        const teamOdds: Record<string, number> = {};
                        loadedMarkets.forEach(m => {
                            teamOdds[m.id] = m.data.initialYesOdds ?? m.yesOdds;
                        });

                        const chart: any[] = [];

                        const now = Date.now();
                        const minTime = points.length > 0
                            ? points[0].t - Math.max(60000, (now - points[0].t) * 0.10)
                            : now - 60000;
                        chart.push({ t: minTime, ...teamOdds });

                        for (const pt of sortedPoints) {
                            teamOdds[pt.teamId] = pt.y;
                            chart.push({ t: pt.t, ...teamOdds });
                        }

                        chart.push({ t: Date.now(), ...teamOdds });

                        setGlobalHistory(chart);
                    }
                } catch (e) {
                    console.error(e);
                } finally {
                    setGlobalLoading(false);
                }
            }
        };

        if (id) {
            loadData();
        }
    }, [id]);

    const fetchHistory = async (marketId: string) => {
        setHistoryLoading(true);
        try {
            const res = await fetch(`${import.meta.env.VITE_API_URL}/api/v1/competitions/${id}/markets/${marketId}/history`);
            if (res.ok) {
                const data = await res.json();
                setHistory(data.history || []);
            }
        } catch (e) {
            console.error(e);
        } finally {
            setHistoryLoading(false);
        }
    };

    const handleSelectMarket = (market: Market, type: 'YES' | 'NO') => {
        setSelected(market);
        setTradeType(type);
        setTradeMode('buy');
        setAmount('');
        setTradeResult(null);
        setTradeError(null);
        fetchHistory(market.id);
    };

    const handleTrade = async () => {
        if (!selected || !amount || !auth.currentUser) return;
        setSubmitting(true);
        setTradeResult(null);
        setTradeError(null);
        const uid = auth.currentUser.uid;

        try {
            if (tradeMode === 'buy') {
                const res = await fetch(`${import.meta.env.VITE_API_URL}/api/v1/trade`, {
                    method: 'POST',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify({
                        competitionId: id,
                        marketId: selected.id,
                        userId: uid,
                        tradeType: tradeType,
                        amount: parseFloat(amount),
                    }),
                });
                const data = await res.json();
                if (!res.ok) throw new Error(data.error || 'Trade failed');
                setTradeResult({ shares: data.shares });
                const newOdds = data.newYesPool / (data.newYesPool + data.newNoPool);
                setMarkets(prev => prev.map(m => m.id === selected.id ? { ...m, yesOdds: newOdds } : m));
                setSelected(prev => prev ? { ...prev, yesOdds: newOdds } : prev);
                setHistory(prev => [...prev, { t: Date.now(), y: newOdds }]);
                setGlobalHistory(prev => {
                    const last = prev.length ? prev[prev.length - 1] : {};
                    return [...prev, { t: Date.now(), ...last, [selected.id]: newOdds }];
                });
            } else {
                const res = await fetch(`${import.meta.env.VITE_API_URL}/api/v1/trade/sell`, {
                    method: 'POST',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify({
                        competitionId: id,
                        marketId: selected.id,
                        userId: uid,
                        tradeType: tradeType,
                        shares: parseFloat(amount),
                    }),
                });
                const data = await res.json();
                if (!res.ok) throw new Error(data.error || 'Sell failed');
                setTradeResult({ payout: data.payout });
                const newOdds = data.newYesPool / (data.newYesPool + data.newNoPool);
                setMarkets(prev => prev.map(m => m.id === selected.id ? { ...m, yesOdds: newOdds } : m));
                setSelected(prev => prev ? { ...prev, yesOdds: newOdds } : prev);
                setHistory(prev => [...prev, { t: Date.now(), y: newOdds }]);
                setGlobalHistory(prev => {
                    const last = prev.length ? prev[prev.length - 1] : {};
                    return [...prev, { t: Date.now(), ...last, [selected.id]: newOdds }];
                });
            }
        } catch (e: any) {
            setTradeError(e.message);
        } finally {
            setSubmitting(false);
        }
    };

    if (loading) {
        return (
            <div className="flex items-center justify-center h-screen bg-[#111]">
                <Loader2 className="w-8 h-8 text-blue-400 animate-spin" />
            </div>
        );
    }

    const yesPrice = selected ? oddsToPrice(selected.yesOdds) : 0;
    const noPrice = selected ? 100 - yesPrice : 0;

    return (
        <div className="min-h-screen bg-[#111] text-gray-100 flex flex-col">

            {/* ── header ── */}
            <header className="bg-[#111]/90 backdrop-blur sticky top-0 z-50 px-16 h-14 flex items-center gap-4 border-b border-white/[0.04]">
                <Link to="/" className="p-1.5 rounded-lg hover:bg-white/5 transition-colors">
                    <ArrowLeft className="w-4 h-4 text-gray-400" />
                </Link>
                <h1 className="font-bold text-white text-lg">{id}</h1>
                <span className="ml-2 text-[10px] font-bold uppercase tracking-widest px-2 py-0.5 rounded-full bg-green-500/15 text-green-400 border border-green-500/20">
                    LIVE
                </span>
                {selected && (
                    <span className="ml-4 text-sm text-gray-400">
                        {selected.data.teamName || `Team ${selected.data.teamId}`}
                        <span className="ml-2 text-white font-semibold">{yesPrice}% YES</span>
                    </span>
                )}
            </header>

            {/* ── global chart ── */}
            <div className="mx-auto w-full max-w-[1400px] px-16 pt-6">
                <div className="flex items-center justify-between mb-3">
                    <span className="text-xs uppercase tracking-widest text-gray-500">
                        Market Trends
                    </span>
                </div>
                {globalLoading ? (
                    <div className="flex items-center justify-center h-48 border border-white/[0.04] rounded-xl bg-white/[0.01]">
                        <Loader2 className="w-5 h-5 animate-spin text-gray-600" />
                    </div>
                ) : globalHistory.length < 2 ? (
                    <div className="h-48 flex items-center justify-center text-sm text-gray-600 border border-white/[0.04] rounded-xl bg-white/[0.01]">
                        No market activity yet.
                    </div>
                ) : (
                    <div className="h-72 pt-4 pb-2 pr-[20%] border border-white/[0.04] rounded-xl bg-[#161616]">
                        <ResponsiveContainer width="100%" height="100%">
                            {(() => {
                                // Compute Y domain from actual data so the chart fills the viewport
                                const allVals = globalHistory.flatMap(pt =>
                                    markets.map(m => pt[m.id] as number).filter(v => v != null)
                                );
                                const dataMin = allVals.length ? Math.min(...allVals) : 0;
                                const dataMax = allVals.length ? Math.max(...allVals) : 1;
                                const pad = Math.max((dataMax - dataMin) * 0.15, 0.02);
                                const yMin = Math.max(0, dataMin - pad);
                                const yMax = Math.min(1, dataMax + pad);
                                const step = (yMax - yMin) <= 0.15 ? 0.05 : (yMax - yMin) <= 0.4 ? 0.1 : 0.25;
                                const ticks: number[] = [];
                                const start = Math.ceil(yMin / step) * step;
                                for (let t = start; t <= yMax + 1e-9; t = Math.round((t + step) * 1e6) / 1e6) ticks.push(t);

                                return (
                            <LineChart data={globalHistory} margin={{ top: 12, right: 0, left: 0, bottom: 0 }}>
                                <XAxis
                                    dataKey="t"
                                    type="number"
                                    domain={['dataMin', 'dataMax']}
                                    hide
                                />
                                <YAxis
                                    domain={[yMin, yMax]}
                                    ticks={ticks}
                                    tickFormatter={v => `${Math.round(v * 100)}%`}
                                    tick={{ fontSize: 10, fill: '#4b5563' }}
                                    width={40}
                                    axisLine={false}
                                    tickLine={false}
                                />
                                <Tooltip
                                    cursor={{ stroke: 'rgba(255,255,255,0.15)', strokeWidth: 1, strokeDasharray: '4 4' }}
                                    content={({ active, payload }) => {
                                        if (!active || !payload || !payload.length) return null;

                                        // only show tooltip if we're actively hovering a line
                                        if (!hoveredLine) return null;

                                        const p = payload.find(x => x.dataKey === hoveredLine);
                                        if (!p) return null;

                                        const market = markets.find(m => m.id === p.dataKey);
                                        const label = market ? market.data.teamName || `#${market.data.teamId}` : p.dataKey;

                                        return (
                                            <div className="bg-[#1a1a1a]/95 backdrop-blur-sm border border-white/[0.06] rounded-lg px-3 py-2 text-xs shadow-xl flex items-center gap-2">
                                                <div className="w-2 h-2 rounded-full shadow-sm" style={{ backgroundColor: p.color }}></div>
                                                <span className="text-gray-300 font-medium">{label as string}</span>
                                                <span className="text-white font-bold ml-1">{Math.round((p.value as number) * 100)}% YES</span>
                                            </div>
                                        );
                                    }}
                                />
                                {markets.map((m, i) => (
                                    <Line
                                        key={m.id}
                                        type="stepAfter"
                                        dataKey={m.id}
                                        stroke={CHART_COLORS[i % CHART_COLORS.length]}
                                        strokeWidth={hoveredLine === m.id ? 2.5 : 1.5}
                                        strokeOpacity={hoveredLine && hoveredLine !== m.id ? 0.15 : 0.9}
                                        dot={false}
                                        activeDot={{ r: 4 }}
                                        isAnimationActive={false}
                                        onMouseEnter={() => setHoveredLine(m.id)}
                                        onMouseLeave={() => setHoveredLine(null)}
                                    />
                                ))}
                            </LineChart>
                                );
                            })()}
                        </ResponsiveContainer>
                    </div>
                )}
            </div>

            <div className="flex flex-1 mx-auto w-full max-w-[1400px] justify-between px-16 pb-12 pt-6 gap-6 overflow-hidden">

                {/* market list */}
                <div className="flex-1 overflow-y-auto">
                    {/* column headers */}
                    <div className="grid grid-cols-[1fr_auto_auto] items-center py-3 mb-1 px-4 text-[11px] uppercase tracking-widest text-gray-600">
                        <span>Team</span>
                        <span className="w-32 text-center">Chance</span>
                        <span className="w-56 text-right pr-1">Trade</span>
                    </div>

                    {markets.length === 0 ? (
                        <div className="py-20 text-center text-gray-600">No markets yet.</div>
                    ) : (
                        markets
                            .slice()
                            .sort((a, b) => b.yesOdds - a.yesOdds)
                            .map((m) => {
                                const yes = oddsToPrice(m.yesOdds);
                                const no = 100 - yes;
                                const isSelected = selected?.id === m.id;

                                return (
                                    <div
                                        key={m.id}
                                        className={`grid grid-cols-[1fr_auto_auto] items-center py-5 rounded-xl mb-0.5 px-4 transition-colors cursor-pointer ${isSelected ? 'bg-white/[0.04]' : 'hover:bg-white/[0.02]'}`}
                                        onClick={() => handleSelectMarket(m, 'YES')}
                                    >
                                        {/* team info */}
                                        <div>
                                            <div className="font-semibold text-white text-[15px]">
                                                {m.data.teamName || `Team ${m.data.teamId}`}
                                            </div>
                                            <div className="text-xs text-gray-600 mt-1">
                                                #{m.data.teamId}
                                                {m.data.division && m.data.division !== 'Default' && (
                                                    <span className="ml-2">· {m.data.division}</span>
                                                )}
                                            </div>
                                        </div>

                                        {/* odds */}
                                        <div className="w-32 text-center">
                                            <span className="text-xl font-bold text-white">{yes}%</span>
                                            <span className="text-xs text-gray-600 ml-1">YES</span>
                                        </div>

                                        {/* buy buttons */}
                                        <div className="w-56 flex gap-2 justify-end" onClick={e => e.stopPropagation()}>
                                            <button
                                                onClick={() => handleSelectMarket(m, 'YES')}
                                                className="px-4 py-2.5 rounded-lg bg-[#1a6b45] hover:bg-[#1f7d52] text-white text-xs font-bold transition-colors min-w-[90px] text-center"
                                            >
                                                Buy Yes {yes}¢
                                            </button>
                                            <button
                                                onClick={() => handleSelectMarket(m, 'NO')}
                                                className="px-4 py-2.5 rounded-lg bg-[#7a1f1f] hover:bg-[#8f2424] text-white text-xs font-bold transition-colors min-w-[90px] text-center"
                                            >
                                                Buy No {no}¢
                                            </button>
                                        </div>
                                    </div>
                                );
                            })
                    )}
                </div>

                {/* ── trade panel ── */}
                {selected && (
                    <div
                        ref={panelRef}
                        className="bg-[#161616] rounded-xl flex flex-col overflow-y-auto shrink-0"
                        style={{ width: '22rem' }}
                    >
                        {/* panel header */}
                        <div className="px-6 py-6">
                            <div className="font-bold text-white text-lg leading-tight">
                                {selected.data.teamName || `Team ${selected.data.teamId}`}
                            </div>
                            <div className="text-xs text-gray-500 mt-0.5">#{selected.data.teamId}</div>
                        </div>

                        {/* chart */}
                        <div className="px-6 pt-5 pb-2 border-b border-white/[0.04]">
                            <div className="text-[11px] uppercase tracking-widest text-gray-500 mb-2">Odds History</div>
                            {historyLoading ? (
                                <div className="flex items-center justify-center h-24">
                                    <Loader2 className="w-4 h-4 animate-spin text-gray-600" />
                                </div>
                            ) : history.length < 1 ? (
                                <div className="h-24 flex items-center justify-center text-xs text-gray-600 text-center px-4">
                                    no trades yet — chart appears after first trade
                                </div>
                            ) : (
                                <ResponsiveContainer width="100%" height={100}>
                                    {(() => {
                                        const sortedHistory = [...history].sort((a, b) => a.t - b.t);
                                        const endT = Date.now();
                                        const startT = sortedHistory.length > 0
                                            ? sortedHistory[0].t - Math.max(60000, (endT - sortedHistory[0].t) * 0.10)
                                            : endT - 60000;

                                        const singleChartData = sortedHistory.length === 0
                                            ? []
                                            : [
                                                ...sortedHistory,
                                                { ...sortedHistory[sortedHistory.length - 1], t: endT }
                                            ];

                                        const yVals = singleChartData.map(p => p.y);
                                        const dMin = yVals.length ? Math.min(...yVals) : 0;
                                        const dMax = yVals.length ? Math.max(...yVals) : 1;
                                        const pad = Math.max((dMax - dMin) * 0.2, 0.03);
                                        const yMin = Math.max(0, dMin - pad);
                                        const yMax = Math.min(1, dMax + pad);
                                        const step = (yMax - yMin) <= 0.15 ? 0.05 : (yMax - yMin) <= 0.4 ? 0.1 : 0.25;
                                        const ticks: number[] = [];
                                        const tickStart = Math.ceil(yMin / step) * step;
                                        for (let t = tickStart; t <= yMax + 1e-9; t = Math.round((t + step) * 1e6) / 1e6) ticks.push(t);

                                        return (
                                            <LineChart data={singleChartData} margin={{ top: 8, right: 8, left: 0, bottom: 0 }}>
                                                <XAxis
                                                    dataKey="t"
                                                    type="number"
                                                    domain={['dataMin', 'dataMax']}
                                                    hide
                                                />
                                                <YAxis
                                                    domain={[yMin, yMax]}
                                                    ticks={ticks}
                                                    tickFormatter={v => `${Math.round(v * 100)}%`}
                                                    tick={{ fontSize: 10, fill: '#4b5563' }}
                                                    width={36}
                                                    axisLine={false}
                                                    tickLine={false}
                                                />
                                                <Tooltip
                                                    cursor={{ stroke: 'rgba(255,255,255,0.15)', strokeWidth: 1, strokeDasharray: '4 4' }}
                                                    formatter={(v) => [`${Math.round((v as number) * 100)}%`, 'YES']}
                                                    labelFormatter={() => ''}
                                                    contentStyle={{ background: '#1a1a1a', border: '1px solid rgba(255,255,255,0.06)', borderRadius: 8, fontSize: 11 }}
                                                />
                                                <Line
                                                    type="stepAfter"
                                                    dataKey="y"
                                                    stroke="#3b82f6"
                                                    strokeWidth={2}
                                                    dot={false}
                                                    activeDot={{ r: 3, fill: '#3b82f6' }}
                                                    isAnimationActive={false}
                                                />
                                            </LineChart>
                                        );
                                    })()}
                                </ResponsiveContainer>
                            )}
                        </div>

                        <div className="px-6 py-6 flex flex-col gap-5">
                            {/* buy / sell tabs */}
                            <div className="flex rounded-lg overflow-hidden border border-white/[0.07] text-sm font-semibold">
                                <button
                                    onClick={() => setTradeMode('buy')}
                                    className={`flex-1 py-2 transition-colors ${tradeMode === 'buy' ? 'bg-white/10 text-white' : 'text-gray-500 hover:text-white'}`}
                                >
                                    Buy
                                </button>
                                <button
                                    onClick={() => setTradeMode('sell')}
                                    className={`flex-1 py-2 transition-colors ${tradeMode === 'sell' ? 'bg-white/10 text-white' : 'text-gray-500 hover:text-white'}`}
                                >
                                    Sell
                                </button>
                            </div>

                            {/* yes/no toggle */}
                            <div className="flex gap-2">
                                <button
                                    onClick={() => setTradeType('YES')}
                                    className={`flex-1 py-2.5 rounded-lg text-sm font-bold transition-all ${tradeType === 'YES' ? 'bg-[#1a6b45] text-white' : 'bg-white/5 text-gray-500 hover:bg-white/8'}`}
                                >
                                    Yes {yesPrice}¢
                                </button>
                                <button
                                    onClick={() => setTradeType('NO')}
                                    className={`flex-1 py-2.5 rounded-lg text-sm font-bold transition-all ${tradeType === 'NO' ? 'bg-[#7a1f1f] text-white' : 'bg-white/5 text-gray-500 hover:bg-white/8'}`}
                                >
                                    No {noPrice}¢
                                </button>
                            </div>

                            {/* amount input */}
                            <div>
                                <label className="text-xs text-gray-500 uppercase tracking-wider mb-2 block">
                                    {tradeMode === 'buy' ? 'Amount (S.H.I.T.)' : 'Shares to sell'}
                                </label>
                                <div className="relative">
                                    <span className="absolute left-3 top-1/2 -translate-y-1/2 text-gray-600 font-bold text-lg">
                                        {tradeMode === 'buy' ? '$' : '#'}
                                    </span>
                                    <input
                                        type="number"
                                        min="0"
                                        step="1"
                                        value={amount}
                                        onChange={e => setAmount(e.target.value)}
                                        placeholder="0"
                                        className="w-full bg-[#1e1e1e] border border-white/[0.07] rounded-lg pl-8 pr-4 py-3 text-white text-xl font-mono focus:outline-none focus:border-blue-500/40 transition-colors placeholder:text-gray-700"
                                    />
                                </div>

                                {tradeMode === 'buy' && (
                                    <div className="flex gap-2 mt-2">
                                        {[1, 5, 10, 100].map(v => (
                                            <button
                                                key={v}
                                                onClick={() => setAmount(a => String((parseFloat(a) || 0) + v))}
                                                className="flex-1 py-1.5 text-xs rounded-md bg-white/[0.04] hover:bg-white/[0.08] text-gray-500 transition-colors"
                                            >
                                                +${v}
                                            </button>
                                        ))}
                                    </div>
                                )}
                            </div>

                            {/* estimated return */}
                            {amount && parseFloat(amount) > 0 && (
                                <div className="p-3 rounded-lg bg-white/[0.04] text-xs text-gray-500 space-y-1">
                                    {(() => {
                                        const v = parseFloat(amount) || 0;
                                        if (v <= 0) return null;
                                        const yesPool = selected.data.yesPool;
                                        const noPool = selected.data.noPool;

                                        // mirrors backend buyShares:
                                        //   shares = amount + otherPool * ln((ownPool + amount) / ownPool)
                                        const estBuyShares = (amt: number, ownPool: number, otherPool: number) =>
                                            amt + otherPool * Math.log((ownPool + amt) / ownPool);

                                        // mirrors backend solveSellPayout (Newton's method):
                                        //   shares = X + otherPool * ln(ownPool / (ownPool - X))
                                        const estSellPayout = (shares: number, ownPool: number, otherPool: number) => {
                                            let X = shares * ownPool / (ownPool + otherPool);
                                            for (let i = 0; i < 20; i++) {
                                                const rem = ownPool - X;
                                                if (rem <= 1e-9) { X = ownPool * (1 - 1e-9); break; }
                                                const f = X + otherPool * Math.log(ownPool / rem) - shares;
                                                const df = 1 + otherPool / rem;
                                                const delta = f / df;
                                                X -= delta;
                                                if (Math.abs(delta) < 1e-10) break;
                                            }
                                            return Math.max(0, X);
                                        };

                                        let estShares = 0;
                                        let estPayout = 0;

                                        if (tradeMode === 'buy') {
                                            estShares = tradeType === 'YES'
                                                ? estBuyShares(v, yesPool, noPool)
                                                : estBuyShares(v, noPool, yesPool);
                                        } else {
                                            estPayout = tradeType === 'YES'
                                                ? estSellPayout(v, yesPool, noPool)
                                                : estSellPayout(v, noPool, yesPool);
                                        }

                                        const avgPrice = tradeMode === 'buy'
                                            ? (estShares > 0 ? v / estShares : 0)
                                            : (v > 0 ? estPayout / v : 0);

                                        return (<>
                                            <div className="flex justify-between">
                                                <span>avg execution price</span>
                                                <span className="text-gray-300 font-mono">
                                                    ~{Math.round(avgPrice * 100)}¢
                                                </span>
                                            </div>
                                            <div className="flex justify-between">
                                                <span>{tradeMode === 'buy' ? 'est. shares' : 'est. payout'}</span>
                                                <span className="text-gray-300 font-mono">
                                                    {tradeMode === 'buy'
                                                        ? `~${estShares.toFixed(4)}`
                                                        : `~$${estPayout.toFixed(2)}`}
                                                </span>
                                            </div>
                                        </>);
                                    })()}
                                </div>
                            )}

                            {/* submit */}
                            <button
                                onClick={handleTrade}
                                disabled={submitting || !amount || !auth.currentUser}
                                className="w-full py-3 rounded-lg font-bold text-sm bg-blue-600 hover:bg-blue-500 disabled:opacity-40 disabled:cursor-not-allowed text-white transition-colors"
                            >
                                {submitting ? <Loader2 className="w-4 h-4 animate-spin mx-auto" /> : 'Trade'}
                            </button>

                            {!auth.currentUser && (
                                <p className="text-center text-xs text-gray-600">sign in to trade</p>
                            )}

                            {tradeResult && (
                                <div className="p-3 rounded-lg bg-green-500/10 border border-green-500/15 text-green-400 text-sm flex items-center gap-2">
                                    <TrendingUp className="w-4 h-4 shrink-0" />
                                    {tradeResult.shares
                                        ? `got ${tradeResult.shares.toFixed(4)} shares`
                                        : `received ${tradeResult.payout?.toFixed(2)} S.H.I.T.`}
                                </div>
                            )}
                            {tradeError && (
                                <div className="p-3 rounded-lg bg-red-500/10 border border-red-500/15 text-red-400 text-sm flex items-center gap-2">
                                    <TrendingDown className="w-4 h-4 shrink-0" />
                                    {tradeError}
                                </div>
                            )}

                            {/* pool info */}
                            <div className="pt-2 space-y-1.5 text-xs text-gray-600">
                                <div className="flex justify-between">
                                    <span>YES pool</span>
                                    <span className="font-mono text-gray-500">{selected.data.yesPool?.toFixed(2)}</span>
                                </div>
                                <div className="flex justify-between">
                                    <span>NO pool</span>
                                    <span className="font-mono text-gray-500">{selected.data.noPool?.toFixed(2)}</span>
                                </div>
                            </div>
                        </div>
                    </div>
                )}
            </div>
        </div>
    );
}

export default Competition;

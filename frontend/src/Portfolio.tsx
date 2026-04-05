import { Link } from 'react-router-dom';
import { ArrowLeft, Wallet } from 'lucide-react';
import type { User } from 'firebase/auth';

interface PortfolioProps {
    user: User | null;
    profile: any;
    portfolio: any; // grouped by comp id
}

export default function Portfolio({ user, profile, portfolio }: PortfolioProps) {
    if (!user) {
        return (
            <div className="min-h-screen bg-[#111] text-gray-100 flex items-center justify-center flex-col">
                <h2 className="text-2xl font-bold mb-4">Please log in to view your portfolio</h2>
                <Link to="/" className="text-blue-500 hover:text-blue-400 transition-colors">Return Home</Link>
            </div>
        );
    }

    return (
        <div className="min-h-screen bg-[#111] text-gray-100 flex flex-col pt-4">
            <header className="max-w-7xl mx-auto w-full px-4 sm:px-6 lg:px-8 mb-8 flex items-center gap-4">
                <Link to="/" className="p-2 rounded-xl bg-gray-900 border border-gray-800 hover:bg-gray-800 transition-colors">
                    <ArrowLeft className="w-5 h-5 text-gray-400" />
                </Link>
                <div className="flex items-center gap-3">
                    <div className="p-2.5 bg-blue-500/10 rounded-xl border border-blue-500/20">
                        <Wallet className="w-6 h-6 text-blue-500" />
                    </div>
                    <h1 className="text-3xl font-bold">Your Portfolio</h1>
                </div>
            </header>

            <main className="flex-1 max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 w-full space-y-8 pb-12">
                {/* bal display */}
                {profile && (
                    <div className="p-8 bg-gray-900 rounded-3xl border border-gray-800 shadow-xl max-w-sm">
                        <div className="text-sm font-semibold text-gray-400 mb-1">Available Liquidity</div>
                        <div className="text-4xl font-mono text-green-400 font-bold tracking-tight">
                            {profile.Balance?.toLocaleString(undefined, { minimumFractionDigits: 2, maximumFractionDigits: 2 }) || "0.00"}
                            <span className="text-lg text-green-500/30 ml-2 uppercase tracking-widest italic">S.H.I.T.</span>
                        </div>
                    </div>
                )}

                {/* grouped positions */}
                <div className="space-y-8">
                    {portfolio && Object.keys(portfolio).length > 0 ? (
                        Object.entries(portfolio).map(([compId, markets]: [string, any]) => {
                            let hasAny = false;
                            Object.values(markets).forEach((pos: any) => {
                                if (pos.YesShares > 0 || pos.NoShares > 0) hasAny = true;
                            });
                            if (!hasAny) return null;

                            return (
                                <div key={compId} className="bg-gray-900 border border-gray-800 rounded-3xl overflow-hidden shadow-xl">
                                    <div className="px-8 py-5 border-b border-gray-800 bg-gray-900/50 flex items-center justify-between">
                                        <h2 className="text-xl font-bold text-white uppercase tracking-wider">Event: <span className="text-blue-400">{compId === "unknown_competition" ? "Legacy Trades" : compId}</span></h2>
                                        {compId !== "unknown_competition" && (
                                            <Link to={`/competition/${compId}`} className="text-sm font-medium text-blue-500 hover:text-blue-400 transition-colors">
                                                Go to Market / Trade →
                                            </Link>
                                        )}
                                    </div>
                                    <div className="p-8 grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
                                        {Object.entries(markets).map(([teamId, pos]: [string, any]) => {
                                            const hasYes = pos.YesShares > 0;
                                            const hasNo = pos.NoShares > 0;
                                            if (!hasYes && !hasNo) return null;

                                            return (
                                                <div key={teamId} className="bg-gray-950 rounded-2xl p-6 border border-gray-800 hover:border-gray-700 transition-colors shadow-sm">
                                                    <div className="mb-4">
                                                        <h3 className="font-bold text-lg text-white mb-1">Team {pos.TeamName || teamId}</h3>
                                                        {pos.TeamName && <div className="text-xs text-gray-500 font-mono">#{teamId}</div>}
                                                    </div>
                                                    <div className="flex gap-4">
                                                        {hasYes && (
                                                            <div className="flex-1 p-3 bg-green-500/10 rounded-xl border border-green-500/20 text-center">
                                                                <div className="text-[10px] font-bold text-green-500/70 uppercase tracking-widest mb-1">YES Shares</div>
                                                                <div className="font-mono text-2xl font-bold text-green-400">{pos.YesShares?.toFixed(1)}</div>
                                                            </div>
                                                        )}
                                                        {hasNo && (
                                                            <div className="flex-1 p-3 bg-red-500/10 rounded-xl border border-red-500/20 text-center">
                                                                <div className="text-[10px] font-bold text-red-500/70 uppercase tracking-widest mb-1">NO Shares</div>
                                                                <div className="font-mono text-2xl font-bold text-red-400">{pos.NoShares?.toFixed(1)}</div>
                                                            </div>
                                                        )}
                                                    </div>
                                                </div>
                                            );
                                        })}
                                    </div>
                                </div>
                            );
                        })
                    ) : (
                        <div className="text-center p-12 border border-dashed border-gray-800 rounded-3xl bg-gray-900/30">
                            <h3 className="text-xl font-bold text-gray-400 mb-2">No active positions yet</h3>
                            <p className="text-gray-500 max-w-md mx-auto">Shares you purchase on teams during active competitions will be securely stored and tracked here.</p>
                            <Link to="/" className="inline-block mt-6 px-6 py-3 bg-blue-600 hover:bg-blue-500 text-white font-medium rounded-xl transition-colors shadow-lg shadow-blue-500/20">
                                Browse Events
                            </Link>
                        </div>
                    )}
                </div>
            </main>
        </div>
    );
}

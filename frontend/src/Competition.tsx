import { useState, useEffect } from 'react';
import { useParams, Link } from 'react-router-dom';
import { Loader2, ArrowLeft } from 'lucide-react';

function Competition() {
    const { id } = useParams();
    const [markets, setMarkets] = useState<any[]>([]);
    const [loading, setLoading] = useState(true);

    useEffect(() => {
        const fetchMarkets = async () => {
            try {
                const res = await fetch(`http://localhost:8080/api/v1/competitions/${id}/markets`);
                if (res.ok) {
                    const data = await res.json();
                    setMarkets(data.markets || []);
                }
            } catch (e) {
                console.error(e);
            } finally {
                setLoading(false);
            }
        };

        if (id) {
            fetchMarkets();
        }
    }, [id]);

    if (loading) {
        return (
            <div className="flex flex-col items-center justify-center h-[60vh] text-center">
                <Loader2 className="w-8 h-8 text-indigo-500 animate-spin mx-auto mb-4" />
                <p className="text-gray-400">Loading competition markets...</p>
            </div>
        );
    }

    return (
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-12 w-full">
            <div className="mb-8 flex items-center gap-4">
                <Link to="/" className="p-2 bg-gray-900 hover:bg-gray-800 rounded-xl transition-colors border border-gray-800">
                    <ArrowLeft className="w-5 h-5 text-gray-400" />
                </Link>
                <div>
                    <h1 className="text-3xl font-bold text-white flex items-center gap-3">
                        {id}
                        <span className="text-sm font-mono bg-green-500/10 text-green-400 px-3 py-1 rounded-full border border-green-500/20">LIVE</span>
                    </h1>
                    <p className="text-gray-400 mt-1">Trade YES or NO shares on your favorite teams.</p>
                </div>
            </div>

            <div className="p-8 rounded-2xl bg-gray-900 border border-gray-800">
                <h2 className="text-xl font-bold mb-6">Active Markets</h2>

                {markets.length === 0 ? (
                    <div className="p-8 border border-dashed border-gray-800 rounded-xl text-center text-gray-500">
                        No markets available for this competition yet.
                    </div>
                ) : (
                    <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
                        {markets.map((marketWrapper, idx) => {
                            const market = marketWrapper.data;
                            return (
                                <div key={marketWrapper.id || idx} className="p-5 rounded-2xl bg-gray-950 border border-gray-800 hover:border-indigo-500/50 transition-all shadow-lg hover:shadow-indigo-500/10 group flex flex-col h-full">
                                    <div className="flex justify-between items-start mb-4">
                                        <div>
                                            <h3 className="font-extrabold text-2xl text-white group-hover:text-indigo-400 transition-colors tracking-tight">Team {market.teamId}</h3>
                                            <div className="text-sm font-medium text-gray-400 mt-1 line-clamp-2" title={market.teamName}>{market.teamName}</div>
                                        </div>
                                        {market.division && market.division !== "Default" && (
                                            <div className="text-[10px] text-gray-400 uppercase tracking-widest px-2 py-1 bg-gray-800/80 rounded border border-gray-700 w-max ml-auto shadow-inner">{market.division}</div>
                                        )}
                                    </div>

                                    <div className="text-xs text-gray-500 mb-6 flex-1">
                                        <div className="flex items-center gap-2 mb-1">
                                            <span className="w-4 h-4 rounded bg-gray-800 flex items-center justify-center text-[8px]">🏫</span>
                                            <span className="truncate">{market.organization || 'Unknown Org'}</span>
                                        </div>
                                        <div className="flex items-center gap-2">
                                            <span className="w-4 h-4 rounded bg-gray-800 flex items-center justify-center text-[8px]">📍</span>
                                            <span className="truncate">{market.location || 'Unknown Location'}</span>
                                        </div>
                                    </div>

                                    {/* AMM Pricing Display Skeleton */}
                                    <div className="flex gap-3 w-full mt-auto">
                                        <button className="flex-1 bg-green-500/10 hover:bg-green-500/20 border border-green-500/30 text-green-400 py-3 rounded-xl text-sm font-bold transition-all shadow-sm hover:scale-[1.02] active:scale-[0.98]">
                                            BUY YES
                                        </button>
                                        <button className="flex-1 bg-red-500/10 hover:bg-red-500/20 border border-red-500/30 text-red-400 py-3 rounded-xl text-sm font-bold transition-all shadow-sm hover:scale-[1.02] active:scale-[0.98]">
                                            BUY NO
                                        </button>
                                    </div>
                                </div>
                            );
                        })}
                    </div>
                )}
            </div>
        </div>
    );
}

export default Competition;

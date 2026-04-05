import { useState, useEffect } from 'react';
import { Link } from 'react-router-dom';
import { ArrowLeft, ShoppingBag, Check, Palette, Tag } from 'lucide-react';
import type { User } from 'firebase/auth';

interface StoreProps {
    user: User | null;
    profile: any;
    refreshProfile: () => void;
}

interface StoreItem {
    id: string;
    name: string;
    cost: number;
    type: 'cosmetic' | 'title';
    value: string;
}

const THEME_PREVIEWS: Record<string, { gradient: string; accent: string; desc: string }> = {
    theme_default: {
        gradient: 'from-gray-900 to-gray-950',
        accent: 'border-gray-700 text-gray-500',
        desc: 'og theme         ',
    },
    theme_industrial: {
        gradient: 'from-cyan-900 to-blue-900',
        accent: 'border-blue-400/50 text-blue-200',
        desc: 'windows xp ig idk what to say, 67',
    },
    theme_gold: {
        gradient: 'from-amber-950 to-yellow-900',
        accent: 'border-yellow-400/50 text-yellow-200',
        desc: 'for when you want everyone to know you made money once.',
    },
    theme_neon: {
        gradient: 'from-fuchsia-950 to-violet-950',
        accent: 'border-fuchsia-400/50 text-fuchsia-200',
        desc: 'do not enable this if you have epilepsy.',
    },
    theme_vomit: {
        gradient: 'from-lime-950 to-green-950',
        accent: 'border-lime-500/50 text-lime-300',
        desc: 'sick.',
    },
};


export default function Store({ user, profile, refreshProfile }: StoreProps) {
    const [items, setItems] = useState<StoreItem[]>([]);
    const [loading, setLoading] = useState<string | null>(null);
    const [toast, setToast] = useState<string | null>(null);

    const API = import.meta.env.VITE_API_URL;

    useEffect(() => {
        fetch(`${API}/api/v1/casino/store`)
            .then(r => r.json())
            .then(d => setItems(d.items ?? []));
    }, []);

    const owned = (itemId: string) => {
        if (itemId === 'theme_default' || itemId === 'title_none') return true;
        return (profile?.OwnedItems ?? []).includes(itemId);
    };
    const isEquipped = (item: StoreItem) => {
        if (item.type === 'cosmetic') return (profile?.Cosmetic || "") === item.value;
        return (profile?.Title || "") === item.value;
    };

    const showToast = (msg: string) => {
        setToast(msg);
        setTimeout(() => setToast(null), 3000);
    };

    const buy = async (item: StoreItem) => {
        if (!user || loading) return;
        setLoading(item.id);
        try {
            const res = await fetch(`${API}/api/v1/casino/store/buy`, {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ userId: user.uid, itemId: item.id }),
            });
            const data = await res.json();
            if (!res.ok) throw new Error(data.error);
            showToast(owned(item.id) ? `${item.name} equipped!` : `${item.name} purchased & equipped!`);
            refreshProfile();
        } catch (e: any) {
            showToast(`Error: ${e.message}`);
        } finally {
            setLoading(null);
        }
    };

    const cosmetics = items.filter(i => i.type === 'cosmetic').sort((a, b) => a.cost - b.cost);
    const titles = items.filter(i => i.type === 'title').sort((a, b) => a.cost - b.cost);

    return (
        <div className="min-h-screen bg-[#0a0a12] text-gray-100 flex flex-col pt-4">
            {/* toast */}
            {toast && (
                <div className="fixed top-4 right-4 z-50 px-5 py-3 bg-gray-900 border border-gray-700 rounded-xl text-sm font-medium text-white shadow-2xl transition-all">
                    {toast}
                </div>
            )}

            <header className="max-w-5xl mx-auto w-full px-4 mb-8 flex items-center gap-4">
                <Link to="/casino" className="p-2 rounded-xl bg-gray-900 border border-gray-800 hover:bg-gray-800 transition-colors">
                    <ArrowLeft className="w-5 h-5 text-gray-400" />
                </Link>
                <div className="flex items-center gap-3">
                    <div className="p-2.5 bg-indigo-500/10 rounded-xl border border-indigo-500/20">
                        <ShoppingBag className="w-6 h-6 text-indigo-400" />
                    </div>
                    <div>
                        <h1 className="text-2xl font-bold text-white">Cosmetic Store</h1>
                        <p className="text-xs text-gray-500">spend your winnings on absolutely useless vanity items</p>
                    </div>
                </div>
                {profile && (
                    <div className="ml-auto text-sm text-gray-500 font-medium">
                        Balance: <span className="text-green-400 font-mono font-bold tracking-tight">{profile.Balance?.toLocaleString(undefined, { minimumFractionDigits: 2, maximumFractionDigits: 2 })}</span> S.H.I.T.
                    </div>
                )}
            </header>

            <main className="flex-1 max-w-5xl mx-auto px-4 w-full pb-12 space-y-10">
                {/* themes */}
                <section>
                    <div className="flex items-center gap-2 mb-4">
                        <Palette className="w-4 h-4 text-indigo-400" />
                        <h2 className="text-lg font-bold text-white">Site Themes</h2>
                        <span className="text-xs text-gray-500 ml-1">— makes the whole site look different</span>
                    </div>
                    <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
                        {cosmetics.map(item => {
                            const preview = THEME_PREVIEWS[item.id];
                            const isOwned = owned(item.id);
                            const equipped = isEquipped(item);
                            return (
                                <div key={item.id} className={`rounded-2xl border overflow-hidden transition-all ${equipped ? 'border-indigo-400/60 shadow-lg shadow-indigo-500/10' : 'border-gray-800 hover:border-gray-700'}`}>
                                    {/* preview swatch */}
                                    <div className={`h-20 bg-gradient-to-br ${preview?.gradient ?? 'from-gray-900 to-gray-800'} relative`}>
                                        <div className={`absolute bottom-2 right-2 px-2 py-1 rounded-lg text-xs font-bold border ${preview?.accent ?? 'border-gray-600 text-gray-400'}`}>
                                            preview
                                        </div>
                                        {equipped && (
                                            <div className="absolute top-2 left-2 flex items-center gap-1 px-2 py-1 bg-indigo-500/20 border border-indigo-400/40 rounded-lg text-xs text-indigo-300 font-semibold">
                                                <Check className="w-3 h-3" /> Active
                                            </div>
                                        )}
                                    </div>
                                    <div className="p-4 bg-gray-900">
                                        <div className="font-bold text-white mb-1">{item.name}</div>
                                        <div className="text-xs text-gray-500 mb-3">{preview?.desc}</div>
                                        <div className="flex items-center justify-between">
                                            <span className="text-yellow-400 font-mono font-bold text-sm">{item.cost.toLocaleString()} VEX</span>
                                            <button
                                                onClick={() => buy(item)}
                                                disabled={loading === item.id || (!isOwned && (profile?.Balance ?? 0) < item.cost)}
                                                className={`px-4 py-1.5 rounded-lg text-xs font-bold transition-all ${
                                                    equipped ? 'bg-indigo-600/40 text-indigo-300 cursor-default'
                                                    : isOwned ? 'bg-indigo-600 hover:bg-indigo-500 text-white'
                                                    : (profile?.Balance ?? 0) >= item.cost ? 'bg-yellow-600 hover:bg-yellow-500 text-black'
                                                    : 'bg-gray-700 text-gray-500 cursor-not-allowed'}`}
                                            >
                                                {loading === item.id ? '…' : equipped ? 'Equipped' : isOwned ? 'Equip' : `Buy (${item.cost.toLocaleString()})`}
                                            </button>
                                        </div>
                                    </div>
                                </div>
                            );
                        })}
                    </div>
                </section>

                {/* titles */}
                <section>
                    <div className="flex items-center gap-2 mb-4">
                        <Tag className="w-4 h-4 text-purple-400" />
                        <h2 className="text-lg font-bold text-white">Titles</h2>
                        <span className="text-xs text-gray-500 ml-1">— shown on the leaderboard next to your name</span>
                    </div>
                    <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                        {titles.map(item => {
                            const isOwned = owned(item.id);
                            const equipped = isEquipped(item);
                            return (
                                <div key={item.id} className={`p-5 rounded-2xl border flex items-center justify-between transition-all ${equipped ? 'border-purple-400/60 bg-purple-900/10' : 'border-gray-800 bg-gray-900/40 hover:border-gray-700'}`}>
                                    <div>
                                        <div className="flex items-center gap-2">
                                            {equipped && <Check className="w-3.5 h-3.5 text-purple-400" />}
                                            <span className="font-bold text-white">&ldquo;{item.value}&rdquo;</span>
                                        </div>
                                        <div className="text-yellow-400 font-mono text-sm mt-0.5">{item.cost.toLocaleString()} VEX</div>
                                    </div>
                                    <button
                                        onClick={() => buy(item)}
                                        disabled={loading === item.id || (!isOwned && (profile?.Balance ?? 0) < item.cost)}
                                        className={`px-4 py-2 rounded-xl text-sm font-bold transition-all ${
                                            equipped ? 'bg-purple-600/40 text-purple-300 cursor-default'
                                            : isOwned ? 'bg-purple-600 hover:bg-purple-500 text-white shadow-lg shadow-purple-500/10'
                                            : (profile?.Balance ?? 0) >= item.cost ? 'bg-yellow-600 hover:bg-yellow-500 text-black shadow-lg shadow-yellow-500/10'
                                            : 'bg-gray-800 text-gray-500 cursor-not-allowed'}`}
                                    >
                                        {loading === item.id ? '…' : equipped ? 'Equipped' : isOwned ? 'Equip' : 'Buy'}
                                    </button>
                                </div>
                            );
                        })}
                    </div>
                </section>
            </main>
        </div>
    );
}

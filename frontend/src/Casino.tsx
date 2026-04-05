import { Link } from 'react-router-dom';
import { ArrowLeft, Dices, Cherry, Spade, Droplets, ShoppingBag, Trophy } from 'lucide-react';
import type { User } from 'firebase/auth';
import { getRank } from './utils/ranks';

interface CasinoProps {
    user: User | null;
    profile: any;
}

const GAMES = [
    {
        to: '/casino/roulette',
        emoji: '🎡',
        icon: Dices,
        iconColor: 'text-red-500',
        iconBg: 'bg-red-500/10',
        iconBorder: 'border-red-500/20',
        title: 'Roulette',
        desc: 'give me ur money',
        btn: 'bg-red-600 hover:bg-red-500 shadow-red-500/20',
        hover: 'hover:border-red-500/30',
    },
    {
        to: '/casino/slots',
        emoji: '🎰',
        icon: Cherry,
        iconColor: 'text-yellow-500',
        iconBg: 'bg-yellow-500/10',
        iconBorder: 'border-yellow-500/20',
        title: 'VEX Slots',
        desc: 'match 🍒🍋🍊 or rare jason/charles jackpots.',
        btn: 'bg-yellow-600 hover:bg-yellow-500 shadow-yellow-500/20',
        hover: 'hover:border-yellow-500/30',
    },
    {
        to: '/casino/blackjack',
        emoji: '♠️',
        icon: Spade,
        iconColor: 'text-white',
        iconBg: 'bg-gray-700/50',
        iconBorder: 'border-gray-600/40',
        title: 'Blackjack',
        desc: '21 vs. the house',
        btn: 'bg-gray-700 hover:bg-gray-600 shadow-gray-700/20',
        hover: 'hover:border-gray-500/30',
    },
    {
        to: '/casino/lottery',
        emoji: '🌊',
        icon: Droplets,
        iconColor: 'text-cyan-400',
        iconBg: 'bg-cyan-500/10',
        iconBorder: 'border-cyan-500/20',
        title: 'Water Game Lottery',
        desc: '10 S.H.I.T. per ticket',
        btn: 'bg-cyan-700 hover:bg-cyan-600 shadow-cyan-500/20',
        hover: 'hover:border-cyan-500/30',
    },
    {
        to: '/casino/store',
        emoji: '🛒',
        icon: ShoppingBag,
        iconColor: 'text-indigo-400',
        iconBg: 'bg-indigo-500/10',
        iconBorder: 'border-indigo-500/20',
        title: 'Cosmetic Store',
        desc: 'Buy themes, titles, and other useless vanity items.',
        btn: 'bg-indigo-600 hover:bg-indigo-500 shadow-indigo-500/20',
        hover: 'hover:border-indigo-500/30',
    },
    {
        to: '/leaderboard',
        emoji: '🏆',
        icon: Trophy,
        iconColor: 'text-amber-400',
        iconBg: 'bg-amber-500/10',
        iconBorder: 'border-amber-500/20',
        title: 'Leaderboard',
        desc: 'biggest losers and also biggest losers',
        btn: 'bg-amber-600 hover:bg-amber-500 shadow-amber-500/20',
        hover: 'hover:border-amber-500/30',
    },
];

export default function Casino({ user, profile }: CasinoProps) {
    if (!user) {
        return (
            <div className="min-h-screen bg-[#111] text-gray-100 flex items-center justify-center flex-col">
                <h2 className="text-2xl font-bold mb-4">Please log in to enter the Casino</h2>
                <Link to="/" className="text-blue-500 hover:text-blue-400 transition-colors">Return Home</Link>
            </div>
        );
    }

    const rank = getRank(profile?.TotalLost ?? 0);

    return (
        <div className="min-h-screen bg-[#0a0a12] text-gray-100 flex flex-col pt-4">
            <header className="max-w-7xl mx-auto w-full px-4 sm:px-6 lg:px-8 mb-8 flex items-center gap-4">
                <Link to="/" className="p-2 rounded-xl bg-gray-900 border border-gray-800 hover:bg-gray-800 transition-colors">
                    <ArrowLeft className="w-5 h-5 text-gray-400" />
                </Link>
                <div className="flex items-center gap-3">
                    <div className="p-2.5 bg-purple-500/10 rounded-xl border border-purple-500/20">
                        <Dices className="w-6 h-6 text-purple-500" />
                    </div>
                    <h1 className="text-3xl font-bold bg-gradient-to-r from-purple-400 to-pink-500 bg-clip-text text-transparent">Casino</h1>
                </div>
            </header>

            <main className="flex-1 max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 w-full space-y-8 pb-12">
                {/* profile card */}
                {profile && (
                    <div className="flex flex-wrap gap-4">
                        <div className="p-6 bg-gray-900/50 backdrop-blur rounded-2xl border border-gray-800 shadow-xl">
                            <div className="text-xs font-semibold text-gray-500 mb-1 uppercase tracking-widest">Balance</div>
                            <div className="text-4xl font-mono text-green-400 font-bold tracking-tight">
                                {profile.Balance?.toLocaleString(undefined, { minimumFractionDigits: 2, maximumFractionDigits: 2 })}
                                <span className="text-sm text-green-500/50 ml-2 tracking-widest uppercase italic">S.H.I.T.</span>
                            </div>
                        </div>

                        <div className={`p-6 rounded-2xl border shadow-xl flex items-center gap-4 ${rank.bgClass} ${rank.borderClass}`}>
                            <div className="text-4xl">{rank.emoji}</div>
                            <div>
                                <div className="text-xs font-semibold text-gray-500 uppercase tracking-widest mb-0.5">Your Rank</div>
                                <div className={`text-2xl font-black ${rank.colorClass}`}>{rank.name}</div>
                                <div className="text-xs text-gray-500 mt-0.5">
                                    {(profile.TotalLost ?? 0) > 0
                                        ? `${(profile.TotalLost ?? 0).toLocaleString(undefined, { maximumFractionDigits: 0 })} S.H.I.T. lost`
                                        : 'lose some money to rank up'}
                                </div>
                            </div>
                        </div>

                        {profile.Title && (
                            <div className="p-6 rounded-2xl border border-purple-500/20 bg-purple-500/5 shadow-xl flex items-center">
                                <div>
                                    <div className="text-xs text-gray-500 uppercase tracking-widest mb-1">Title</div>
                                    <div className="text-lg font-bold text-purple-300">"{profile.Title}"</div>
                                </div>
                            </div>
                        )}
                    </div>
                )}

                {/* game grid */}
                <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
                    {GAMES.map(game => {
                        const Icon = game.icon;
                        return (
                            <div key={game.to} className={`bg-gray-900 border border-gray-800 rounded-3xl p-8 ${game.hover} transition-all group flex flex-col`}>
                                <div className={`w-16 h-16 ${game.iconBg} rounded-2xl flex items-center justify-center mb-6 border ${game.iconBorder} group-hover:scale-110 transition-transform`}>
                                    <Icon className={`w-8 h-8 ${game.iconColor}`} />
                                </div>
                                <h2 className="text-2xl font-bold text-white mb-3">{game.title}</h2>
                                <p className="text-gray-400 mb-8 flex-1 text-sm leading-relaxed">{game.desc}</p>
                                <Link
                                    to={game.to}
                                    className={`block text-center w-full py-4 ${game.btn} text-white font-bold rounded-xl transition-colors shadow-lg`}
                                >
                                    {game.title === 'Leaderboard' || game.title === 'Cosmetic Store' ? `Open ${game.title}` : `Play ${game.title}`}
                                </Link>
                            </div>
                        );
                    })}
                </div>
            </main>
        </div>
    );
}

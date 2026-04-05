export interface Rank {
    name: string;
    emoji: string;
    colorClass: string;
    bgClass: string;
    borderClass: string;
    threshold: number;
}

export const RANKS: Rank[] = [
    { name: 'Unranked',  emoji: '❓', colorClass: 'text-gray-500',   bgClass: 'bg-gray-500/10',   borderClass: 'border-gray-500/20',   threshold: 0 },
    { name: 'Gooner',    emoji: '🎮', colorClass: 'text-green-400',  bgClass: 'bg-green-500/10',  borderClass: 'border-green-500/20',  threshold: 1_000 },
    { name: 'Charles',   emoji: '👔', colorClass: 'text-blue-400',   bgClass: 'bg-blue-500/10',   borderClass: 'border-blue-500/20',   threshold: 10_000 },
    { name: 'Andrew',    emoji: '🤵', colorClass: 'text-purple-400', bgClass: 'bg-purple-500/10', borderClass: 'border-purple-500/20', threshold: 100_000 },
    { name: 'Jason',     emoji: '⚡', colorClass: 'text-yellow-400', bgClass: 'bg-yellow-500/10', borderClass: 'border-yellow-500/20', threshold: 1_000_000 },
    { name: 'Femboy',    emoji: '🎀', colorClass: 'text-pink-400',   bgClass: 'bg-pink-500/10',   borderClass: 'border-pink-500/20',   threshold: 10_000_000 },
    { name: 'Jason++',   emoji: '💀', colorClass: 'text-red-400',    bgClass: 'bg-red-500/10',    borderClass: 'border-red-500/20',    threshold: 67_000_000 },
];

export function getRank(totalLost: number): Rank {
    for (let i = RANKS.length - 1; i >= 0; i--) {
        if (totalLost >= RANKS[i].threshold) return RANKS[i];
    }
    return RANKS[0];
}

export function formatNumber(n: number): string {
    if (n >= 1_000_000) return (n / 1_000_000).toFixed(1) + 'M';
    if (n >= 1_000) return (n / 1_000).toFixed(1) + 'K';
    return n.toFixed(2);
}

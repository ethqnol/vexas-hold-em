import { useState, useEffect } from 'react';
import { Activity, Server, Database, CheckCircle2, XCircle, Clock } from 'lucide-react';
import { clsx } from 'clsx';
import { twMerge } from 'tailwind-merge';

function cn(...inputs: (string | undefined | null | false)[]) {
    return twMerge(clsx(inputs));
}

function Health() {
    const [status, setStatus] = useState<'checking' | 'healthy' | 'unhealthy'>('checking');
    const [latency, setLatency] = useState<number | null>(null);
    const [lastCheck, setLastCheck] = useState<Date | null>(null);

    const checkHealth = async () => {
        setStatus('checking');
        const start = performance.now();
        try {
            const response = await fetch('http://localhost:8080/health');
            if (response.ok) {
                setStatus('healthy');
                setLatency(Math.round(performance.now() - start));
            } else {
                setStatus('unhealthy');
                setLatency(null);
            }
        } catch {
            setStatus('unhealthy');
            setLatency(null);
        }
        setLastCheck(new Date());
    };

    useEffect(() => {
        checkHealth();
        const interval = setInterval(checkHealth, 30000); // check every 30s
        return () => clearInterval(interval);
    }, []);

    return (
        <div className="min-h-screen bg-[#0a0a0a] text-gray-100 flex flex-col pt-20 px-4 sm:px-6 lg:px-8">
            <div className="max-w-3xl mx-auto w-full space-y-8">
                {/* header */}
                <div className="flex flex-col items-center text-center space-y-4">
                    <div className="p-3 bg-gray-900 rounded-2xl border border-gray-800 shadow-xl">
                        <Activity className="w-8 h-8 text-blue-500" />
                    </div>
                    <h1 className="text-4xl font-bold tracking-tight">System Status</h1>
                    <p className="text-gray-400 max-w-lg">
                        Real-time latency and uptime monitoring for VEXAS Hold'em infrastructure.
                    </p>
                </div>

                {/* global status banner */}
                <div
                    className={cn(
                        "p-6 rounded-2xl border flex items-center justify-between transition-all duration-500",
                        status === 'checking' && "bg-gray-900/50 border-gray-800",
                        status === 'healthy' && "bg-emerald-950/30 border-emerald-900/50 shadow-[0_0_30px_rgba(16,185,129,0.1)]",
                        status === 'unhealthy' && "bg-red-950/30 border-red-900/50 shadow-[0_0_30px_rgba(239,68,68,0.1)]"
                    )}
                >
                    <div className="flex items-center gap-4">
                        <div className="relative">
                            {status === 'checking' && <div className="w-3 h-3 rounded-full bg-blue-500 animate-ping" />}
                            {status === 'healthy' && (
                                <>
                                    <div className="w-3 h-3 rounded-full bg-emerald-500 absolute animate-ping opacity-75" />
                                    <div className="w-3 h-3 rounded-full bg-emerald-500 relative" />
                                </>
                            )}
                            {status === 'unhealthy' && <div className="w-3 h-3 rounded-full bg-red-500" />}
                        </div>
                        <span className="text-lg font-medium">
                            {status === 'checking' && 'Checking systems...'}
                            {status === 'healthy' && 'All Systems Operational'}
                            {status === 'unhealthy' && 'Systems Outage'}
                        </span>
                    </div>

                    <button
                        onClick={checkHealth}
                        disabled={status === 'checking'}
                        className="text-sm px-4 py-2 rounded-lg bg-gray-800 hover:bg-gray-700 transition-colors border border-gray-700 disabled:opacity-50"
                    >
                        Refresh
                    </button>
                </div>

                {/* comp grid */}
                <div className="grid gap-4 sm:grid-cols-2">
                    {/* api server card */}
                    <div className="p-6 rounded-2xl bg-gray-900 border border-gray-800 flex flex-col justify-between group hover:border-gray-700 transition-colors">
                        <div className="flex justify-between items-start mb-4">
                            <div className="flex items-center gap-3">
                                <div className="p-2 bg-gray-800 rounded-lg group-hover:bg-gray-700 transition-colors">
                                    <Server className="w-5 h-5 text-gray-300" />
                                </div>
                                <div>
                                    <h3 className="font-semibold">API Server</h3>
                                    <p className="text-sm text-gray-500">us-east1 (Go/Fiber)</p>
                                </div>
                            </div>
                            {status === 'healthy' ? (
                                <CheckCircle2 className="w-5 h-5 text-emerald-500" />
                            ) : status === 'unhealthy' ? (
                                <XCircle className="w-5 h-5 text-red-500" />
                            ) : (
                                <Activity className="w-5 h-5 text-gray-600 animate-pulse" />
                            )}
                        </div>

                        <div className="flex items-end justify-between mt-4">
                            <div className="space-y-1">
                                <p className="text-sm text-gray-500 flex items-center gap-1.5">
                                    <Clock className="w-3.5 h-3.5" />
                                    Latency
                                </p>
                                <div className="flex items-baseline gap-1">
                                    <span className="text-2xl font-semibold">
                                        {latency !== null ? latency : '--'}
                                    </span>
                                    <span className="text-sm text-gray-500 tracking-wide">ms</span>
                                </div>
                            </div>

                            {/* fake latency sparkline */}
                            <div className="flex items-end gap-1 h-10">
                                {[40, 60, 45, 80, 55, 45, latency ?? 40].map((h, i) => (
                                    <div
                                        key={i}
                                        className={cn(
                                            "w-2 rounded-t-sm transition-all duration-500",
                                            status === 'healthy' ? "bg-emerald-500/20" : "bg-gray-800",
                                            i === 6 && status === 'healthy' && "bg-emerald-500/50"
                                        )}
                                        style={{ height: `${Math.min(100, (h / 100) * 100)}%` }}
                                    />
                                ))}
                            </div>
                        </div>
                    </div>

                    {/* db card */}
                    <div className="p-6 rounded-2xl bg-gray-900 border border-gray-800 flex flex-col justify-between group hover:border-gray-700 transition-colors">
                        <div className="flex justify-between items-start mb-4">
                            <div className="flex items-center gap-3">
                                <div className="p-2 bg-gray-800 rounded-lg group-hover:bg-gray-700 transition-colors">
                                    <Database className="w-5 h-5 text-gray-300" />
                                </div>
                                <div>
                                    <h3 className="font-semibold">Firestore Gen2</h3>
                                    <p className="text-sm text-gray-500">Database Cluster</p>
                                </div>
                            </div>
                            {status === 'healthy' ? (
                                <CheckCircle2 className="w-5 h-5 text-emerald-500" />
                            ) : status === 'unhealthy' ? (
                                <XCircle className="w-5 h-5 text-red-500" />
                            ) : (
                                <Activity className="w-5 h-5 text-gray-600 animate-pulse" />
                            )}
                        </div>

                        <div className="flex items-end justify-between mt-4">
                            <div className="space-y-1">
                                <p className="text-sm text-gray-500 flex items-center gap-1.5">
                                    <Clock className="w-3.5 h-3.5" />
                                    Connection
                                </p>
                                <div className="flex items-baseline gap-1">
                                    <span className="text-2xl font-semibold">
                                        {status === 'healthy' ? 'Linked' : '--'}
                                    </span>
                                </div>
                            </div>
                            {/* fixed height fake sparkline */}
                            <div className="flex items-end gap-1 h-10">
                                {[100, 100, 100, 100, 100, 100, 100].map((h, i) => (
                                    <div
                                        key={i}
                                        className={cn(
                                            "w-2 rounded-t-sm transition-all duration-500",
                                            status === 'healthy' ? "bg-emerald-500/20" : "bg-gray-800",
                                            i === 6 && status === 'healthy' && "bg-emerald-500/50"
                                        )}
                                        style={{ height: `${h}%` }}
                                    />
                                ))}
                            </div>
                        </div>
                    </div>
                </div>

                {/* footer */}
                <div className="text-center text-sm text-gray-600">
                    Last updated: {lastCheck ? lastCheck.toLocaleTimeString() : 'Never'}
                </div>
            </div>
        </div>
    );
}

export default Health;

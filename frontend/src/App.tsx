import { useState, useEffect } from 'react';
import { Routes, Route, Link } from 'react-router-dom';
import { signInWithPopup, GoogleAuthProvider, signOut, onAuthStateChanged } from 'firebase/auth';
import type { User } from 'firebase/auth';
import { auth } from './firebase';
import { LogIn, LogOut, Loader2, ShieldAlert } from 'lucide-react';
import Health from './Health';
import Admin from './Admin';
import Competition from './Competition';

function App() {
  const [user, setUser] = useState<User | null>(null);
  const [loading, setLoading] = useState(true);
  const [profile, setProfile] = useState<any>(null);
  const [portfolio, setPortfolio] = useState<any>(null);
  const [competitions, setCompetitions] = useState<any[]>([]);

  const fetchUserProfile = async (uid: string) => {
    try {
      const res = await fetch(`http://localhost:8080/api/v1/users/${uid}`);
      if (res.ok) {
        const data = await res.json();
        setProfile(data.data);
      }
    } catch (e) {
      console.error(e);
    }
  };

  const fetchUserPortfolio = async (uid: string) => {
    try {
      const res = await fetch(`http://localhost:8080/api/v1/users/${uid}/portfolio`);
      if (res.ok) {
        const data = await res.json();
        setPortfolio(data.data);
      }
    } catch (e) {
      console.error(e);
    }
  };

  useEffect(() => {
    // Fetch all competitions on load
    const fetchCompetitions = async () => {
      try {
        const res = await fetch('http://localhost:8080/api/v1/competitions');
        if (res.ok) {
          const data = await res.json();
          setCompetitions(data.competitions || []);
        }
      } catch (e) {
        console.error(e);
      }
    };

    fetchCompetitions();

    const unsubscribe = onAuthStateChanged(auth, async (currentUser) => {
      setUser(currentUser);
      setLoading(false);

      if (currentUser) {
        // Automatically sync/create user in Firestore DB
        try {
          await fetch(`http://localhost:8080/api/v1/users/${currentUser.uid}/sync`, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({
              email: currentUser.email,
              displayName: currentUser.displayName
            })
          });
        } catch (e) {
          console.error("Failed to sync user", e);
        }

        fetchUserProfile(currentUser.uid);
        fetchUserPortfolio(currentUser.uid);
      } else {
        setProfile(null);
        setPortfolio(null);
      }
    });
    return () => unsubscribe();
  }, []);

  const isAdmin = user?.email === 'drinkfood.exe@gmail.com';

  const handleLogin = async () => {
    const provider = new GoogleAuthProvider();
    try {
      await signInWithPopup(auth, provider);
    } catch (error) {
      console.error("Error signing in with Google", error);
    }
  };

  const handleLogout = async () => {
    try {
      await signOut(auth);
    } catch (error) {
      console.error("Error signing out", error);
    }
  };

  if (loading) {
    return (
      <div className="flex items-center justify-center min-h-screen bg-gray-950">
        <Loader2 className="w-8 h-8 text-blue-500 animate-spin" />
      </div>
    );
  }

  return (
    <Routes>
      <Route path="/" element={
        <div className="min-h-screen bg-gray-950 text-gray-100 flex flex-col">
          <header className="border-b border-gray-800 bg-gray-900/50 backdrop-blur-sm sticky top-0 z-50">
            <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 h-16 flex items-center justify-between">
              <div className="flex items-center gap-2">
                <span className="text-xl font-bold bg-gradient-to-r from-blue-400 to-indigo-500 bg-clip-text text-transparent">
                  VEXAS Hold'em
                </span>
              </div>

              <div>
                {user ? (
                  <div className="flex items-center gap-4">
                    <div className="hidden sm:block text-sm text-gray-400 text-right">
                      <div>{user.displayName}</div>
                      {profile && <div className="text-green-400 font-mono text-xs">{profile.Balance?.toFixed(2) || "0.00"} S.H.I.T.</div>}
                    </div>
                    {isAdmin && (
                      <Link
                        to="/admin"
                        className="flex items-center gap-2 px-4 py-2 text-sm font-medium text-indigo-400 bg-indigo-500/10 hover:bg-indigo-500/20 rounded-lg transition-colors border border-indigo-500/20"
                      >
                        <ShieldAlert className="w-4 h-4" />
                        Admin Panel
                      </Link>
                    )}
                    <button
                      onClick={handleLogout}
                      className="flex items-center gap-2 px-4 py-2 text-sm font-medium text-gray-300 bg-gray-800 hover:bg-gray-700 rounded-lg transition-colors border border-gray-700 hover:border-gray-600"
                    >
                      <LogOut className="w-4 h-4" />
                      Sign Out
                    </button>
                  </div>
                ) : (
                  <button
                    onClick={handleLogin}
                    className="flex items-center gap-2 px-4 py-2 text-sm font-medium text-white bg-blue-600 hover:bg-blue-500 rounded-lg transition-colors shadow-lg shadow-blue-500/20"
                  >
                    <LogIn className="w-4 h-4" />
                    Sign In with Google
                  </button>
                )}
              </div>
            </div>
          </header>

          <main className="flex-1 max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-12 w-full">
            {!user ? (
              <div className="flex flex-col items-center justify-center h-[60vh] text-center">
                <h1 className="text-4xl sm:text-6xl font-extrabold tracking-tight mb-6">
                  Trade <span className="text-blue-500">VEX Robotics</span>
                  <br />Predictions
                </h1>
                <p className="text-xl text-gray-400 max-w-2xl mb-10">
                  The first AMM-based prediction market for the VEX Robotics Competition.
                  Buy YES or NO shares on your favorite teams and win big.
                </p>
                <button
                  onClick={handleLogin}
                  className="flex items-center gap-3 px-8 py-4 text-lg font-semibold text-white bg-blue-600 hover:bg-blue-500 rounded-xl transition-all shadow-xl shadow-blue-500/20 hover:scale-105 active:scale-95"
                >
                  <LogIn className="w-6 h-6" />
                  Start Trading Now
                </button>
              </div>
            ) : (
              <div className="grid grid-cols-1 lg:grid-cols-3 gap-8">
                <div className="lg:col-span-2 space-y-6">
                  <div className="p-8 rounded-2xl bg-gray-900 border border-gray-800 shadow-xl">
                    <h2 className="text-2xl font-bold mb-6 flex items-center justify-between">
                      <span>Active Competitions</span>
                    </h2>

                    {competitions.length === 0 ? (
                      <div className="p-8 border border-dashed border-gray-800 rounded-xl text-center text-gray-500">
                        <Loader2 className="w-8 h-8 text-indigo-500/50 animate-spin mx-auto mb-4" />
                        Fetching active events...
                      </div>
                    ) : (
                      <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                        {competitions.map((compWrapper, idx) => {
                          const compId = compWrapper.id;
                          const status = compWrapper.data?.status || 'active';
                          return (
                            <Link key={compId || idx} to={`/competition/${compId}`} className="block">
                              <div className="p-6 rounded-2xl bg-gray-950 border border-gray-800 hover:border-indigo-500/50 hover:bg-gray-800/80 transition-all shadow-sm hover:shadow-indigo-500/10 group cursor-pointer h-full flex flex-col">
                                <div className="flex justify-between items-start mb-4">
                                  <h3 className="font-bold text-xl text-white group-hover:text-indigo-400 transition-colors">{compId}</h3>
                                  <div className={`text-[10px] uppercase tracking-widest px-2 py-1 rounded w-max ml-auto shadow-inner border ${status === 'active' ? 'bg-green-500/10 text-green-400 border-green-500/20' : 'bg-gray-800 text-gray-400 border-gray-700'}`}>
                                    {status}
                                  </div>
                                </div>
                                <p className="text-sm text-gray-400 mt-auto">Click to view prediction markets and begin trading shares.</p>
                              </div>
                            </Link>
                          );
                        })}
                      </div>
                    )}
                  </div>
                </div>

                <div className="space-y-6">
                  {/* Portfolio will go here */}
                  <div className="p-8 rounded-2xl bg-gray-900 border border-gray-800 shadow-xl">
                    <div className="flex justify-between items-center mb-6">
                      <h2 className="text-xl font-bold flex items-center gap-2 text-white">
                        Your Portfolio
                      </h2>
                    </div>

                    {profile && (
                      <div className="mb-6 p-6 bg-gray-950 rounded-xl border border-gray-800 shadow-inner">
                        <div className="flex justify-between items-start mb-2">
                          <div>
                            <div className="text-sm font-semibold text-gray-400">Available Balance</div>
                            <div className="text-[10px] text-gray-500 uppercase tracking-widest mt-1">
                              Secure High-yield Investment Token
                            </div>
                          </div>
                          <div className="bg-green-500/10 px-2 py-1 rounded text-[10px] font-bold text-green-400 border border-green-500/20">
                            S.H.I.T.
                          </div>
                        </div>
                        <div className="text-4xl font-mono text-green-400 mt-2 font-bold tracking-tight drop-shadow-sm">
                          {profile.Balance?.toFixed(2) || "0.00"}
                        </div>
                      </div>
                    )}

                    {portfolio && Object.keys(portfolio).length > 0 ? (
                      <div className="space-y-4">
                        <h3 className="text-sm text-gray-400 font-medium uppercase tracking-wider mb-2">Active Positions</h3>
                        {Object.entries(portfolio).map(([teamId, pos]: [string, any]) => {
                          const hasYes = pos.YesShares > 0;
                          const hasNo = pos.NoShares > 0;

                          if (!hasYes && !hasNo) return null;

                          return (
                            <div key={teamId} className="flex flex-col gap-3 p-5 bg-gray-950/50 hover:bg-gray-800 transition-all rounded-xl border border-gray-800 hover:border-gray-600 shadow-sm">
                              <div className="flex justify-between items-center border-b border-gray-800/50 pb-2 mb-1">
                                <span className="font-bold text-gray-100 text-lg tracking-wide">Team {teamId}</span>
                              </div>
                              <div className="flex gap-3">
                                {hasYes && (
                                  <div className="flex-1 flex flex-col items-center justify-center p-2 bg-green-500/10 text-green-400 rounded-lg border border-green-500/20">
                                    <span className="font-bold text-[10px] uppercase tracking-wider opacity-80 mb-1">YES Shares</span>
                                    <span className="font-mono text-xl font-bold">{pos.YesShares?.toFixed(1)}</span>
                                  </div>
                                )}
                                {hasNo && (
                                  <div className="flex-1 flex flex-col items-center justify-center p-2 bg-red-500/10 text-red-400 rounded-lg border border-red-500/20">
                                    <span className="font-bold text-[10px] uppercase tracking-wider opacity-80 mb-1">NO Shares</span>
                                    <span className="font-mono text-xl font-bold">{pos.NoShares?.toFixed(1)}</span>
                                  </div>
                                )}
                              </div>
                            </div>
                          );
                        })}
                      </div>
                    ) : (
                      <div className="text-center p-6 border border-dashed border-gray-700 rounded-xl">
                        <p className="text-gray-400">No active positions yet.</p>
                        <p className="text-sm text-gray-500 mt-2">Trades you make will appear here.</p>
                      </div>
                    )}
                  </div>
                </div>
              </div>
            )}
          </main>
        </div>
      } />
      <Route path="/competition/:id" element={<Competition />} />
      <Route path="/health" element={<Health />} />
      <Route path="/admin" element={<Admin user={user} />} />
    </Routes>
  );
}

export default App;

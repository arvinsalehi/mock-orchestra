import React, { createContext, useContext, useEffect, useState, ReactNode } from 'react';
import { useQueryClient, useQuery } from '@tanstack/react-query';
import { TestResult } from '../types';

interface SessionState {
  sessionHash: string | null;
  status: 'active' | 'pending' | 'idle';
  operator: string;
  buildNumber: string;
  tests: TestResult[];
  login: (hash: string, status: string, op: string, build: string) => void;
  logout: () => void;
  setStatus: (status: 'active' | 'pending') => void;
}

const SessionContext = createContext<SessionState | undefined>(undefined);

export const SessionProvider = ({ children }: { children: ReactNode }) => {
  const [sessionHash, setSessionHash] = useState<string | null>(null);
  const [status, setStatus] = useState<'active' | 'pending' | 'idle'>('idle');
  const [operator, setOperator] = useState('');
  const [buildNumber, setBuildNumber] = useState('');
  
  const queryClient = useQueryClient();

  // 1. TanStack Query for Tests Data
  const { data: tests = [] } = useQuery({
    queryKey: ['tests', sessionHash],
    queryFn: () => [], // Ideally fetch existing results from API on reconnect
    staleTime: Infinity,
    enabled: !!sessionHash,
  });

  // 2. WebSocket Logic (Centralized)
  useEffect(() => {
    if (!sessionHash || status !== 'active') return;

    // Direct connection to Backend (127.0.0.1:8000)
    const ws = new WebSocket(`ws://127.0.0.1:8000/ws/${sessionHash}`);

    ws.onmessage = (event) => {
      try {
        const newResult: TestResult = JSON.parse(event.data);
        console.log("🔍 WS Received Object:", newResult); 
        // Direct Cache Update
        queryClient.setQueryData<TestResult[]>(['tests', sessionHash], (old) => {
          const list = old ? [...old] : [];
          const index = list.findIndex(t => t.test_name === newResult.test_name);
          if (index > -1) list[index] = newResult;
          else list.push(newResult);
          return list;
        });
      } catch (e) { console.error("WS Parse Error", e); }
    };

    return () => { if (ws.readyState === 1) ws.close(); };
  }, [sessionHash, status, queryClient]);

  // Actions
  const login = (hash: string, st: string, op: string, build: string) => {
    setSessionHash(hash);
    setStatus(st as any);
    setOperator(op);
    setBuildNumber(build);
  };

  const logout = () => {
    setSessionHash(null);
    setStatus('idle');
    setOperator('');
    setBuildNumber('');
    queryClient.removeQueries({ queryKey: ['tests'] });
  };

  return (
    <SessionContext.Provider value={{ 
      sessionHash, status, operator, buildNumber, tests, 
      login, logout, setStatus 
    }}>
      {children}
    </SessionContext.Provider>
  );
};

export const useSession = () => {
  const context = useContext(SessionContext);
  if (!context) throw new Error('useSession must be used within SessionProvider');
  return context;
};


import React from 'react';
import { useMutation } from '@tanstack/react-query';
import { useSession } from '../context/SessionContext';
import { TestTable } from './TestTable';
import { pauseSession, resumeSession, finishSession } from '../api';

export const SessionDashboard = () => {
  const { sessionHash, status, operator, buildNumber, setStatus, logout } = useSession();

  // Mutations (Logic moved close to buttons)
  const pauseMut = useMutation({
    mutationFn: () => pauseSession(sessionHash!),
    onSuccess: () => setStatus('pending')
  });

  const resumeMut = useMutation({
    mutationFn: () => resumeSession(sessionHash!),
    onSuccess: () => setStatus('active')
  });

  const finishMut = useMutation({
    mutationFn: () => finishSession(sessionHash!),
    onSuccess: logout
  });

  return (
    <div className="max-w-4xl mx-auto mt-10 p-6">
      <header className="flex justify-between items-center mb-6 border-b pb-4">
        <div>
          <h1 className="text-2xl font-bold text-gray-800">Test Session</h1>
          <div className="text-sm text-gray-500 mt-1">
            <span className="mr-4">OP: <strong>{operator}</strong></span>
            <span>BUILD: <strong>{buildNumber}</strong></span>
          </div>
        </div>
        
        <div className="space-x-3">
          {status === 'active' ? (
            <button 
              onClick={() => pauseMut.mutate()} 
              className="px-4 py-2 bg-yellow-500 text-white rounded hover:bg-yellow-600 transition"
            >
              Pause
            </button>
          ) : (
            <button 
              onClick={() => resumeMut.mutate()} 
              className="px-4 py-2 bg-green-600 text-white rounded hover:bg-green-700 transition"
            >
              Resume
            </button>
          )}
          <button 
            onClick={() => finishMut.mutate()} 
            className="px-4 py-2 bg-red-600 text-white rounded hover:bg-red-700 transition"
          >
            Finish
          </button>
        </div>
      </header>

      {/* The Table component will get data from context itself */}
      <TestTable />
    </div>
  );
};


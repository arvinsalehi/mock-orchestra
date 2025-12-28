import React, { useState } from 'react';
import { useMutation } from '@tanstack/react-query';
import { useSession } from '../context/SessionContext';
import { startTest } from '../api';

export const LoginForm = () => {
  const { login } = useSession();
  const [operator, setOp] = useState('');
  const [build, setBuild] = useState('');

  const mutation = useMutation({
    mutationFn: () => startTest(operator, build),
    onSuccess: (data) => login(data.session_hash, data.status, operator, build),
    onError: (e) => alert(e.message)
  });

  return (
    <div className="flex items-center justify-center min-h-screen bg-gray-100">
      <div className="w-full max-w-md p-8 space-y-6 bg-white rounded shadow-md">
        <h2 className="text-2xl font-bold text-center text-gray-800">Device Test Hub</h2>
        <div className="space-y-4">
          <div>
            <label className="block text-sm font-medium text-gray-700">Operator</label>
            <input value={operator} onChange={e => setOp(e.target.value)} className="w-full p-2 mt-1 border rounded focus:ring-2 focus:ring-blue-500" />
          </div>
          <div>
            <label className="block text-sm font-medium text-gray-700">Build Number</label>
            <input value={build} onChange={e => setBuild(e.target.value)} className="w-full p-2 mt-1 border rounded focus:ring-2 focus:ring-blue-500" />
          </div>
          <button 
            onClick={() => mutation.mutate()} 
            disabled={mutation.isPending}
            className="w-full py-2 text-white bg-blue-600 rounded hover:bg-blue-700 disabled:opacity-50"
          >
            {mutation.isPending ? 'Starting...' : 'Start Session'}
          </button>
        </div>
      </div>
    </div>
  );
};


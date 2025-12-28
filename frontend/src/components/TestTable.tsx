import React from 'react';
import { useSession } from '../context/SessionContext';

export const TestTable = () => {
  const { tests, status } = useSession();

  if (status === 'pending') {
    return (
      <div className="p-6 bg-yellow-50 border-l-4 border-yellow-400 rounded-r">
        <h3 className="font-bold text-yellow-800">Session Paused</h3>
        <p className="text-yellow-700">Real-time updates are suspended.</p>
        <div className="opacity-40 mt-4 pointer-events-none grayscale">
           <TableMarkup tests={tests} />
        </div>
      </div>
    );
  }

  return <TableMarkup tests={tests} />;
};

// Reusable Markup
const TableMarkup = ({ tests }: { tests: any[] }) => (
  <div className="overflow-hidden border rounded-lg shadow-sm">
    <table className="min-w-full bg-white">
      <thead className="bg-gray-50">
        <tr>
          <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Test Name</th>
          <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Status</th>
          <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Timestamp</th>
        </tr>
      </thead>
      <tbody className="divide-y divide-gray-200">
        {tests.map((t) => (
          <tr key={t.test_name} className="transition-colors duration-150">
            <td className="px-6 py-4 whitespace-nowrap font-medium text-gray-900">{t.test_name}</td>
            <td className="px-6 py-4 whitespace-nowrap">
              <span className={`px-2 inline-flex text-xs leading-5 font-semibold rounded-full 
                ${t.status === 'pass' ? 'bg-green-100 text-green-800' : 
                  t.status === 'fail' ? 'bg-red-100 text-red-800' : 'bg-gray-100 text-gray-800'}`}>
                {t.status.toUpperCase()}
              </span>
            </td>
            <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
              {new Date(t.timestamp).toLocaleTimeString()}
            </td>
          </tr>
        ))}
        {tests.length === 0 && (
          <tr><td colSpan={3} className="px-6 py-10 text-center text-gray-400 italic">Waiting for incoming test data...</td></tr>
        )}
      </tbody>
    </table>
  </div>
);

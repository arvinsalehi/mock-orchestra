import React from 'react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { SessionProvider, useSession } from './context/SessionContext';
import { LoginForm } from './components/LoginForm';
import { SessionDashboard } from './components/SessionDashboard';

const queryClient = new QueryClient();

const MainContent = () => {
  const { status } = useSession();
  // Render based on state
  // return status === 'idle' ? <LoginForm /> : <SessionDashboard />; // We can auto navigate to the active test session
  return <LoginForm />;
};

export default function App() {
  return (
    <QueryClientProvider client={queryClient}>
      <SessionProvider>
        <MainContent />
      </SessionProvider>
    </QueryClientProvider>
  );
}

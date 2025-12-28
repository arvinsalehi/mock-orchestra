import React from 'react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { SessionProvider, useSession } from './context/SessionContext';
import { LoginForm } from './components/LoginForm';
import { SessionDashboard } from './components/SessionDashboard';

const queryClient = new QueryClient();

const MainContent = () => {
  const { status } = useSession();
  // Render based on state
  // There is no status idle in the context, but we use it here to indicate no session
  return status === 'idle' ? <LoginForm /> : <SessionDashboard />;
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

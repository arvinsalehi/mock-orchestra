const API_URL = "/api";  // Use relative path

export const startTest = async (operator: string, build_number: string) => {
  const res = await fetch(`${API_URL}/start-test`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ operator, build_number })
  });
  if (!res.ok) throw new Error(await res.text());
  return res.json();
};

export const pauseSession = (hash: string) => 
  fetch(`${API_URL}/pause/${hash}`, { method: 'PATCH' });

export const resumeSession = (hash: string) => 
  fetch(`${API_URL}/resume/${hash}`, { method: 'PATCH' });

export const finishSession = (hash: string) => 
  fetch(`${API_URL}/finish/${hash}`, { method: 'PATCH' });


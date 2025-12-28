export interface TestResult {
  test_name: string;
  status: 'pass' | 'fail' | 'pending';
  timestamp: string;
  metrics?: Record<string, any>;
}

export interface SessionData {
  session_hash: string;
  status: 'active' | 'pending' | 'completed';
  tests: TestResult[];
}


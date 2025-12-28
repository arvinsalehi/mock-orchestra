from pydantic import BaseModel
from typing import List, Optional
from datetime import datetime

class TestResult(BaseModel):
    test_name: str
    status: str  # "pass", "fail"
    metrics: Optional[dict] = {}
    timestamp: datetime = datetime.utcnow()

class SessionStartRequest(BaseModel):
    operator: str
    build_number: str

class SessionState(BaseModel):
    session_hash: str
    build_number: str
    operator: str
    status: str  # "active", "pending"
    start_time: str


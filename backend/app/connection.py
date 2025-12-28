from fastapi import WebSocket
from typing import Dict, List

class ConnectionManager:
    def __init__(self):
        # Map session_hash -> List[WebSocket]
        self.active_connections: Dict[str, List[WebSocket]] = {}

    async def connect(self, session_hash: str, websocket: WebSocket):
        await websocket.accept()
        if session_hash not in self.active_connections:
            self.active_connections[session_hash] = []
        self.active_connections[session_hash].append(websocket)

    def disconnect(self, session_hash: str, websocket: WebSocket):
        if session_hash in self.active_connections:
            if websocket in self.active_connections[session_hash]:
                self.active_connections[session_hash].remove(websocket)
            if not self.active_connections[session_hash]:
                del self.active_connections[session_hash]

    async def disconnect_all(self, session_hash: str):
        """Force close all connections for a paused/finished session."""
        if session_hash in self.active_connections:
            for ws in self.active_connections[session_hash]:
                try:
                    await ws.close(code=1000, reason="Session Paused/Finished")
                except Exception:
                    pass
            del self.active_connections[session_hash]

    async def broadcast(self, session_hash: str, message: str ):
        if session_hash in self.active_connections:
            for connection in self.active_connections[session_hash]:
                await connection.send_text(message)

manager = ConnectionManager()

